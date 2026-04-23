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
	"github.com/MimoJanra/TestOpsMCP/internal/tools"
)

const (
	sessionSendBuffer = 16
	heartbeatInterval = 25 * time.Second
	maxMessageBody    = 1 << 20 // 1 MiB
)

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

	body, err := io.ReadAll(io.LimitReader(r.Body, maxMessageBody))
	if err != nil {
		s.logger.Error("read request body", err, map[string]any{"session": sessionID})
		s.sendToSession(sess, s.errorResponse(nil, ErrCodeParse, "Parse error"))
		http.Error(w, "read error", http.StatusBadRequest)
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

	resp := s.dispatch(r.Context(), &req)
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
	resp.Capabilities.Tools = struct{}{}
	resp.ServerInfo.Name = "allure-mcp-server"
	resp.ServerInfo.Version = "1.0.0"

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
		})
	}

	s.logger.Debug("tools/list response", map[string]any{"count": len(result.Tools)})
	return s.okResponse(req.ID, result)
}

func (s *Server) handleToolsCall(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
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

	return s.okResponse(req.ID, ToolCallResponse{
		Content: []any{TextContent{
			Type: "text",
			Text: resultToJSON(result),
		}},
	})
}

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
	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":"marshal failed: %v"}`, err)
	}
	return string(bytes)
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
		return fmt.Sprintf("sess-%d", time.Now().UnixNano())
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
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	if origin != "*" {
		w.Header().Set("Vary", "Origin")
	}
}
