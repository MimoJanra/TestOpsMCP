package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/example/allure-mcp-server/internal/core"
	"github.com/example/allure-mcp-server/internal/tools"
)

type Server struct {
	registry *tools.Registry
	logger   *core.Logger
	stdin    io.Reader
	stdout   io.Writer
	mu       sync.Mutex
}

func NewServer(registry *tools.Registry, logger *core.Logger) *Server {
	return &Server{
		registry: registry,
		logger:   logger,
		stdin:    os.Stdin,
		stdout:   os.Stdout,
	}
}

func (s *Server) Start(ctx context.Context) error {
	scanner := bufio.NewScanner(s.stdin)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.logger.Error("Failed to parse request", err, nil)
			s.sendError(nil, -32700, "Parse error")
			continue
		}

		if req.JSONRPC != "2.0" {
			s.logger.Error("Invalid JSON-RPC version", nil, map[string]interface{}{
				"version": req.JSONRPC,
			})
			s.sendError(req.ID, -32600, "Invalid Request")
			continue
		}

		s.handleRequest(ctx, &req)
	}

	if err := scanner.Err(); err != nil {
		s.logger.Error("Scanner error", err, nil)
		return err
	}

	return nil
}

func (s *Server) handleRequest(ctx context.Context, req *JSONRPCRequest) {
	isNotification := len(req.ID) == 0 || string(req.ID) == "null"
	s.logger.Info("Handling request", map[string]interface{}{
		"method":       req.Method,
		"notification": isNotification,
	})

	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "notifications/initialized":
		s.logger.Info("Initialization complete", nil)
		return
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolsCall(ctx, req)
	default:
		s.logger.Error("Unknown method", nil, map[string]interface{}{
			"method": req.Method,
		})
		if !isNotification {
			s.sendError(req.ID, -32601, "Method not found")
		}
	}
}

func (s *Server) handleInitialize(req *JSONRPCRequest) {
	var initReq InitializeRequest
	if err := json.Unmarshal(req.Params, &initReq); err != nil {
		s.logger.Error("Failed to parse initialize params", err, nil)
		s.sendError(req.ID, -32602, "Invalid params")
		return
	}

	resp := InitializeResponse{
		ProtocolVersion: "2024-11-05",
	}
	resp.Capabilities.Tools = struct{}{}
	resp.ServerInfo.Name = "allure-mcp-server"
	resp.ServerInfo.Version = "1.0.0"

	s.logger.Info("Initialize response sent", map[string]interface{}{
		"version": resp.ProtocolVersion,
	})

	s.sendResponse(req.ID, resp)
}

func (s *Server) handleToolsList(req *JSONRPCRequest) {
	toolsList := s.registry.ListTools()
	tools := make([]Tool, 0, len(toolsList))

	for _, t := range toolsList {
		tools = append(tools, Tool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}

	resp := ToolsListResponse{
		Tools: tools,
	}

	s.logger.Info("Tools list response sent", map[string]interface{}{
		"count": len(tools),
	})

	s.sendResponse(req.ID, resp)
}

func (s *Server) handleToolsCall(ctx context.Context, req *JSONRPCRequest) {
	var callReq ToolCallRequest
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		s.logger.Error("Failed to parse tools/call params", err, nil)
		s.sendError(req.ID, -32602, "Invalid params")
		return
	}

	s.logger.Info("Tool call requested", map[string]interface{}{
		"tool": callReq.Name,
	})

	tool := s.registry.GetTool(callReq.Name)
	if tool == nil {
		s.logger.Error("Unknown tool", nil, map[string]interface{}{
			"tool": callReq.Name,
		})
		s.sendError(req.ID, -32602, fmt.Sprintf("Unknown tool: %s", callReq.Name))
		return
	}

	result, err := tool.Handler(ctx, callReq.Arguments)
	if err != nil {
		s.logger.Error("Tool execution failed", err, map[string]interface{}{
			"tool": callReq.Name,
		})
		content := ToolCallResponse{
			IsError: true,
			Content: []interface{}{
				TextContent{
					Type: "text",
					Text: fmt.Sprintf("Tool execution failed: %v", err),
				},
			},
		}
		s.sendResponse(req.ID, content)
		return
	}

	content := ToolCallResponse{
		Content: []interface{}{
			TextContent{
				Type: "text",
				Text: s.resultToJSON(result),
			},
		},
	}

	s.logger.Info("Tool executed successfully", map[string]interface{}{
		"tool": callReq.Name,
	})

	s.sendResponse(req.ID, content)
}

func (s *Server) resultToJSON(result interface{}) string {
	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%v"}`, err)
	}
	return string(bytes)
}

func (s *Server) sendResponse(id json.RawMessage, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	s.sendJSON(resp)
}

func (s *Server) sendError(id json.RawMessage, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}

	s.sendJSON(resp)
}

func (s *Server) sendJSON(v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bytes, err := json.Marshal(v)
	if err != nil {
		s.logger.Error("Failed to marshal response", err, nil)
		return
	}

	fmt.Fprintln(s.stdout, string(bytes))
}
