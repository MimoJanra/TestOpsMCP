package mcp

import (
	"bufio"
	"bytes"
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

	reader := bufio.NewReaderSize(os.Stdin, 64<<10)

	var wg sync.WaitGroup
	for {
		line, tooLong, readErr := readLine(reader, maxMessageBody)
		if readErr != nil && readErr != io.EOF {
			wg.Wait()
			sh.logger.Error("stdio read error", readErr, nil)
			return readErr
		}
		if tooLong {
			// Answer and keep serving: a bufio.Scanner would stop at the first
			// oversized line and take the whole server (and client) down.
			sh.logger.Warn("stdio message too large", map[string]any{"limit_bytes": maxMessageBody})
			sh.respond(&JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &JSONRPCError{Code: ErrCodeInvalidRequest, Message: fmt.Sprintf("message exceeds %d bytes", maxMessageBody)},
			})
		} else if len(bytes.TrimSpace(line)) > 0 {
			sh.handleLine(ctx, &wg, line)
		}
		if readErr == io.EOF {
			break
		}
	}
	wg.Wait()
	return nil
}

// readLine reads one newline-terminated line of at most limit bytes. A longer
// line is consumed to its end and reported as tooLong instead of returned.
func readLine(r *bufio.Reader, limit int) (line []byte, tooLong bool, err error) {
	for {
		chunk, isPrefix, e := r.ReadLine()
		if !tooLong {
			if len(line)+len(chunk) > limit {
				tooLong, line = true, nil
			} else {
				line = append(line, chunk...)
			}
		}
		if e != nil {
			return line, tooLong, e
		}
		if !isPrefix {
			return line, tooLong, nil
		}
	}
}

func (sh *StdioHandler) handleLine(ctx context.Context, wg *sync.WaitGroup, line []byte) {
	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		sh.logger.Error("parse JSON-RPC request", err, nil)
		sh.respond(&JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error:   &JSONRPCError{Code: ErrCodeParse, Message: "Parse error"},
		})
		return
	}

	if req.JSONRPC != "2.0" {
		sh.logger.Error("invalid JSON-RPC version", nil, map[string]any{"version": req.JSONRPC})
		sh.respond(&JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &JSONRPCError{Code: ErrCodeInvalidRequest, Message: "Invalid Request"},
		})
		return
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
