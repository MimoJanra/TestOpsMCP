package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
	"github.com/MimoJanra/TestOpsMCP/internal/session"
)

// newTestRegistryWithServer wires a Registry to an httptest server so handler
// tests exercise the real allure.Client HTTP path instead of only input
// validation (which registry_test.go already covers). Requests to the JWT
// endpoint are answered transparently so callers only need to handle the
// actual API call(s) their test cares about.
func newTestRegistryWithServer(t *testing.T, handler http.HandlerFunc) *Registry {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/uaa/oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"test-jwt","expires_in":3600}`))
	})
	mux.HandleFunc("/", handler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := allure.NewClient(server.URL, "test-token", 5*time.Second)
	return NewRegistry(client, core.NewLogger(core.LevelError))
}

func jsonHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestListTestCases_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{
		"content":[{"id":1,"name":"tc1","projectId":5,"status":"ACTIVE","automationStatus":"AUTOMATED"}],
		"empty":false,"last":true,"number":0,"size":10,"totalElements":1
	}`))

	res, err := r.listTestCases(context.Background(), listTestCasesArgs{ProjectID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := res.(map[string]any)
	if m["total"] != 1 {
		t.Errorf("total = %v, want 1", m["total"])
	}
}

func TestListTestCases_ClampsSize(t *testing.T) {
	var gotQuery string
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotQuery = req.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[],"empty":true,"last":true,"number":0,"size":100,"totalElements":0}`))
	})

	if _, err := r.listTestCases(context.Background(), listTestCasesArgs{ProjectID: 1, Size: 9999}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery == "" {
		t.Fatal("expected a query string to be sent")
	}
}

func TestListTestCases_ServerError(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusInternalServerError, `{"message":"boom"}`))
	if _, err := r.listTestCases(context.Background(), listTestCasesArgs{ProjectID: 1}); err == nil {
		t.Fatal("expected error on 500")
	}
}

func TestGetTestCase_Handler(t *testing.T) {
	call := 0
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		call++
		w.Header().Set("Content-Type", "application/json")
		if call == 1 {
			_, _ = w.Write([]byte(`{"id":1,"name":"tc1"}`))
		} else {
			_, _ = w.Write([]byte(`{"steps":[{"body":"do it"}]}`))
		}
	})

	res, err := r.getTestCase(context.Background(), getTestCaseArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := res.(map[string]any)
	if _, ok := m["manual_scenario"]; !ok {
		t.Error("expected manual_scenario to be merged in when scenario fetch succeeds")
	}
}

func TestGetTestCase_ScenarioFetchFailsGracefully(t *testing.T) {
	call := 0
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		call++
		if call == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":1,"name":"tc1"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	res, err := r.getTestCase(context.Background(), getTestCaseArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("expected overview to still return despite scenario 404: %v", err)
	}
	m := res.(map[string]any)
	if _, ok := m["manual_scenario"]; ok {
		t.Error("did not expect manual_scenario when scenario fetch failed")
	}
}

func TestRunTestCase_Handler(t *testing.T) {
	var gotMethod string
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotMethod = req.Method
		w.WriteHeader(http.StatusOK)
	})

	res, err := r.runTestCase(context.Background(), runTestCaseArgs{TestCaseID: 1, LaunchID: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "started" {
		t.Errorf("unexpected result: %v", res)
	}
	if gotMethod == "" {
		t.Error("expected a request to reach the server")
	}
}

func TestCreateTestCase_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{
		"id":10,"uuid":"abc","name":"n","projectId":5,"description":"d","status":"ACTIVE","automationStatus":"MANUAL","fullName":"n"
	}`))

	res, err := r.createTestCase(context.Background(), createTestCaseArgs{ProjectID: 5, Name: "n"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := res.(map[string]any)
	if m["id"] != int64(10) {
		t.Errorf("id = %v, want 10", m["id"])
	}
}

func TestUpdateTestCase_Handler(t *testing.T) {
	var gotBody map[string]any
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		_ = json.NewDecoder(req.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	})

	res, err := r.updateTestCase(context.Background(), updateTestCaseArgs{TestCaseID: 1, Name: "new name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "updated" {
		t.Errorf("unexpected result: %v", res)
	}
	if gotBody["name"] != "new name" {
		t.Errorf("request body name = %v, want %q", gotBody["name"], "new name")
	}
}

func TestUpdateTestCase_NoFieldsErrors(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.updateTestCase(context.Background(), updateTestCaseArgs{TestCaseID: 1}); err == nil {
		t.Fatal("expected error when no fields are set")
	}
}

func TestDeleteTestCase_NoElicitContext(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.deleteTestCase(context.Background(), deleteTestCaseArgs{TestCaseID: 1}); err == nil {
		t.Fatal("expected error without an interactive session")
	}
}

func TestDeleteTestCase_Cancelled(t *testing.T) {
	r := newTestRegistry(t)
	ctx := session.WithElicit(context.Background(), func(ctx context.Context, message string, schema []byte) (*session.ElicitResult, error) {
		return &session.ElicitResult{Action: "reject"}, nil
	})

	res, err := r.deleteTestCase(ctx, deleteTestCaseArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["cancelled"] != true {
		t.Errorf("expected cancelled=true, got %v", res)
	}
}

func TestDeleteTestCase_ConfirmedDeletes(t *testing.T) {
	var gotMethod string
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotMethod = req.Method
		w.WriteHeader(http.StatusOK)
	})
	ctx := session.WithElicit(context.Background(), func(ctx context.Context, message string, schema []byte) (*session.ElicitResult, error) {
		return &session.ElicitResult{Action: "accept"}, nil
	})

	res, err := r.deleteTestCase(ctx, deleteTestCaseArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "deleted" {
		t.Errorf("unexpected result: %v", res)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
}

func TestCloneTestCase_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"id":99}`))
	res, err := r.cloneTestCase(context.Background(), cloneTestCaseArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["cloned_test_case_id"] != int64(99) {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestRestoreTestCase_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.restoreTestCase(context.Background(), restoreTestCaseArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "restored" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestCreateTestCaseStep_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"createdStepId":7}`))
	res, err := r.createTestCaseStep(context.Background(), createTestCaseStepArgs{TestCaseID: 1, Body: "step"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["step_id"] != int64(7) {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestCreateTestCaseStep_ValidatesBody(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.createTestCaseStep(context.Background(), createTestCaseStepArgs{TestCaseID: 1}); err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestUpdateTestCaseStep_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.updateTestCaseStep(context.Background(), updateTestCaseStepArgs{StepID: 1, Body: "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "updated" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestUpdateTestCaseStep_RequiresAField(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.updateTestCaseStep(context.Background(), updateTestCaseStepArgs{StepID: 1}); err == nil {
		t.Fatal("expected error when neither body nor expected_result is set")
	}
}

func TestDeleteTestCaseStep_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.deleteTestCaseStep(context.Background(), deleteTestCaseStepArgs{StepID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "deleted" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestGetTestCaseCustomFields_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `[
		{"customField":{"id":1,"name":"Priority"},"values":[{"id":10,"name":"High"}]}
	]`))

	res, err := r.getTestCaseCustomFields(context.Background(), getTestCaseCustomFieldsArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fields := res.(map[string]any)["custom_fields"].([]map[string]any)
	if len(fields) != 1 || fields[0]["custom_field_name"] != "Priority" {
		t.Errorf("unexpected fields: %+v", fields)
	}
}

func TestUpdateTestCaseCustomFields_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	args := updateTestCaseCustomFieldsArgs{TestCaseID: 1}
	args.CustomFields = append(args.CustomFields, struct {
		CustomFieldID int64   `json:"custom_field_id"`
		ValueIDs      []int64 `json:"value_ids"`
	}{CustomFieldID: 1, ValueIDs: []int64{10}})

	res, err := r.updateTestCaseCustomFields(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "updated" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestUpdateTestCaseCustomFields_RequiresEntries(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.updateTestCaseCustomFields(context.Background(), updateTestCaseCustomFieldsArgs{TestCaseID: 1}); err == nil {
		t.Fatal("expected error for empty custom_fields")
	}
}

func TestGetTestCaseHistory_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"content":[]}`))
	res, err := r.getTestCaseHistory(context.Background(), getTestCaseHistoryArgs{TestCaseID: 1, Size: 500})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Error("expected a non-nil result")
	}
}
