package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
)

// newOpIndexRegistry builds a minimal Registry with a synthetic operations
// index, bypassing NewRegistry's disk-based spec discovery (which doesn't
// find spec/testops.json when the test binary's cwd is internal/tools).
func newOpIndexRegistry(t *testing.T, allureClient *allure.Client) *Registry {
	t.Helper()
	spec := &OpenAPISpec{
		Paths: map[string]map[string]interface{}{
			"/api/thing/{id}": {
				"get": map[string]interface{}{
					"operationId": "get_thing",
					"summary":     "Get a thing",
					"tags":        []interface{}{"things"},
					"parameters": []interface{}{
						map[string]interface{}{"name": "id", "in": "path", "required": true},
					},
				},
			},
		},
	}
	idx, err := BuildOperationsIndex(spec)
	if err != nil {
		t.Fatal(err)
	}
	return &Registry{
		tools:         make(map[string]*Tool),
		allure:        allureClient,
		logger:        core.NewLogger(core.LevelError),
		opIndex:       idx,
		sessionTokens: make(map[string]string),
		resources:     make(map[string]*Resource),
		prompts:       make(map[string]*RegistryPrompt),
	}
}

// --- session token management -----------------------------------------------

func TestSessionTokenLifecycle(t *testing.T) {
	r := newTestRegistry(t)

	if got := r.getSessionToken("sess1"); got != "" {
		t.Fatalf("expected empty token before Set, got %q", got)
	}

	r.SetSessionToken("sess1", "tok-abc")
	if got := r.getSessionToken("sess1"); got != "tok-abc" {
		t.Errorf("getSessionToken = %q, want tok-abc", got)
	}

	// A second session's token is independent.
	r.SetSessionToken("sess2", "tok-xyz")
	if got := r.getSessionToken("sess1"); got != "tok-abc" {
		t.Errorf("sess1 token clobbered by sess2: got %q", got)
	}

	r.ClearSessionToken("sess1")
	if got := r.getSessionToken("sess1"); got != "" {
		t.Errorf("expected empty token after Clear, got %q", got)
	}
	if got := r.getSessionToken("sess2"); got != "tok-xyz" {
		t.Errorf("ClearSessionToken(sess1) affected sess2: got %q", got)
	}
}

func TestConfigureAllureToken(t *testing.T) {
	r := newTestRegistry(t)

	if _, err := r.configureAllureToken(context.Background(), configureAllureTokenArgs{Token: ""}); err == nil {
		t.Error("expected error for empty token")
	}

	result, err := r.configureAllureToken(context.Background(), configureAllureTokenArgs{Token: "my-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok || m["status"] != "configured" {
		t.Errorf("unexpected result: %#v", result)
	}

	// No session ID in context falls back to the stdio session.
	if got := r.getSessionToken("stdio"); got != "my-token" {
		t.Errorf("expected token stored under stdio session, got %q (all=%v)", got, r.sessionTokens)
	}
}

// --- resources ---------------------------------------------------------------

func TestResourcesRegistryDefaults(t *testing.T) {
	r := newTestRegistry(t)

	list := r.ListResources()
	if len(list) == 0 {
		t.Fatal("expected at least the quickstart resource to be registered")
	}

	res := r.GetResource("allure://docs/quickstart")
	if res == nil {
		t.Fatal("expected quickstart resource to be registered")
	}
	if res.MimeType != "text/markdown" {
		t.Errorf("quickstart MimeType = %q", res.MimeType)
	}
	if html := res.GetHTML(); html == "" {
		t.Error("quickstart GetHTML() returned empty content")
	}

	if r.GetResource("ui://does-not-exist") != nil {
		t.Error("expected nil for unregistered resource URI")
	}
}

func TestSetPublishResourceAndOnSubscribe(t *testing.T) {
	r := newTestRegistry(t)

	var published []string
	r.SetPublishResource(func(uri string) { published = append(published, uri) })
	if r.publishResource == nil {
		t.Fatal("expected publishResource callback to be set")
	}

	// A non-launch-dashboard URI is a no-op: no watcher goroutine started.
	r.OnSubscribe(context.Background(), "ui://widgets/action-picker")

	// A launch-dashboard URI starts a background watcher; cancel immediately
	// so the test doesn't wait out the 10s poll ticker.
	ctx, cancel := context.WithCancel(context.Background())
	r.OnSubscribe(ctx, launchDashboardURIFor(42))
	cancel()
}

// --- prompts -------------------------------------------------------------

func TestPromptsRegistryDefaults(t *testing.T) {
	r := newTestRegistry(t)

	list := r.ListPrompts()
	if len(list) == 0 {
		t.Fatal("expected default prompts to be registered")
	}

	messages, desc, err := r.GetPrompt("analyze-test-failures", map[string]string{"launch_id": "123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc == "" {
		t.Error("expected non-empty prompt description")
	}
	if len(messages) == 0 {
		t.Error("expected at least one prompt message")
	}

	// nil args map must not panic; GetPrompt substitutes an empty map.
	if _, _, err := r.GetPrompt("launch-report-summary", nil); err != nil {
		t.Errorf("unexpected error with nil args: %v", err)
	}

	if _, _, err := r.GetPrompt("no-such-prompt", nil); err == nil {
		t.Error("expected error for unknown prompt name")
	}
}

// --- completion ------------------------------------------------------------

func TestComplete_UnknownRefTypeOrArg(t *testing.T) {
	r := newTestRegistry(t)
	if got := r.Complete(context.Background(), "ref/resource", "p", "project_id", ""); got != nil {
		t.Errorf("expected nil for non-prompt ref type, got %v", got)
	}
	if got := r.Complete(context.Background(), "ref/prompt", "p", "unknown_arg", ""); got != nil {
		t.Errorf("expected nil for unrecognized argument name, got %v", got)
	}
}

func TestComplete_NoAllureClientYieldsEmpty(t *testing.T) {
	r := newTestRegistry(t) // r.allure == nil
	if got := r.completeProjectIDs(context.Background(), ""); len(got) != 0 {
		t.Errorf("expected no project IDs without an Allure client, got %v", got)
	}
	if got := r.completeLaunchIDs(context.Background(), ""); len(got) != 0 {
		t.Errorf("expected no launch IDs without an Allure client, got %v", got)
	}
}

func TestWarmCompletionCache_FetchesOnceWithinTTL(t *testing.T) {
	var projectCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		projectCalls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"content":[{"id":1,"name":"Proj"}],"totalElements":1}`))
	}))
	defer server.Close()

	client := allure.NewClient(server.URL, "tok", 5*time.Second)
	r := newTestRegistry(t)
	r.allure = client

	ids := r.completeProjectIDs(context.Background(), "")
	if len(ids) != 1 || ids[0] != "1" {
		t.Fatalf("expected [1], got %v", ids)
	}
	firstCallCount := projectCalls

	// Second call within the TTL window must hit the cache, not the server again.
	ids2 := r.completeProjectIDs(context.Background(), "1")
	if len(ids2) != 1 || ids2[0] != "1" {
		t.Fatalf("expected [1] from cache with prefix filter, got %v", ids2)
	}
	if projectCalls != firstCallCount {
		t.Errorf("expected cached warmCompletionCache to avoid a second fetch; calls went from %d to %d", firstCallCount, projectCalls)
	}

	// A prefix that matches nothing filters everything out.
	if ids3 := r.completeProjectIDs(context.Background(), "9"); len(ids3) != 0 {
		t.Errorf("expected no matches for prefix 9, got %v", ids3)
	}
}

// --- search/execute handlers -------------------------------------------------

func TestBuildSearchResults_LimitHandling(t *testing.T) {
	ops := []*Operation{
		{OperationID: "op1", Path: "/a", Method: "GET"},
		{OperationID: "op2", Path: "/b", Method: "POST"},
		{OperationID: "op3", Path: "/c", Method: "DELETE"},
	}

	if got := buildSearchResults(ops, 0); len(got) != 3 {
		t.Errorf("limit=0 should default to 10 (capped at len(ops)=3), got %d", len(got))
	}
	if got := buildSearchResults(ops, 500); len(got) != 3 {
		t.Errorf("limit=500 should default to 10 then cap at len(ops)=3, got %d", len(got))
	}
	if got := buildSearchResults(ops, 2); len(got) != 2 {
		t.Errorf("limit=2 should return exactly 2, got %d", len(got))
	}
	if got := buildSearchResults(nil, 5); len(got) != 0 {
		t.Errorf("empty ops should return 0 results, got %d", len(got))
	}

	got := buildSearchResults(ops, 1)
	if got[0].OperationID != "op1" || got[0].Path != "/a" || got[0].Method != "GET" {
		t.Errorf("field mapping wrong: %+v", got[0])
	}
}

func TestSearchTestOpsOperationsHandler(t *testing.T) {
	r := newTestRegistry(t) // opIndex nil
	if _, err := r.searchTestOpsOperations(context.Background(), SearchRequest{Intent: ""}); err == nil {
		t.Error("expected error for empty intent")
	}
	if _, err := r.searchTestOpsOperations(context.Background(), SearchRequest{Intent: "launch"}); err == nil {
		t.Error("expected error when opIndex is unavailable")
	}

	r2 := newOpIndexRegistry(t, nil)
	result, err := r2.searchTestOpsOperations(context.Background(), SearchRequest{Intent: "thing", Limit: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type: %T", result)
	}
	if m["total"] != 1 {
		t.Errorf("total = %v, want 1", m["total"])
	}
}

func TestExecuteTestOpsOperationHandler(t *testing.T) {
	r := newTestRegistry(t) // opIndex nil
	if _, err := r.executeTestOpsOperation(context.Background(), ExecuteRequest{OperationID: ""}); err == nil {
		t.Error("expected error for empty operation_id")
	}
	if _, err := r.executeTestOpsOperation(context.Background(), ExecuteRequest{OperationID: "get_thing"}); err == nil {
		t.Error("expected error when opIndex is unavailable")
	}

	r2 := newOpIndexRegistry(t, nil)
	if _, err := r2.executeTestOpsOperation(context.Background(), ExecuteRequest{OperationID: "no_such_op"}); err == nil {
		t.Error("expected error for unknown operation id")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	r3 := newOpIndexRegistry(t, allure.NewClient(server.URL, "tok", 5*time.Second))
	result, err := r3.executeTestOpsOperation(context.Background(), ExecuteRequest{
		OperationID: "get_thing",
		Parameters:  map[string]interface{}{"id": 7},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected a non-nil result from the fake server")
	}
}
