package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sessctx "github.com/MimoJanra/TestOpsMCP/internal/session"
)

func postMCP(t *testing.T, s *Server, body, sessionID, authToken string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)
	return w
}

const initBody = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{"elicitation":{}}}}`

// initialize used to hand out a session id that was never registered, so every
// later lookup for it missed: elicitation answered "reject" (rendered as
// "Cancelled.") and sampling failed outright on this transport.
func TestHandleMCP_InitializeRegistersSession(t *testing.T) {
	s := newServerForTest(t, Options{})
	w := postMCP(t, s, initBody, "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	sessID := w.Header().Get("Mcp-Session-Id")
	if sessID == "" {
		t.Fatal("no Mcp-Session-Id header on the initialize response")
	}
	sess := s.getSession(sessID)
	if sess == nil {
		t.Fatal("initialize did not register the session")
	}
	if !sess.expirable {
		t.Error("a streamable-HTTP session must be expirable")
	}
	// Capabilities from initialize must land on the session, so clientSupports
	// can tell an elicitation-capable client from one that never declared it.
	if !sess.clientSupports("elicitation") {
		t.Error("declared elicitation capability was not recorded")
	}
	if sess.clientSupports("sampling") {
		t.Error("sampling was not declared but is reported as supported")
	}
}

func TestHandleMCP_UnknownSessionIsNotFound(t *testing.T) {
	s := newServerForTest(t, Options{})
	w := postMCP(t, s, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, "deadbeef", "")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 so the client re-initializes", w.Code)
	}
}

// A session id is a capability: whoever holds it drives that session and the
// Allure token stored under it. It must stay with the user who opened it.
func TestHandleMCP_SessionIsBoundToItsUser(t *testing.T) {
	s := newServerForTest(t, Options{Users: []User{
		{Name: "alice", Token: "alice-token"},
		{Name: "bob", Token: "bob-token"},
	}})

	w := postMCP(t, s, initBody, "", "alice-token")
	sessID := w.Header().Get("Mcp-Session-Id")
	if sessID == "" {
		t.Fatal("no session id returned")
	}

	if got := postMCP(t, s, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, sessID, "alice-token").Code; got != http.StatusOK {
		t.Errorf("owner: status = %d, want 200", got)
	}
	if got := postMCP(t, s, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`, sessID, "bob-token").Code; got != http.StatusForbidden {
		t.Errorf("other user: status = %d, want 403", got)
	}

	req := httptest.NewRequest(http.MethodDelete, "/mcp", nil)
	req.Header.Set("Mcp-Session-Id", sessID)
	req.Header.Set("Authorization", "Bearer bob-token")
	w = httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("DELETE by another user: status = %d, want 403", w.Code)
	}
	if s.getSession(sessID) == nil {
		t.Error("another user's DELETE must not close the session")
	}
}

func TestHandleMCP_DeleteClosesSession(t *testing.T) {
	s := newServerForTest(t, Options{})
	sessID := postMCP(t, s, initBody, "", "").Header().Get("Mcp-Session-Id")

	req := httptest.NewRequest(http.MethodDelete, "/mcp", nil)
	req.Header.Set("Mcp-Session-Id", sessID)
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if s.getSession(sessID) != nil {
		t.Error("DELETE did not remove the session")
	}
}

// GET /mcp is the server→client channel. With it open, an elicitation request
// reaches the client and its reply unblocks the waiting tool call.
func TestHandleMCP_GetStreamCarriesElicitation(t *testing.T) {
	s := newServerForTest(t, Options{})
	sessID := postMCP(t, s, initBody, "", "").Header().Get("Mcp-Session-Id")
	sess := s.getSession(sessID)
	if sess == nil {
		t.Fatal("no session")
	}

	srv := httptest.NewServer(http.HandlerFunc(s.HandleMCP))
	defer srv.Close()

	streamReq, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("build stream request: %v", err)
	}
	streamReq.Header.Set("Mcp-Session-Id", sessID)
	resp, err := http.DefaultClient.Do(streamReq)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stream status = %d, want 200", resp.StatusCode)
	}

	deadline := time.Now().Add(2 * time.Second)
	for !sess.hasStream() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !sess.hasStream() {
		t.Fatal("stream was not registered on the session")
	}

	ctx := sessctx.WithID(context.Background(), sessID)
	done := make(chan *ElicitResult, 1)
	errCh := make(chan error, 1)
	go func() {
		r, err := s.Elicit(ctx, ElicitRequest{Message: "delete it?"})
		done <- r
		errCh <- err
	}()

	// Read the elicitation request off the SSE stream.
	buf := make([]byte, 4096)
	var payload string
	for time.Now().Before(deadline) {
		n, err := resp.Body.Read(buf)
		if n > 0 && strings.Contains(string(buf[:n]), "elicitation/create") {
			payload = string(buf[:n])
			break
		}
		if err != nil {
			t.Fatalf("read stream: %v", err)
		}
	}
	if payload == "" {
		t.Fatal("elicitation/create never arrived on the stream")
	}
	data := payload[strings.Index(payload, "data: ")+len("data: "):]
	data = data[:strings.Index(data, "\n")]
	var notif struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(data), &notif); err != nil {
		t.Fatalf("decode notification %q: %v", data, err)
	}

	// The client answers through a normal POST, as it would in real use.
	reply := `{"jsonrpc":"2.0","id":"` + notif.ID + `","result":{"action":"accept"}}`
	postMCP(t, s, reply, sessID, "")

	select {
	case r := <-done:
		if err := <-errCh; err != nil {
			t.Fatalf("Elicit: %v", err)
		}
		if r.Action != "accept" {
			t.Errorf("Action = %q, want accept", r.Action)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Elicit did not return after the client replied")
	}
}

func TestExpireIdleSessions(t *testing.T) {
	s := newServerForTest(t, Options{})

	idle := s.newSession(context.Background(), newSessionID(), "")
	idle.expirable = true
	idle.lastSeen.Store(time.Now().Add(-2 * streamableSessionIdleTimeout).UnixNano())

	streaming := s.newSession(context.Background(), newSessionID(), "")
	streaming.expirable = true
	streaming.lastSeen.Store(time.Now().Add(-2 * streamableSessionIdleTimeout).UnixNano())
	streaming.attachStream()

	connected := s.newSession(context.Background(), newSessionID(), "") // SSE/stdio
	connected.lastSeen.Store(time.Now().Add(-2 * streamableSessionIdleTimeout).UnixNano())

	if n := s.expireIdleSessions(time.Now()); n != 1 {
		t.Errorf("expired %d sessions, want 1", n)
	}
	if s.getSession(idle.id) != nil {
		t.Error("idle streamable session should have been expired")
	}
	if s.getSession(streaming.id) == nil {
		t.Error("a session with an open stream must survive")
	}
	if s.getSession(connected.id) == nil {
		t.Error("a connection-bound session must survive")
	}
}

// A token handed to the server on connect used to stay in memory for the whole
// process lifetime, long after the client was gone.
func TestCloseSessionClearsStoredToken(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID(), "")
	s.registry.SetSessionToken(sess.id, "secret-token")

	s.closeSession(sess)

	if s.registry.HasSessionToken(sess.id) {
		t.Error("the session's Allure token is still stored after closeSession")
	}
}
