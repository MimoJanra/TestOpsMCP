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
	sess, ctx := sh.registry.StdioSession(context.Background())
	defer sh.registry.closeSession(sess)

	// Drain server-initiated messages (elicitation/create, sampling/createMessage
	// requests) and write them to stdout, through the same locked writer used
	// for ordinary responses so lines never interleave.
	go func() {
		for {
			select {
			case msg, ok := <-sess.send:
				if !ok {
					return
				}
				sh.writeRaw(msg)
			case <-sess.ctx.Done():
				return
			}
		}
	}()

	scanner := bufio.NewScanner(io.Reader(os.Stdin))
	scanner.Buffer(make([]byte, 4096), 1<<20) // 1 MiB max line

	var wg sync.WaitGroup
	for scanner.Scan() {
		// Copy: scanner.Bytes() is reused on the next Scan(), but json.Unmarshal
		// lets json.RawMessage fields (Params, Result) alias into it — and the
		// request is read again later, in a goroutine, well after Scan() moves on.
		line := append([]byte(nil), scanner.Bytes()...)

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
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

		// Dispatch on its own goroutine: a tool call that blocks awaiting an
		// elicitation/sampling reply (delivered as a later stdin line, routed
		// through the same session by handleJSONRPCResponse) would otherwise
		// deadlock against this very read loop.
		wg.Add(1)
		go func(req JSONRPCRequest) {
			defer wg.Done()
			resp := sh.registry.dispatch(ctx, &req)
			if resp != nil {
				sh.respond(resp)
			}
		}(req)
	}
	wg.Wait()

	if err := scanner.Err(); err != nil {
		sh.logger.Error("stdio read error", err, nil)
		return err
	}
	return nil
}

func (sh *StdioHandler) respond(resp *JSONRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		sh.logger.Error("marshal response", err, nil)
		return
	}
	sh.writeRaw(data)
}

func (sh *StdioHandler) writeRaw(data []byte) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	_, _ = fmt.Fprintln(os.Stdout, string(data))
}
