package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/MimoJanra/TestOpsMCP/internal/core"
)

// StdioHandler runs the MCP server in stdio mode: reads JSON-RPC requests from
// stdin and writes responses to stdout, one JSON object per line.
type StdioHandler struct {
	registry *Server
	logger   *core.Logger
	mu       sync.Mutex
}

func NewStdioHandler(registry *Server, logger *core.Logger) *StdioHandler {
	return &StdioHandler{
		registry: registry,
		logger:   logger,
	}
}

// Run reads from stdin and processes requests until EOF.
func (sh *StdioHandler) Run() error {
	scanner := bufio.NewScanner(io.Reader(os.Stdin))
	scanner.Buffer(make([]byte, 4096), 1<<20) // 1 MiB max line

	for scanner.Scan() {
		var req JSONRPCRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			sh.logger.Error("parse JSON-RPC request", err, nil)
			sh.respond(&JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &JSONRPCError{Code: ErrCodeParse, Message: "Parse error"},
			})
			continue
		}

		if req.JSONRPC != "2.0" {
			sh.logger.Error("invalid JSON-RPC version", nil, map[string]any{"version": req.JSONRPC})
			sh.respond(&JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &JSONRPCError{Code: ErrCodeInvalidRequest, Message: "Invalid Request"},
			})
			continue
		}

		resp := sh.registry.dispatch(context.Background(), &req)
		if resp != nil {
			sh.respond(resp)
		}
	}

	if err := scanner.Err(); err != nil {
		sh.logger.Error("stdio read error", err, nil)
		return err
	}
	return nil
}

func (sh *StdioHandler) respond(resp *JSONRPCResponse) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	bytes, err := json.Marshal(resp)
	if err != nil {
		sh.logger.Error("marshal response", err, nil)
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, string(bytes))
}
