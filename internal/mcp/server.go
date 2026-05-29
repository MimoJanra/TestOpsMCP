package mcp

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/core"
	sessctx "github.com/MimoJanra/TestOpsMCP/internal/session"
	"github.com/MimoJanra/TestOpsMCP/internal/tools"
)

const (
	sessionSendBuffer = 64           // increased from 16 to reduce response drops under burst load
	heartbeatInterval = 25 * time.Second
	maxMessageBody    = 1 << 20 // 1 MiB
)

// Version is set at build time via -ldflags "-X github.com/MimoJanra/TestOpsMCP/internal/mcp.Version=x.y.z"
var Version = "dev"

type Options struct {
	AuthToken       string
	CORSAllowOrigin string
}

type Server struct {
	registry *tools.Registry
	logger   *core.Logger
	opts     Options

	mu       sync.RWMutex
	sessions map[string]*session
}

type session struct {
	id     string
	send   chan []byte
	ctx    context.Context
	cancel context.CancelFunc
}

func NewServer(registry *tools.Registry, logger *core.Logger, opts Options) *Server {
	return &Server{
		registry: registry,
		logger:   logger,
		opts:     opts,
		sessions: make(map[string]*session),
	}
}

// HandleSSE serves the MCP SSE transport: streams the per-session endpoint URL
// and subsequent JSON-RPC responses to the client.
func (s *Server) HandleSSE(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.checkAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.logger.Error("streaming not supported by ResponseWriter", nil, nil)
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	sess := s.newSession(r.Context())
	defer s.closeSession(sess)

	// If the client supplies their Allure token as a header, store it for this session
	// so all tool calls use it automatically — no need to call configure_allure_token.
	if allureToken := r.Header.Get("X-Allure-Token"); allureToken != "" {
		s.registry.SetSessionToken(sess.id, allureToken)
	}

	s.logger.Info("SSE client connected", map[string]any{"session": sess.id})

	if _, err := fmt.Fprintf(w, "event: endpoint\ndata: /messages?sessionId=%s\n\n", sess.id); err != nil {
		s.logger.Error("write endpoint event", err, map[string]any{"session": sess.id})
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-sess.ctx.Done():
			s.logger.Info("SSE client disconnected", map[string]any{"session": sess.id})
			return
		case msg, ok := <-sess.send:
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg); err != nil {
				s.logger.Error("write SSE message", err, map[string]any{"session": sess.id})
				return
			}
			flusher.Flush()
		case <-heartbeat.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// HandleMessages accepts JSON-RPC requests from the client. Responses are
// delivered back through the SSE stream identified by the sessionId query
// parameter.
func (s *Server) HandleMessages(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.checkAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "missing sessionId", http.StatusBadRequest)
		return
	}
	sess := s.getSession(sessionID)
	if sess == nil {
		http.Error(w, "unknown session", http.StatusNotFound)
		return
	}

	// Read at most maxMessageBody+1 bytes so we can detect oversized requests.
	body, err := io.ReadAll(io.LimitReader(r.Body, maxMessageBody+1))
	if err != nil {
		s.logger.Error("read request body", err, map[string]any{"session": sessionID})
		s.sendToSession(sess, s.errorResponse(nil, ErrCodeParse, "Parse error"))
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	if int64(len(body)) > maxMessageBody {
		s.logger.Warn("request body too large", map[string]any{"session": sessionID, "limit_bytes": maxMessageBody})
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.logger.Error("parse JSON-RPC request", err, map[string]any{"session": sessionID})
		s.sendToSession(sess, s.errorResponse(nil, ErrCodeParse, "Parse error"))
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if req.JSONRPC != "2.0" {
		s.logger.Error("invalid JSON-RPC version", nil, map[string]any{
			"session": sessionID,
			"version": req.JSONRPC,
		})
		s.sendToSession(sess, s.errorResponse(req.ID, ErrCodeInvalidRequest, "Invalid Request"))
		w.WriteHeader(http.StatusAccepted)
		return
	}

	reqCtx := sessctx.WithID(r.Context(), sess.id)
	resp := s.dispatch(reqCtx, &req)
	if resp != nil {
		s.sendToSession(sess, resp)
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) dispatch(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	notification := req.IsNotification()
	s.logger.Debug("handling request", map[string]any{
		"method":       req.Method,
		"notification": notification,
	})

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		s.logger.Info("initialization complete", nil)
		if notification {
			return nil
		}
		return s.okResponse(req.ID, map[string]any{})
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(req)
	default:
		s.logger.Warn("unknown method", map[string]any{"method": req.Method})
		if notification {
			return nil
		}
		return s.errorResponse(req.ID, ErrCodeMethodNotFound, "Method not found")
	}
}

func (s *Server) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	var initReq InitializeRequest
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &initReq); err != nil {
			s.logger.Error("parse initialize params", err, nil)
			return s.errorResponse(req.ID, ErrCodeInvalidParams, "Invalid params")
		}
	}

	resp := InitializeResponse{ProtocolVersion: ProtocolVersion}
	resp.ServerInfo.Name = "allure-mcp-server"
	resp.ServerInfo.Version = Version

	s.logger.Info("initialize response sent", map[string]any{
		"version": resp.ProtocolVersion,
		"client":  initReq.ClientInfo.Name,
	})

	return s.okResponse(req.ID, resp)
}

func (s *Server) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	toolsList := s.registry.ListTools()
	result := ToolsListResponse{Tools: make([]Tool, 0, len(toolsList))}
	for _, t := range toolsList {
		result.Tools = append(result.Tools, Tool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
			Annotations: t.Annotations,
			Meta:        t.Meta,
		})
	}

	s.logger.Debug("tools/list response", map[string]any{"count": len(result.Tools)})
	return s.okResponse(req.ID, result)
}

func (s *Server) handleResourcesList(req *JSONRPCRequest) *JSONRPCResponse {
	resources := s.registry.ListResources()
	result := ResourcesListResponse{Resources: make([]MCPResource, 0, len(resources))}
	for _, r := range resources {
		result.Resources = append(result.Resources, MCPResource{
			URI:      r.URI,
			Name:     r.Name,
			MimeType: r.MimeType,
		})
	}
	s.logger.Debug("resources/list response", map[string]any{"count": len(result.Resources)})
	return s.okResponse(req.ID, result)
}

func (s *Server) handleResourcesRead(req *JSONRPCRequest) *JSONRPCResponse {
	if len(req.Params) == 0 {
		return s.errorResponse(req.ID, ErrCodeInvalidParams, "Missing params")
	}
	var readReq ResourcesReadRequest
	if err := json.Unmarshal(req.Params, &readReq); err != nil {
		s.logger.Error("parse resources/read params", err, nil)
		return s.errorResponse(req.ID, ErrCodeInvalidParams, "Invalid params")
	}
	res := s.registry.GetResource(readReq.URI)
	if res == nil {
		s.logger.Warn("resource not found", map[string]any{"uri": readReq.URI})
		return s.errorResponse(req.ID, ErrCodeInvalidParams, fmt.Sprintf("Resource not found: %s", readReq.URI))
	}
	s.logger.Debug("resources/read", map[string]any{"uri": readReq.URI})
	return s.okResponse(req.ID, ResourcesReadResponse{
		Contents: []ResourceContent{{
			URI:      res.URI,
			MimeType: res.MimeType,
			Text:     res.GetHTML(),
		}},
	})
}

func (s *Server) handleToolsCall(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	if len(req.Params) == 0 {
		return s.errorResponse(req.ID, ErrCodeInvalidParams, "Missing params")
	}
	var callReq ToolCallRequest
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		s.logger.Error("parse tools/call params", err, nil)
		return s.errorResponse(req.ID, ErrCodeInvalidParams, "Invalid params")
	}

	s.logger.Info("tool call", map[string]any{"tool": callReq.Name})

	tool := s.registry.GetTool(callReq.Name)
	if tool == nil {
		s.logger.Warn("unknown tool", map[string]any{"tool": callReq.Name})
		return s.errorResponse(req.ID, ErrCodeInvalidParams, fmt.Sprintf("Unknown tool: %s", callReq.Name))
	}

	// Normalise nil Arguments (e.g. params.arguments omitted or null) to an
	// empty JSON object so handlers can safely call json.Unmarshal on it.
	if callReq.Arguments == nil {
		callReq.Arguments = json.RawMessage("{}")
	}

	result, err := tool.Handler(ctx, callReq.Arguments)
	if err != nil {
		s.logger.Error("tool execution failed", err, map[string]any{"tool": callReq.Name})
		return s.okResponse(req.ID, ToolCallResponse{
			IsError: true,
			Content: []any{TextContent{
				Type: "text",
				Text: fmt.Sprintf("Tool execution failed: %v", err),
			}},
		})
	}

	resp := ToolCallResponse{
		Content: []any{TextContent{
			Type: "text",
			Text: resultToJSON(result),
		}},
	}
	// Forward the tool's _meta (e.g. ui.resourceUri) so Claude Desktop
	// knows which widget to open and routes this result to its ontoolresult callback.
	if tool.Meta != nil {
		resp.Meta = tool.Meta
	}
	return s.okResponse(req.ID, resp)
}

// ---------------------------------------------------------------------------
// Streamable HTTP transport (MCP spec 2025-03-26)
// ---------------------------------------------------------------------------

// HandleMCP serves the standard MCP Streamable HTTP transport.
// All JSON-RPC traffic is handled by a single endpoint supporting POST, GET,
// and DELETE methods:
//
//	POST /mcp   — send a JSON-RPC message; receives an inline JSON response
//	GET  /mcp   — reserved for server-initiated SSE (returns 405 for now)
//	DELETE /mcp — explicit session termination (clears stored tokens)
//
// This transport coexists with the legacy HTTP+SSE endpoints (/sse, /messages)
// so existing Claude Desktop configurations continue to work.
//
// Security note: callers should also validate the Origin header to guard against
// DNS-rebinding when running locally.
func (s *Server) HandleMCP(w http.ResponseWriter, r *http.Request) {
	s.setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if !s.checkAuth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodPost:
		s.handleMCPPost(w, r)
	case http.MethodGet:
		// Server-initiated SSE is not yet implemented.
		// Per spec, returning 405 is valid.
		w.Header().Set("Allow", "POST, DELETE, OPTIONS")
		http.Error(w, "server-initiated SSE not supported; use POST for requests", http.StatusMethodNotAllowed)
	case http.MethodDelete:
		s.handleMCPDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMCPPost processes a single JSON-RPC message and writes an inline JSON response.
// For initialize requests, the server creates a session ID and returns it via
// the Mcp-Session-Id response header. Clients must then include that header on
// all subsequent requests.
func (s *Server) handleMCPPost(w http.ResponseWriter, r *http.Request) {
	// Per MCP spec 2025-03-26: clients must send Content-Type: application/json.
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxMessageBody+1))
	if err != nil {
		s.logger.Error("read streamable-HTTP body", err, nil)
		writeMCPError(w, nil, ErrCodeParse, "read error", http.StatusBadRequest)
		return
	}
	if int64(len(body)) > maxMessageBody {
		s.logger.Warn("streamable-HTTP body too large", map[string]any{"limit_bytes": maxMessageBody})
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.logger.Error("parse streamable-HTTP JSON-RPC", err, nil)
		writeMCPError(w, nil, ErrCodeParse, "parse error", http.StatusBadRequest)
		return
	}
	if req.JSONRPC != "2.0" {
		writeMCPError(w, req.ID, ErrCodeInvalidRequest, "invalid JSON-RPC version", http.StatusOK)
		return
	}

	// For initialization, generate the session ID before dispatching so that
	// an X-Allure-Token present on the same request can be stored under the
	// new session ID immediately.
	sessID := r.Header.Get("Mcp-Session-Id")
	if req.Method == "initialize" {
		sessID = newSessionID()
	}

	// Persist per-user Allure token whenever the client sends it.
	// This can arrive on any request (most commonly on initialize or the first
	// tool call), so we store it every time the header is present.
	if tok := r.Header.Get("X-Allure-Token"); tok != "" && sessID != "" {
		s.registry.SetSessionToken(sessID, tok)
	}

	s.logger.Debug("streamable-HTTP request", map[string]any{
		"method":  req.Method,
		"session": sessID,
	})

	ctx := sessctx.WithID(r.Context(), sessID)
	resp := s.dispatch(ctx, &req)

	// Notifications (no id) require no response body; reply 202 Accepted.
	if req.IsNotification() || resp == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// Return the freshly created session ID to the client on initialize.
	if req.Method == "initialize" && sessID != "" {
		w.Header().Set("Mcp-Session-Id", sessID)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("write streamable-HTTP response", err, nil)
	}
}

// handleMCPDelete terminates a session and clears any stored per-session state.
func (s *Server) handleMCPDelete(w http.ResponseWriter, r *http.Request) {
	sessID := r.Header.Get("Mcp-Session-Id")
	if sessID != "" {
		s.registry.ClearSessionToken(sessID)
		s.logger.Info("streamable-HTTP session terminated", map[string]any{"session": sessID})
	}
	w.WriteHeader(http.StatusOK)
}

// writeMCPError writes a JSON-RPC error response with the specified HTTP status code.
func writeMCPError(w http.ResponseWriter, id json.RawMessage, code int, msg string, status int) {
	resp := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &JSONRPCError{Code: code, Message: msg},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------
// Shared helpers (used by both transports)
// ---------------------------------------------------------------------------

func (s *Server) okResponse(id json.RawMessage, result any) *JSONRPCResponse {
	return &JSONRPCResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func (s *Server) errorResponse(id json.RawMessage, code int, message string) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &JSONRPCError{Code: code, Message: message},
	}
}

func (s *Server) sendToSession(sess *session, resp *JSONRPCResponse) {
	if resp == nil {
		return
	}
	data, err := json.Marshal(resp)
	if err != nil {
		s.logger.Error("marshal response", err, map[string]any{"session": sess.id})
		return
	}
	select {
	case sess.send <- data:
	case <-sess.ctx.Done():
	default:
		s.logger.Warn("session send buffer full; dropping response", map[string]any{"session": sess.id})
	}
}

func resultToJSON(result any) string {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fallback, _ := json.Marshal(map[string]string{"error": "marshal failed: " + err.Error()})
		return string(fallback)
	}
	return string(b)
}

// --- sessions ---

func (s *Server) newSession(parent context.Context) *session {
	ctx, cancel := context.WithCancel(parent)
	sess := &session{
		id:     newSessionID(),
		send:   make(chan []byte, sessionSendBuffer),
		ctx:    ctx,
		cancel: cancel,
	}
	s.mu.Lock()
	s.sessions[sess.id] = sess
	s.mu.Unlock()
	return sess
}

func (s *Server) closeSession(sess *session) {
	s.mu.Lock()
	delete(s.sessions, sess.id)
	s.mu.Unlock()
	sess.cancel()
}

func (s *Server) getSession(id string) *session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

func newSessionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}

// --- auth & CORS ---

func (s *Server) checkAuth(r *http.Request) bool {
	if s.opts.AuthToken == "" {
		return true
	}
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	got := strings.TrimSpace(header[len(prefix):])
	return subtle.ConstantTimeCompare([]byte(got), []byte(s.opts.AuthToken)) == 1
}

func (s *Server) setCORSHeaders(w http.ResponseWriter) {
	origin := s.opts.CORSAllowOrigin
	if origin == "" {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Mcp-Session-Id, X-Allure-Token")
	if origin != "*" {
		w.Header().Set("Vary", "Origin")
	}
}
