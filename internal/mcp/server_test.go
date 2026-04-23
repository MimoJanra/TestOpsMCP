package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/core"
	"github.com/MimoJanra/TestOpsMCP/internal/tools"
)

func newTestServer(t *testing.T, opts Options) *httptest.Server {
	t.Helper()
	logger := core.NewLogger(core.LevelError)
	registry := tools.NewRegistry(nil, logger)
	srv := NewServer(registry, logger, opts)

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", srv.HandleSSE)
	mux.HandleFunc("/messages", srv.HandleMessages)
	return httptest.NewServer(mux)
}

// readSSEEvent reads events from an SSE stream until it finds one with the
// requested event name, then returns its data field.
func readSSEEvent(t *testing.T, r io.Reader, want string, deadline time.Duration) string {
	t.Helper()
	done := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 4096), 1<<20)
		var event, data string
		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event:"):
				event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "data:"):
				data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			case line == "":
				if event == want {
					done <- data
					return
				}
				event, data = "", ""
			}
		}
		errCh <- scanner.Err()
	}()

	select {
	case d := <-done:
		return d
	case err := <-errCh:
		t.Fatalf("SSE scanner ended without %q event: %v", want, err)
	case <-time.After(deadline):
		t.Fatalf("timeout waiting for SSE event %q", want)
	}
	return ""
}

func TestServer_InitializeOverSSE(t *testing.T) {
	ts := newTestServer(t, Options{})
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/sse")
	if err != nil {
		t.Fatalf("GET /sse: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	endpoint := readSSEEvent(t, resp.Body, "endpoint", 2*time.Second)
	if !strings.HasPrefix(endpoint, "/messages?sessionId=") {
		t.Fatalf("unexpected endpoint: %q", endpoint)
	}

	reqBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	post, err := http.Post(ts.URL+endpoint, "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST /messages: %v", err)
	}
	io.Copy(io.Discard, post.Body)
	post.Body.Close()
	if post.StatusCode != http.StatusAccepted {
		t.Fatalf("POST status = %d, want 202", post.StatusCode)
	}

	msg := readSSEEvent(t, resp.Body, "message", 2*time.Second)
	var parsed JSONRPCResponse
	if err := json.Unmarshal([]byte(msg), &parsed); err != nil {
		t.Fatalf("unmarshal response: %v (%s)", err, msg)
	}
	if parsed.Error != nil {
		t.Fatalf("unexpected error: %+v", parsed.Error)
	}
	if string(parsed.ID) != "1" {
		t.Errorf("id = %s, want 1", parsed.ID)
	}
}

func TestServer_ToolsListOverSSE(t *testing.T) {
	ts := newTestServer(t, Options{})
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/sse")
	if err != nil {
		t.Fatalf("GET /sse: %v", err)
	}
	defer resp.Body.Close()

	endpoint := readSSEEvent(t, resp.Body, "endpoint", 2*time.Second)

	post, err := http.Post(ts.URL+endpoint, "application/json",
		strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	post.Body.Close()

	msg := readSSEEvent(t, resp.Body, "message", 2*time.Second)
	if !strings.Contains(msg, "run_allure_launch") {
		t.Errorf("tools/list response missing tools: %s", msg)
	}
}

func TestServer_MissingSessionID(t *testing.T) {
	ts := newTestServer(t, Options{})
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/messages", "application/json",
		bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`)))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestServer_UnknownSessionID(t *testing.T) {
	ts := newTestServer(t, Options{})
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/messages?sessionId=nonexistent", "application/json",
		bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`)))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServer_AuthRequired(t *testing.T) {
	ts := newTestServer(t, Options{AuthToken: "s3cret"})
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/sse")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no-token status = %d, want 401", resp.StatusCode)
	}

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/sse", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong-token status = %d, want 401", resp.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodGet, ts.URL+"/sse", nil)
	req.Header.Set("Authorization", "Bearer s3cret")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid-token status = %d, want 200", resp.StatusCode)
	}
}

func TestServer_UnknownMethodReturnsError(t *testing.T) {
	ts := newTestServer(t, Options{})
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/sse")
	if err != nil {
		t.Fatalf("GET /sse: %v", err)
	}
	defer resp.Body.Close()

	endpoint := readSSEEvent(t, resp.Body, "endpoint", 2*time.Second)
	body := `{"jsonrpc":"2.0","id":99,"method":"nope"}`
	post, err := http.Post(ts.URL+endpoint, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	post.Body.Close()

	msg := readSSEEvent(t, resp.Body, "message", 2*time.Second)
	var parsed JSONRPCResponse
	if err := json.Unmarshal([]byte(msg), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Error == nil || parsed.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("expected method-not-found error, got %+v", parsed.Error)
	}
}
