package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MimoJanra/TestOpsMCP/internal/audit"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
	sessctx "github.com/MimoJanra/TestOpsMCP/internal/session"
)

func TestPanicRecoveryMiddleware_RecoversAndReturnsInternalError(t *testing.T) {
	logger := core.NewLogger(core.LevelError)
	mw := panicRecoveryMiddleware(logger)

	panicking := func(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
		panic("boom")
	}
	handler := mw(panicking)

	req := &JSONRPCRequest{JSONRPC: "2.0", ID: json.RawMessage("1"), Method: "tools/call"}
	resp := handler(context.Background(), req)

	if resp == nil || resp.Error == nil {
		t.Fatalf("expected an error response after recovering from panic, got %+v", resp)
	}
	if resp.Error.Code != ErrCodeInternal {
		t.Errorf("Error.Code = %d, want %d", resp.Error.Code, ErrCodeInternal)
	}
}

func TestPanicRecoveryMiddleware_PassesThroughWithoutPanic(t *testing.T) {
	logger := core.NewLogger(core.LevelError)
	mw := panicRecoveryMiddleware(logger)

	want := &JSONRPCResponse{JSONRPC: "2.0", ID: json.RawMessage("1"), Result: "ok"}
	handler := mw(func(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
		return want
	})

	got := handler(context.Background(), &JSONRPCRequest{ID: json.RawMessage("1")})
	if got != want {
		t.Errorf("expected passthrough response, got %+v", got)
	}
}

func newTestAuditLogger(t *testing.T) (*audit.Logger, string) {
	t.Helper()
	dir := t.TempDir()
	l, err := audit.NewLogger(dir, 1)
	if err != nil {
		t.Fatalf("audit.NewLogger: %v", err)
	}
	t.Cleanup(l.Close)
	return l, dir
}

func readAuditEntries(t *testing.T, dir string) []audit.Entry {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "audit-*.jsonl"))
	if err != nil || len(matches) == 0 {
		t.Fatalf("no audit log file found in %s (err=%v)", dir, err)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	var entries []audit.Entry
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var e audit.Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("unmarshal audit entry %q: %v", line, err)
		}
		entries = append(entries, e)
	}
	return entries
}

func TestAuditMiddleware_NilLoggerPassesThrough(t *testing.T) {
	mw := auditMiddleware(nil)
	called := false
	handler := mw(func(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
		called = true
		return &JSONRPCResponse{ID: req.ID, Result: "ok"}
	})
	resp := handler(context.Background(), &JSONRPCRequest{ID: json.RawMessage("1")})
	if !called {
		t.Error("expected next handler to be called")
	}
	if resp == nil || resp.Result != "ok" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestAuditMiddleware_LogsNotification(t *testing.T) {
	dir := t.TempDir()
	logger, err := audit.NewLogger(dir, 1)
	if err != nil {
		t.Fatalf("audit.NewLogger: %v", err)
	}
	defer logger.Close()

	mw := auditMiddleware(logger)
	handler := mw(func(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
		return &JSONRPCResponse{Result: "ignored for notifications"}
	})

	// No ID => notification.
	resp := handler(context.Background(), &JSONRPCRequest{Method: "notifications/initialized"})
	if resp != nil {
		t.Errorf("expected nil response for notification, got %+v", resp)
	}

	entries := readAuditEntries(t, dir)
	if len(entries) != 1 {
		t.Fatalf("got %d audit entries, want 1", len(entries))
	}
	if entries[0].Status != "notification" {
		t.Errorf("Status = %q, want %q", entries[0].Status, "notification")
	}
}

func TestAuditMiddleware_LogsErrorAndToolName(t *testing.T) {
	dir := t.TempDir()
	logger, err := audit.NewLogger(dir, 1)
	if err != nil {
		t.Fatalf("audit.NewLogger: %v", err)
	}
	defer logger.Close()

	mw := auditMiddleware(logger)
	handler := mw(func(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
		return &JSONRPCResponse{ID: req.ID, Error: &JSONRPCError{Code: ErrCodeInternal, Message: "nope"}}
	})

	ctx := sessctx.WithRemoteAddr(sessctx.WithUser(sessctx.WithID(context.Background(), "sess-1"), "alice"), "1.2.3.4")
	req := &JSONRPCRequest{
		ID:     json.RawMessage("7"),
		Method: "tools/call",
		Params: json.RawMessage(`{"name":"list_running_tasks","arguments":{}}`),
	}
	resp := handler(ctx, req)
	if resp == nil || resp.Error == nil {
		t.Fatalf("expected error response to pass through, got %+v", resp)
	}

	entries := readAuditEntries(t, dir)
	if len(entries) != 1 {
		t.Fatalf("got %d audit entries, want 1", len(entries))
	}
	e := entries[0]
	if e.Status != "error" {
		t.Errorf("Status = %q, want %q", e.Status, "error")
	}
	if e.Tool != "list_running_tasks" {
		t.Errorf("Tool = %q, want %q", e.Tool, "list_running_tasks")
	}
	if e.User != "alice" || e.SessionID != "sess-1" || e.RemoteAddr != "1.2.3.4" {
		t.Errorf("unexpected session metadata: %+v", e)
	}
}

func TestAuditMiddleware_LogsOKStatus(t *testing.T) {
	l, dir := newTestAuditLogger(t)

	mw := auditMiddleware(l)
	handler := mw(func(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
		return &JSONRPCResponse{ID: req.ID, Result: "fine"}
	})
	handler(context.Background(), &JSONRPCRequest{ID: json.RawMessage("1"), Method: "tools/list"})

	entries := readAuditEntries(t, dir)
	if len(entries) != 1 || entries[0].Status != "ok" {
		t.Fatalf("unexpected entries: %+v", entries)
	}
}
