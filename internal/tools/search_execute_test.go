package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
)

// TestExecuteOperation_DestructiveRequiresConfirm guards against
// execute_testops_operation running an arbitrary DELETE/PUT/PATCH from the
// OpenAPI spec with no server-side check at all (the only prior signal was
// the advisory destructiveHint annotation, which no client is required to
// honor).
func TestExecuteOperation_DestructiveRequiresConfirm(t *testing.T) {
	r := newTestRegistry(t) // r.allure is nil; the guard must fire before it's touched.

	for _, method := range []string{http.MethodDelete, http.MethodPut, http.MethodPatch} {
		op := &Operation{OperationID: "test_op", Path: "/api/thing/{id}", Method: method}

		_, err := r.executeOperation(context.Background(), op, map[string]interface{}{"id": 1})
		if err == nil {
			t.Errorf("%s: expected error without confirm, got nil", method)
		} else if !strings.Contains(err.Error(), "confirm") {
			t.Errorf("%s: error %q does not mention confirm", method, err)
		}

		_, err = r.executeOperation(context.Background(), op, map[string]interface{}{"id": 1, "confirm": false})
		if err == nil {
			t.Errorf("%s: expected error with confirm=false, got nil", method)
		}
	}
}

// TestExecuteOperation_ConfirmedDestructiveProceeds ensures the guard doesn't
// block legitimate destructive calls once the caller has acknowledged them,
// and that "confirm" is stripped before parameters are matched against the
// spec (it must never leak into path/query/body).
func TestExecuteOperation_ConfirmedDestructiveProceeds(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		gotMethod = req.Method
		gotPath = req.URL.Path
		buf := make([]byte, req.ContentLength)
		_, _ = req.Body.Read(buf)
		gotBody = string(buf)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := allure.NewClient(server.URL, "test-token", 5*time.Second)
	r := NewRegistry(client, core.NewLogger(core.LevelError))

	op := &Operation{
		OperationID: "delete_thing",
		Path:        "/api/thing/{id}",
		Method:      http.MethodDelete,
		Parameters: []OperationParameter{
			{Name: "id", In: "path"},
		},
	}

	result, err := r.executeOperation(context.Background(), op, map[string]interface{}{"id": 42, "confirm": true})
	if err != nil {
		t.Fatalf("expected confirmed delete to proceed, got error: %v", err)
	}
	if result == nil {
		t.Error("expected a result from the fake server")
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/api/thing/42" {
		t.Errorf("path = %q, want /api/thing/42 (id path param not substituted)", gotPath)
	}
	if strings.Contains(gotBody, "confirm") {
		t.Errorf("body %q leaked the confirm flag to the upstream request", gotBody)
	}
}

// TestExecuteOperation_NonDestructiveIgnoresConfirm ensures GET/POST are
// unaffected by the guard — it only gates DELETE/PUT/PATCH.
func TestExecuteOperation_NonDestructiveIgnoresConfirm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := allure.NewClient(server.URL, "test-token", 5*time.Second)
	r := NewRegistry(client, core.NewLogger(core.LevelError))

	op := &Operation{OperationID: "list_things", Path: "/api/thing", Method: http.MethodGet}

	if _, err := r.executeOperation(context.Background(), op, nil); err != nil {
		t.Fatalf("expected GET without confirm to proceed, got error: %v", err)
	}
}
