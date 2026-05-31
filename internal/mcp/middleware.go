package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/audit"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
	sessctx "github.com/MimoJanra/TestOpsMCP/internal/session"
)

type requestHandler func(context.Context, *JSONRPCRequest) *JSONRPCResponse
type middlewareFunc func(next requestHandler) requestHandler

func buildChain(base requestHandler, middlewares ...middlewareFunc) requestHandler {
	h := base
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func panicRecoveryMiddleware(logger *core.Logger) middlewareFunc {
	return func(next requestHandler) requestHandler {
		return func(ctx context.Context, req *JSONRPCRequest) (resp *JSONRPCResponse) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic in handler", fmt.Errorf("%v", r), map[string]any{
						"method": req.Method,
						"stack":  string(debug.Stack()),
					})
					resp = &JSONRPCResponse{
						JSONRPC: "2.0",
						ID:      req.ID,
						Error:   &JSONRPCError{Code: ErrCodeInternal, Message: "internal server error"},
					}
				}
			}()
			return next(ctx, req)
		}
	}
}

func auditMiddleware(auditLog *audit.Logger) middlewareFunc {
	return func(next requestHandler) requestHandler {
		return func(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
			if auditLog == nil || req.IsNotification() {
				return next(ctx, req)
			}
			start := time.Now()
			resp := next(ctx, req)
			status := "ok"
			if resp != nil && resp.Error != nil {
				status = "error"
			}
			entry := audit.Entry{
				User:       sessctx.UserFromContext(ctx),
				SessionID:  sessctx.IDFromContext(ctx),
				RemoteAddr: sessctx.RemoteAddrFromContext(ctx),
				Method:     req.Method,
				Status:     status,
				DurationMS: time.Since(start).Milliseconds(),
			}
			if req.Method == "tools/call" && len(req.Params) > 0 {
				var p struct {
					Name string `json:"name"`
				}
				if err := json.Unmarshal(req.Params, &p); err == nil {
					entry.Tool = p.Name
				}
			}
			auditLog.Write(entry)
			return resp
		}
	}
}
