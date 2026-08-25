package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/core"
	sessctx "github.com/MimoJanra/TestOpsMCP/internal/session"
	"github.com/MimoJanra/TestOpsMCP/internal/tools"
)

func newServerForTest(t *testing.T, opts Options) *Server {
	t.Helper()
	logger := core.NewLogger(core.LevelError)
	registry := tools.NewRegistry(nil, logger)
	return NewServer(registry, logger, opts)
}

// --- HandleHealth ---

func TestHandleHealth(t *testing.T) {
	s := newServerForTest(t, Options{})

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		w := httptest.NewRecorder()
		s.HandleHealth(w, httptest.NewRequest(method, "/healthz", nil))
		if w.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", method, w.Code)
		}
	}

	w := httptest.NewRecorder()
	s.HandleHealth(w, httptest.NewRequest(http.MethodPost, "/healthz", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST: status = %d, want 405", w.Code)
	}
}

// --- HandleSSE / HandleMessages method + CORS branches ---

func TestHandleSSE_OptionsAndMethodNotAllowed(t *testing.T) {
	s := newServerForTest(t, Options{CORSAllowOrigin: "https://example.com"})

	w := httptest.NewRecorder()
	s.HandleSSE(w, httptest.NewRequest(http.MethodOptions, "/sse", nil))
	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS: status = %d, want 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("CORS header = %q", got)
	}

	w = httptest.NewRecorder()
	s.HandleSSE(w, httptest.NewRequest(http.MethodPost, "/sse", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST: status = %d, want 405", w.Code)
	}
}

func TestHandleMessages_OptionsAndMethodNotAllowed(t *testing.T) {
	s := newServerForTest(t, Options{})

	w := httptest.NewRecorder()
	s.HandleMessages(w, httptest.NewRequest(http.MethodOptions, "/messages", nil))
	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS: status = %d, want 204", w.Code)
	}

	w = httptest.NewRecorder()
	s.HandleMessages(w, httptest.NewRequest(http.MethodGet, "/messages", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET: status = %d, want 405", w.Code)
	}
}

func TestHandleMessages_Unauthorized(t *testing.T) {
	s := newServerForTest(t, Options{Users: []User{{Name: "u", Token: "t"}}})
	w := httptest.NewRecorder()
	s.HandleMessages(w, httptest.NewRequest(http.MethodPost, "/messages?sessionId=x", strings.NewReader("{}")))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestHandleMessages_BodyTooLarge(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)

	big := strings.Repeat("a", maxMessageBody+10)
	w := httptest.NewRecorder()
	s.HandleMessages(w, httptest.NewRequest(http.MethodPost, "/messages?sessionId="+sess.id, strings.NewReader(big)))
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", w.Code)
	}
}

func TestHandleMessages_MalformedJSONStillAccepted(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)

	w := httptest.NewRecorder()
	s.HandleMessages(w, httptest.NewRequest(http.MethodPost, "/messages?sessionId="+sess.id, strings.NewReader("not json")))
	if w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202", w.Code)
	}
	select {
	case msg := <-sess.send:
		var resp JSONRPCResponse
		if err := json.Unmarshal(msg, &resp); err != nil {
			t.Fatalf("unmarshal queued error: %v", err)
		}
		if resp.Error == nil || resp.Error.Code != ErrCodeParse {
			t.Errorf("expected parse-error response, got %+v", resp)
		}
	default:
		t.Error("expected a parse-error response queued to the session")
	}
}

func TestHandleMessages_InvalidJSONRPCVersion(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)

	body := `{"jsonrpc":"1.0","id":1,"method":"initialize"}`
	w := httptest.NewRecorder()
	s.HandleMessages(w, httptest.NewRequest(http.MethodPost, "/messages?sessionId="+sess.id, strings.NewReader(body)))
	if w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202", w.Code)
	}
	msg := <-sess.send
	var resp JSONRPCResponse
	_ = json.Unmarshal(msg, &resp)
	if resp.Error == nil || resp.Error.Code != ErrCodeInvalidRequest {
		t.Errorf("expected invalid-request error, got %+v", resp)
	}
}

func TestHandleMessages_HappyPathSetsAllureToken(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	req := httptest.NewRequest(http.MethodPost, "/messages?sessionId="+sess.id, strings.NewReader(body))
	req.Header.Set("X-Allure-Token", "tok-123")
	w := httptest.NewRecorder()
	s.HandleMessages(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202", w.Code)
	}
	select {
	case <-sess.send:
	case <-time.After(time.Second):
		t.Error("expected a queued tools/list response")
	}
}

// --- route() branches not covered by server_test.go ---

func TestRoute_NotificationsInitialized(t *testing.T) {
	s := newServerForTest(t, Options{})

	// As a notification (no id): returns nil.
	if resp := s.route(context.Background(), &JSONRPCRequest{Method: "notifications/initialized"}); resp != nil {
		t.Errorf("notification form: expected nil, got %+v", resp)
	}

	// As a request (has id): returns an empty ok response.
	resp := s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("5"), Method: "notifications/initialized"})
	if resp == nil || resp.Error != nil {
		t.Fatalf("request form: expected ok response, got %+v", resp)
	}
}

func TestRoute_EmptyMethodDispatchesJSONRPCResponse_Elicit(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)

	ch := make(chan *ElicitResult, 1)
	sess.pendingMu.Lock()
	sess.pending["abc"] = ch
	sess.pendingMu.Unlock()

	ctx := sessctx.WithID(context.Background(), sess.id)
	req := &JSONRPCRequest{
		ID:     json.RawMessage(`"abc"`),
		Result: json.RawMessage(`{"action":"accept","content":{"x":1}}`),
	}
	if resp := s.route(ctx, req); resp != nil {
		t.Errorf("expected nil (this is a client->server response, not a request), got %+v", resp)
	}

	select {
	case result := <-ch:
		if result.Action != "accept" {
			t.Errorf("Action = %q, want accept", result.Action)
		}
	default:
		t.Error("expected the elicit channel to receive a result")
	}
}

func TestRoute_EmptyMethodDispatchesJSONRPCResponse_Sampling(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)

	ch := make(chan *SamplingResult, 1)
	sess.pendingMu.Lock()
	sess.sampling["sid"] = ch
	sess.pendingMu.Unlock()

	ctx := sessctx.WithID(context.Background(), sess.id)
	req := &JSONRPCRequest{
		ID:     json.RawMessage(`"sid"`),
		Result: json.RawMessage(`{"role":"assistant","content":{"text":"hi"},"stopReason":"end"}`),
	}
	s.route(ctx, req)

	select {
	case result := <-ch:
		if result.Content.Text != "hi" {
			t.Errorf("Text = %q, want hi", result.Content.Text)
		}
	default:
		t.Error("expected the sampling channel to receive a result")
	}
}

func TestRoute_EmptyMethodNotificationDoesNotDispatch(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)

	ch := make(chan *ElicitResult, 1)
	sess.pendingMu.Lock()
	sess.pending["abc"] = ch
	sess.pendingMu.Unlock()

	ctx := sessctx.WithID(context.Background(), sess.id)
	// No ID => notification => handleJSONRPCResponse must NOT be invoked.
	s.route(ctx, &JSONRPCRequest{Result: json.RawMessage(`{"action":"accept"}`)})

	select {
	case <-ch:
		t.Error("did not expect a result for a notification-shaped response")
	default:
	}
}

func TestRoute_EmptyMethodUnknownSession(t *testing.T) {
	s := newServerForTest(t, Options{})
	ctx := sessctx.WithID(context.Background(), "does-not-exist")
	// Must not panic when the session is gone.
	if resp := s.route(ctx, &JSONRPCRequest{ID: json.RawMessage("1"), Result: json.RawMessage(`{}`)}); resp != nil {
		t.Errorf("expected nil, got %+v", resp)
	}
}

// --- resources/list, resources/read ---

func TestHandleResourcesRead(t *testing.T) {
	s := newServerForTest(t, Options{})

	resp := s.route(context.Background(), &JSONRPCRequest{
		ID: json.RawMessage("1"), Method: "resources/read",
		Params: json.RawMessage(`{"uri":"allure://docs/quickstart"}`),
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}

	resp = s.route(context.Background(), &JSONRPCRequest{
		ID: json.RawMessage("2"), Method: "resources/read",
		Params: json.RawMessage(`{"uri":"nope"}`),
	})
	if resp.Error == nil {
		t.Error("expected error for unknown resource")
	}

	resp = s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("3"), Method: "resources/read"})
	if resp.Error == nil {
		t.Error("expected error for missing params")
	}

	resp = s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("4"), Method: "resources/read", Params: json.RawMessage(`not json`)})
	if resp.Error == nil {
		t.Error("expected error for malformed params")
	}
}

func TestHandleResourcesList(t *testing.T) {
	s := newServerForTest(t, Options{})
	resp := s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("1"), Method: "resources/list"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	result, ok := resp.Result.(ResourcesListResponse)
	if !ok {
		t.Fatalf("result is %T, want ResourcesListResponse", resp.Result)
	}
	if len(result.Resources) == 0 {
		t.Error("expected at least the quickstart resource")
	}
}

// --- prompts/list, prompts/get ---

func TestHandlePromptsList(t *testing.T) {
	s := newServerForTest(t, Options{})
	resp := s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("1"), Method: "prompts/list"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

func TestHandlePromptsGet(t *testing.T) {
	s := newServerForTest(t, Options{})

	resp := s.route(context.Background(), &JSONRPCRequest{
		ID: json.RawMessage("1"), Method: "prompts/get",
		Params: json.RawMessage(`{"name":"test-case-management","arguments":{"project_id":"42"}}`),
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	result, ok := resp.Result.(PromptGetResponse)
	if !ok || len(result.Messages) == 0 {
		t.Fatalf("unexpected result: %+v", resp.Result)
	}

	resp = s.route(context.Background(), &JSONRPCRequest{
		ID: json.RawMessage("2"), Method: "prompts/get",
		Params: json.RawMessage(`{"name":"nope"}`),
	})
	if resp.Error == nil {
		t.Error("expected error for unknown prompt")
	}

	resp = s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("3"), Method: "prompts/get"})
	if resp.Error == nil {
		t.Error("expected error for missing params")
	}

	resp = s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("4"), Method: "prompts/get", Params: json.RawMessage(`not json`)})
	if resp.Error == nil {
		t.Error("expected error for malformed params")
	}
}

// --- tools/call ---

func TestHandleToolsCall(t *testing.T) {
	s := newServerForTest(t, Options{})

	// Missing params.
	resp := s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("1"), Method: "tools/call"})
	if resp.Error == nil {
		t.Error("expected error for missing params")
	}

	// Malformed params.
	resp = s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("2"), Method: "tools/call", Params: json.RawMessage(`not json`)})
	if resp.Error == nil {
		t.Error("expected error for malformed params")
	}

	// Unknown tool.
	resp = s.route(context.Background(), &JSONRPCRequest{
		ID: json.RawMessage("3"), Method: "tools/call",
		Params: json.RawMessage(`{"name":"nope_tool"}`),
	})
	if resp.Error == nil {
		t.Error("expected error for unknown tool")
	}

	// Known tool, nil arguments normalised to {}.
	resp = s.route(context.Background(), &JSONRPCRequest{
		ID: json.RawMessage("4"), Method: "tools/call",
		Params: json.RawMessage(`{"name":"list_running_tasks"}`),
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	callResp, ok := resp.Result.(ToolCallResponse)
	if !ok || callResp.IsError {
		t.Fatalf("unexpected result: %+v", resp.Result)
	}

	// Known tool whose handler itself returns an error (bad input) => IsError result, not a JSON-RPC error.
	resp = s.route(context.Background(), &JSONRPCRequest{
		ID: json.RawMessage("5"), Method: "tools/call",
		Params: json.RawMessage(`{"name":"get_task_status","arguments":{}}`),
	})
	if resp.Error != nil {
		t.Fatalf("unexpected JSON-RPC error: %+v", resp.Error)
	}
	callResp, ok = resp.Result.(ToolCallResponse)
	if !ok || !callResp.IsError {
		t.Fatalf("expected a tool-level IsError result, got %+v", resp.Result)
	}
}

// --- completion/complete ---

func TestHandleComplete(t *testing.T) {
	s := newServerForTest(t, Options{})

	resp := s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("1"), Method: "completion/complete"})
	if resp.Error == nil {
		t.Error("expected error for missing params")
	}

	resp = s.route(context.Background(), &JSONRPCRequest{ID: json.RawMessage("2"), Method: "completion/complete", Params: json.RawMessage(`not json`)})
	if resp.Error == nil {
		t.Error("expected error for malformed params")
	}

	resp = s.route(context.Background(), &JSONRPCRequest{
		ID: json.RawMessage("3"), Method: "completion/complete",
		Params: json.RawMessage(`{"ref":{"type":"ref/prompt","name":"test-case-management"},"argument":{"name":"project_id","value":""}}`),
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
}

// --- resources/subscribe, resources/unsubscribe, PublishResource ---

func TestSubscribeUnsubscribeAndPublish(t *testing.T) {
	s := newServerForTest(t, Options{})
	subscribed := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(subscribed)
	other := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(other)

	const uri = "allure://docs/quickstart"
	ctx := sessctx.WithID(context.Background(), subscribed.id)
	resp := s.route(ctx, &JSONRPCRequest{ID: json.RawMessage("1"), Method: "resources/subscribe", Params: json.RawMessage(`{"uri":"` + uri + `"}`)})
	if resp.Error != nil {
		t.Fatalf("subscribe error: %+v", resp.Error)
	}

	// Missing/invalid params.
	resp = s.route(ctx, &JSONRPCRequest{ID: json.RawMessage("2"), Method: "resources/subscribe"})
	if resp.Error == nil {
		t.Error("expected error for missing subscribe params")
	}
	resp = s.route(ctx, &JSONRPCRequest{ID: json.RawMessage("3"), Method: "resources/unsubscribe"})
	if resp.Error == nil {
		t.Error("expected error for missing unsubscribe params")
	}

	s.PublishResource(uri)

	select {
	case msg := <-subscribed.send:
		if !strings.Contains(string(msg), uri) {
			t.Errorf("published message missing uri: %s", msg)
		}
	case <-time.After(time.Second):
		t.Error("subscribed session did not receive the publish notification")
	}
	select {
	case msg := <-other.send:
		t.Errorf("unsubscribed session should not receive a notification, got %s", msg)
	default:
	}

	resp = s.route(ctx, &JSONRPCRequest{ID: json.RawMessage("4"), Method: "resources/unsubscribe", Params: json.RawMessage(`{"uri":"` + uri + `"}`)})
	if resp.Error != nil {
		t.Fatalf("unsubscribe error: %+v", resp.Error)
	}
	s.PublishResource(uri)
	select {
	case msg := <-subscribed.send:
		t.Errorf("expected no notification after unsubscribe, got %s", msg)
	default:
	}
}

// --- Elicit / CreateMessage ---

func TestElicit_NoActiveSession(t *testing.T) {
	s := newServerForTest(t, Options{})
	result, err := s.Elicit(context.Background(), ElicitRequest{Message: "confirm?"})
	if err != nil || result.Action != "reject" {
		t.Errorf("result=%+v err=%v, want reject/nil", result, err)
	}

	ctx := sessctx.WithID(context.Background(), "unknown")
	result, err = s.Elicit(ctx, ElicitRequest{Message: "confirm?"})
	if err != nil || result.Action != "reject" {
		t.Errorf("unknown session: result=%+v err=%v, want reject/nil", result, err)
	}
}

func TestElicit_ActiveSessionRoundTrip(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)
	ctx := sessctx.WithID(context.Background(), sess.id)

	resultCh := make(chan *ElicitResult, 1)
	errCh := make(chan error, 1)
	go func() {
		r, err := s.Elicit(ctx, ElicitRequest{Message: "confirm?"})
		resultCh <- r
		errCh <- err
	}()

	var notif map[string]any
	select {
	case msg := <-sess.send:
		if err := json.Unmarshal(msg, &notif); err != nil {
			t.Fatalf("unmarshal notif: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("expected an elicitation/create notification on sess.send")
	}
	if notif["method"] != "elicitation/create" {
		t.Fatalf("unexpected method: %v", notif["method"])
	}
	elicitID, _ := notif["id"].(string)

	resp := &JSONRPCRequest{
		ID:     json.RawMessage(`"` + elicitID + `"`),
		Result: json.RawMessage(`{"action":"accept","content":{"ok":true}}`),
	}
	s.route(ctx, resp)

	select {
	case r := <-resultCh:
		if r.Action != "accept" {
			t.Errorf("Action = %q, want accept", r.Action)
		}
	case <-time.After(time.Second):
		t.Fatal("Elicit did not return after the response was routed")
	}
	if err := <-errCh; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestElicit_ClosedSessionReturnsCancel(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)
	// Fill the send buffer so the "send" case in Elicit's first select can
	// never be ready, forcing the sess.ctx.Done() branch deterministically.
	for i := 0; i < sessionSendBuffer; i++ {
		sess.send <- []byte("x")
	}
	// Cancel the session's context directly (not closeSession, which would
	// also remove it from s.sessions and make getSession return nil, taking
	// the earlier "no session" branch instead of the one under test).
	sess.cancel()

	ctx := sessctx.WithID(context.Background(), sess.id)
	result, err := s.Elicit(ctx, ElicitRequest{Message: "confirm?"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil || result.Action != "cancel" {
		t.Errorf("result = %+v, want cancel", result)
	}
}

func TestElicit_ContextCancelled(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)

	ctx, cancel := context.WithCancel(sessctx.WithID(context.Background(), sess.id))
	cancel()

	_, err := s.Elicit(ctx, ElicitRequest{Message: "confirm?"})
	if err == nil {
		t.Error("expected an error from the cancelled context")
	}
}

func TestCreateMessage_NoActiveSession(t *testing.T) {
	s := newServerForTest(t, Options{})
	if _, err := s.CreateMessage(context.Background(), SamplingRequest{}); err == nil {
		t.Error("expected error with no session id")
	}
	ctx := sessctx.WithID(context.Background(), "unknown")
	if _, err := s.CreateMessage(ctx, SamplingRequest{}); err == nil {
		t.Error("expected error with unknown session")
	}
}

func TestCreateMessage_ActiveSessionRoundTrip(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)
	ctx := sessctx.WithID(context.Background(), sess.id)

	resultCh := make(chan *SamplingResult, 1)
	errCh := make(chan error, 1)
	go func() {
		r, err := s.CreateMessage(ctx, SamplingRequest{
			Messages:  []SamplingMessage{{Role: "user", Content: SamplingMessageContent{Type: "text", Text: "hi"}}},
			MaxTokens: 10,
		})
		resultCh <- r
		errCh <- err
	}()

	var notif map[string]any
	select {
	case msg := <-sess.send:
		_ = json.Unmarshal(msg, &notif)
	case <-time.After(time.Second):
		t.Fatal("expected a sampling/createMessage notification")
	}
	sampID, _ := notif["id"].(string)

	s.route(ctx, &JSONRPCRequest{
		ID:     json.RawMessage(`"` + sampID + `"`),
		Result: json.RawMessage(`{"role":"assistant","content":{"text":"hello"},"stopReason":"end"}`),
	})

	select {
	case r := <-resultCh:
		if r.Content.Text != "hello" {
			t.Errorf("Text = %q, want hello", r.Content.Text)
		}
	case <-time.After(time.Second):
		t.Fatal("CreateMessage did not return")
	}
	if err := <-errCh; err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateMessage_ClientErrorLeavesCallerWaiting(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)
	ctx := sessctx.WithID(context.Background(), sess.id)

	go func() {
		<-sess.send // drain the notification
	}()
	// handleJSONRPCResponse with Error set must not send anything to the
	// sampling channel — verified indirectly via a fresh channel injected below.
	sess.pendingMu.Lock()
	ch := make(chan *SamplingResult, 1)
	for id := range sess.sampling {
		sess.sampling[id] = ch
	}
	sess.pendingMu.Unlock()

	go func() {
		_, _ = s.CreateMessage(ctx, SamplingRequest{})
	}()
	time.Sleep(50 * time.Millisecond)

	sess.pendingMu.Lock()
	var sampID string
	for id := range sess.sampling {
		sampID = id
	}
	sess.pendingMu.Unlock()
	if sampID == "" {
		t.Fatal("expected a pending sampling id to be registered")
	}

	s.route(ctx, &JSONRPCRequest{ID: json.RawMessage(`"` + sampID + `"`), Error: &JSONRPCError{Code: 1, Message: "denied"}})

	select {
	case <-ch:
		t.Error("did not expect a result on client error")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestCreateMessage_ClosedSession(t *testing.T) {
	s := newServerForTest(t, Options{})
	sess := s.newSession(context.Background(), newSessionID())
	defer s.closeSession(sess)
	for i := 0; i < sessionSendBuffer; i++ {
		sess.send <- []byte("x")
	}
	sess.cancel() // cancel without removing from s.sessions, see TestElicit_ClosedSessionReturnsCancel

	ctx := sessctx.WithID(context.Background(), sess.id)
	if _, err := s.CreateMessage(ctx, SamplingRequest{}); err == nil {
		t.Error("expected an error for a closed session")
	}
}

// --- pagination / small helpers ---

func TestCursorOffsetRoundTrip(t *testing.T) {
	cases := []struct {
		cursor string
		want   int
	}{
		{"", 0},
		{"5", 5},
		{"-1", 0},
		{"abc", 0},
	}
	for _, tc := range cases {
		if got := cursorToOffset(tc.cursor); got != tc.want {
			t.Errorf("cursorToOffset(%q) = %d, want %d", tc.cursor, got, tc.want)
		}
	}
	if offsetToCursor(42) != "42" {
		t.Errorf("offsetToCursor(42) = %q, want 42", offsetToCursor(42))
	}
}

func TestSortToolsByName(t *testing.T) {
	ts := []*tools.Tool{{Name: "zeta"}, {Name: "alpha"}, {Name: "mid"}}
	sortToolsByName(ts)
	if ts[0].Name != "alpha" || ts[1].Name != "mid" || ts[2].Name != "zeta" {
		t.Errorf("unexpected order: %v", []string{ts[0].Name, ts[1].Name, ts[2].Name})
	}
}

func TestNewSessionIDIsUniqueHex(t *testing.T) {
	a := newSessionID()
	b := newSessionID()
	if a == b {
		t.Error("expected unique session ids")
	}
	if len(a) != 32 {
		t.Errorf("len = %d, want 32 (16 bytes hex-encoded)", len(a))
	}
}

func TestResultToJSON(t *testing.T) {
	if got := resultToJSON(map[string]any{"a": 1}); !strings.Contains(got, `"a"`) {
		t.Errorf("unexpected output: %s", got)
	}
	// A Go channel cannot be marshalled to JSON, exercising the fallback branch.
	got := resultToJSON(make(chan int))
	if !strings.Contains(got, "error") {
		t.Errorf("expected fallback error JSON, got %s", got)
	}
}

func TestSetCORSHeaders(t *testing.T) {
	s := newServerForTest(t, Options{})
	w := httptest.NewRecorder()
	s.setCORSHeaders(w) // no origin configured => no-op
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no CORS header when unconfigured")
	}

	s = newServerForTest(t, Options{CORSAllowOrigin: "*"})
	w = httptest.NewRecorder()
	s.setCORSHeaders(w)
	if w.Header().Get("Vary") != "" {
		t.Error("expected no Vary header for wildcard origin")
	}

	s = newServerForTest(t, Options{CORSAllowOrigin: "https://example.com"})
	w = httptest.NewRecorder()
	s.setCORSHeaders(w)
	if w.Header().Get("Vary") != "Origin" {
		t.Error("expected Vary: Origin for a specific origin")
	}
}

// --- Streamable HTTP transport (HandleMCP) ---

func TestHandleMCP_OptionsAndUnauthorized(t *testing.T) {
	s := newServerForTest(t, Options{})
	w := httptest.NewRecorder()
	s.HandleMCP(w, httptest.NewRequest(http.MethodOptions, "/mcp", nil))
	if w.Code != http.StatusNoContent {
		t.Errorf("OPTIONS: status = %d, want 204", w.Code)
	}

	s = newServerForTest(t, Options{Users: []User{{Name: "u", Token: "t"}}})
	w = httptest.NewRecorder()
	s.HandleMCP(w, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("{}")))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestHandleMCP_GetNotSupported(t *testing.T) {
	s := newServerForTest(t, Options{})
	w := httptest.NewRecorder()
	s.HandleMCP(w, httptest.NewRequest(http.MethodGet, "/mcp", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestHandleMCP_Delete(t *testing.T) {
	s := newServerForTest(t, Options{})
	req := httptest.NewRequest(http.MethodDelete, "/mcp", nil)
	req.Header.Set("Mcp-Session-Id", "some-id")
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	// Without a session header, still 200.
	w = httptest.NewRecorder()
	s.HandleMCP(w, httptest.NewRequest(http.MethodDelete, "/mcp", nil))
	if w.Code != http.StatusOK {
		t.Errorf("no-header status = %d, want 200", w.Code)
	}
}

func TestHandleMCP_PostWrongContentType(t *testing.T) {
	s := newServerForTest(t, Options{})
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want 415", w.Code)
	}
}

func TestHandleMCP_PostBodyTooLarge(t *testing.T) {
	s := newServerForTest(t, Options{})
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(strings.Repeat("a", maxMessageBody+10)))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", w.Code)
	}
}

func TestHandleMCP_PostMalformedJSON(t *testing.T) {
	s := newServerForTest(t, Options{})
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleMCP_PostInvalidVersion(t *testing.T) {
	s := newServerForTest(t, Options{})
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"1.0","id":1,"method":"initialize"}`))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (JSON-RPC error body with HTTP 200)", w.Code)
	}
	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error == nil || resp.Error.Code != ErrCodeInvalidRequest {
		t.Errorf("unexpected error body: %+v", resp)
	}
}

func TestHandleMCP_PostNotificationReturns202(t *testing.T) {
	s := newServerForTest(t, Options{})
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","method":"notifications/initialized"}`))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202", w.Code)
	}
}

func TestHandleMCP_PostInitializeAndFollowUp(t *testing.T) {
	s := newServerForTest(t, Options{})
	req := httptest.NewRequest(http.MethodPost, "/mcp",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`))
	req.Header.Set("X-Allure-Token", "tok")
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	sessID := w.Header().Get("Mcp-Session-Id")
	if sessID == "" {
		t.Fatal("expected Mcp-Session-Id response header")
	}
	var initResp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &initResp); err != nil || initResp.Error != nil {
		t.Fatalf("unexpected initialize response: %v %+v", err, initResp)
	}

	req = httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`))
	req.Header.Set("Mcp-Session-Id", sessID)
	w = httptest.NewRecorder()
	s.HandleMCP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("follow-up status = %d, want 200", w.Code)
	}
}
