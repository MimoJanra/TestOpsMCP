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
	"github.com/MimoJanra/TestOpsMCP/internal/tasks"
)

// newBulkTestRegistry builds a Registry backed by a fake Allure HTTP server
// running the given handler, for exercising bulk_* handlers end to end.
func newBulkTestRegistry(t *testing.T, handler http.HandlerFunc) *Registry {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := allure.NewClient(server.URL, "test-token", 5*time.Second)
	return NewRegistry(client, core.NewLogger(core.LevelError))
}

func bulkOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{}`))
}

func bulkErrHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{"message":"boom"}`))
}

// waitForTask polls the task store until the task leaves StatusWorking or the
// timeout elapses.
func waitForTask(t *testing.T, r *Registry, taskID string) *tasks.Task {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		task, ok := r.taskStore.Get(taskID)
		if !ok {
			t.Fatalf("task %s not found", taskID)
		}
		if task.Status != tasks.StatusWorking {
			return task
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("task %s did not finish in time", taskID)
	return nil
}

func TestBulkHandlers_HappyAndErrorPaths(t *testing.T) {
	cases := []struct {
		name string
		run  func(r *Registry) (any, error)
	}{
		{"bulkAddTestCaseMembersTool", func(r *Registry) (any, error) {
			return r.bulkAddTestCaseMembersTool(context.Background(), bulkAddTestCaseMembersArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, Members: []allure.MemberDto{{ID: 1}},
			})
		}},
		{"bulkRemoveTestCaseMembersTool", func(r *Registry) (any, error) {
			return r.bulkRemoveTestCaseMembersTool(context.Background(), bulkRemoveTestCaseMembersArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, MemberIDs: []int64{1},
			})
		}},
		{"bulkAddTestCaseCustomFields", func(r *Registry) (any, error) {
			return r.bulkAddTestCaseCustomFields(context.Background(), bulkAddTestCaseCustomFieldsArgs{
				ProjectID: 1, TestCaseIDs: []int64{1},
				CustomFields: []struct {
					CustomFieldID int64   `json:"custom_field_id"`
					ValueIDs      []int64 `json:"value_ids"`
				}{{CustomFieldID: 5, ValueIDs: []int64{12}}},
			})
		}},
		{"bulkRemoveTestCaseCustomFields", func(r *Registry) (any, error) {
			return r.bulkRemoveTestCaseCustomFields(context.Background(), bulkRemoveTestCaseCustomFieldsArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, CustomFieldIDs: []int64{5},
			})
		}},
		{"bulkAddTestCaseExternalLinks", func(r *Registry) (any, error) {
			return r.bulkAddTestCaseExternalLinks(context.Background(), bulkAddTestCaseExternalLinksArgs{
				ProjectID: 1, TestCaseIDs: []int64{1},
				Links: []struct {
					Name string `json:"name"`
					Type string `json:"type"`
					URL  string `json:"url"`
				}{{Name: "n", Type: "t", URL: "http://example.com"}},
			})
		}},
		{"bulkAddTestCaseIssues", func(r *Registry) (any, error) {
			return r.bulkAddTestCaseIssues(context.Background(), bulkAddTestCaseIssuesArgs{
				ProjectID: 1, TestCaseIDs: []int64{1},
				Issues: []struct {
					ID            int64  `json:"id"`
					DisplayName   string `json:"display_name"`
					URL           string `json:"url"`
					IntegrationID int64  `json:"integration_id"`
				}{{ID: 1, DisplayName: "BUG-1"}},
			})
		}},
		{"bulkRemoveTestCaseIssues", func(r *Registry) (any, error) {
			return r.bulkRemoveTestCaseIssues(context.Background(), bulkRemoveTestCaseIssuesArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, IssueIDs: []int64{1},
			})
		}},
		{"bulkSetTestCaseLayer", func(r *Registry) (any, error) {
			return r.bulkSetTestCaseLayer(context.Background(), bulkSetTestCaseLayerArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, LayerID: 2,
			})
		}},
		{"bulkMoveTestCases", func(r *Registry) (any, error) {
			return r.bulkMoveTestCases(context.Background(), bulkMoveTestCasesArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, ToProjectID: 2,
			})
		}},
		{"bulkCreateTestPlan", func(r *Registry) (any, error) {
			return r.bulkCreateTestPlan(context.Background(), bulkCreateTestPlanArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, TestPlanName: "plan",
			})
		}},
		{"bulkMuteTestCases", func(r *Registry) (any, error) {
			return r.bulkMuteTestCases(context.Background(), bulkMuteTestCasesArgs{
				ProjectID: 1, TestCaseIDs: []int64{1},
			})
		}},
		{"bulkSetTestCaseStatus", func(r *Registry) (any, error) {
			return r.bulkSetTestCaseStatus(context.Background(), bulkSetTestCaseStatusArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, StatusID: 2, WorkflowID: 3,
			})
		}},
		{"bulkAddTestCaseTags", func(r *Registry) (any, error) {
			return r.bulkAddTestCaseTags(context.Background(), bulkAddTestCaseTagsArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, Tags: []allure.TestTagDto{{Name: "smoke"}},
			})
		}},
		{"bulkRemoveTestCaseTags", func(r *Registry) (any, error) {
			return r.bulkRemoveTestCaseTags(context.Background(), bulkRemoveTestCaseTagsArgs{
				ProjectID: 1, TestCaseIDs: []int64{1}, Tags: []allure.TestTagDto{{Name: "smoke"}},
			})
		}},
		{"bulkAssignTestResults", func(r *Registry) (any, error) {
			return r.bulkAssignTestResults(context.Background(), bulkAssignTestResultsArgs{
				LaunchID: 1, TestResultIDs: []int64{1}, Assignees: []string{"alice"},
			})
		}},
		{"bulkMuteTestResults", func(r *Registry) (any, error) {
			return r.bulkMuteTestResults(context.Background(), bulkMuteTestResultsArgs{
				LaunchID: 1, TestResultIDs: []int64{1}, Reason: "flaky",
			})
		}},
		{"bulkUnmuteTestResults", func(r *Registry) (any, error) {
			return r.bulkUnmuteTestResults(context.Background(), bulkUnmuteTestResultsArgs{
				LaunchID: 1, TestResultIDs: []int64{1},
			})
		}},
		{"bulkResolveTestResults", func(r *Registry) (any, error) {
			return r.bulkResolveTestResults(context.Background(), bulkResolveTestResultsArgs{
				LaunchID: 1, TestResultIDs: []int64{1}, Status: "failed",
			})
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/ok", func(t *testing.T) {
			r := newBulkTestRegistry(t, bulkOKHandler)
			result, err := tc.run(r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result on success")
			}
		})

		t.Run(tc.name+"/error", func(t *testing.T) {
			r := newBulkTestRegistry(t, bulkErrHandler)
			if _, err := tc.run(r); err == nil {
				t.Error("expected error when upstream fails, got nil")
			}
		})
	}
}

func TestBulkDeleteTestCases_RequiresElicitation(t *testing.T) {
	r := newBulkTestRegistry(t, bulkOKHandler)
	args := bulkDeleteTestCasesArgs{ProjectID: 1, TestCaseIDs: []int64{1, 2}}

	// No elicitation function in context: must refuse rather than delete silently.
	if _, err := r.bulkDeleteTestCases(context.Background(), args); err == nil {
		t.Error("expected error when no interactive session is available")
	}

	// Elicitation rejects: handler must report cancellation, not call the API.
	rejectCtx := session.WithElicit(context.Background(), func(ctx context.Context, msg string, schema []byte) (*session.ElicitResult, error) {
		return &session.ElicitResult{Action: "reject"}, nil
	})
	result, err := r.bulkDeleteTestCases(rejectCtx, args)
	if err != nil {
		t.Fatalf("unexpected error on reject: %v", err)
	}
	m, ok := result.(map[string]any)
	if !ok || m["cancelled"] != true {
		t.Errorf("expected cancelled result on reject, got %#v", result)
	}

	// Elicitation accepts: handler proceeds to call the (fake) API.
	acceptCtx := session.WithElicit(context.Background(), func(ctx context.Context, msg string, schema []byte) (*session.ElicitResult, error) {
		return &session.ElicitResult{Action: "accept"}, nil
	})
	if _, err := r.bulkDeleteTestCases(acceptCtx, args); err != nil {
		t.Fatalf("unexpected error on accept: %v", err)
	}

	// Elicitation itself errors (e.g. transport failure).
	errCtx := session.WithElicit(context.Background(), func(ctx context.Context, msg string, schema []byte) (*session.ElicitResult, error) {
		return nil, context.DeadlineExceeded
	})
	if _, err := r.bulkDeleteTestCases(errCtx, args); err == nil {
		t.Error("expected error when elicitation itself fails")
	}
}

func TestBulkDeleteTestCases_UpstreamErrorAfterConfirm(t *testing.T) {
	r := newBulkTestRegistry(t, bulkErrHandler)
	acceptCtx := session.WithElicit(context.Background(), func(ctx context.Context, msg string, schema []byte) (*session.ElicitResult, error) {
		return &session.ElicitResult{Action: "accept"}, nil
	})
	args := bulkDeleteTestCasesArgs{ProjectID: 1, TestCaseIDs: []int64{1}}
	if _, err := r.bulkDeleteTestCases(acceptCtx, args); err == nil {
		t.Error("expected error when upstream delete fails after confirmation")
	}
}

func TestBulkRunTestCasesNewLaunch_Async(t *testing.T) {
	r := newBulkTestRegistry(t, bulkOKHandler)
	result, err := r.bulkRunTestCasesNewLaunch(context.Background(), bulkRunTestCasesNewLaunchArgs{
		ProjectID: 1, TestCaseIDs: []int64{1, 2}, LaunchName: "nightly", Assignees: []string{"bob"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	task := waitForTask(t, r, m["task_id"].(string))
	if task.Status != tasks.StatusSucceeded {
		t.Errorf("task status = %s, want succeeded", task.Status)
	}
}

func TestBulkRunTestCasesNewLaunch_AsyncFails(t *testing.T) {
	r := newBulkTestRegistry(t, bulkErrHandler)
	result, err := r.bulkRunTestCasesNewLaunch(context.Background(), bulkRunTestCasesNewLaunchArgs{
		ProjectID: 1, TestCaseIDs: []int64{1},
	})
	if err != nil {
		t.Fatalf("unexpected synchronous error: %v", err)
	}
	m := result.(map[string]any)
	task := waitForTask(t, r, m["task_id"].(string))
	if task.Status != tasks.StatusFailed {
		t.Errorf("task status = %s, want failed", task.Status)
	}
}

func TestBulkRunTestCasesExistingLaunch_Async(t *testing.T) {
	r := newBulkTestRegistry(t, bulkOKHandler)
	result, err := r.bulkRunTestCasesExistingLaunch(context.Background(), bulkRunTestCasesExistingLaunchArgs{
		ProjectID: 1, TestCaseIDs: []int64{1}, LaunchID: 42,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	task := waitForTask(t, r, m["task_id"].(string))
	if task.Status != tasks.StatusSucceeded {
		t.Errorf("task status = %s, want succeeded", task.Status)
	}
}

func TestBulkCloneTestCases_Async(t *testing.T) {
	r := newBulkTestRegistry(t, bulkOKHandler)
	result, err := r.bulkCloneTestCases(context.Background(), bulkCloneTestCasesArgs{
		ProjectID: 1, TestCaseIDs: []int64{1, 2, 3},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	task := waitForTask(t, r, m["task_id"].(string))
	if task.Status != tasks.StatusSucceeded {
		t.Errorf("task status = %s, want succeeded", task.Status)
	}
}

func TestBulkCloneTestCases_AsyncFails(t *testing.T) {
	r := newBulkTestRegistry(t, bulkErrHandler)
	result, err := r.bulkCloneTestCases(context.Background(), bulkCloneTestCasesArgs{
		ProjectID: 1, TestCaseIDs: []int64{1},
	})
	if err != nil {
		t.Fatalf("unexpected synchronous error: %v", err)
	}
	m := result.(map[string]any)
	task := waitForTask(t, r, m["task_id"].(string))
	if task.Status != tasks.StatusFailed {
		t.Errorf("task status = %s, want failed", task.Status)
	}
}

// TestBulkTCSchema exercises the shared schema builder for both array and
// scalar field shapes.
func TestBulkTCSchema(t *testing.T) {
	arraySchema := bulkTCSchema("members", "array", "desc", map[string]any{"type": "object"})
	b, err := json.Marshal(arraySchema)
	if err != nil || len(b) == 0 {
		t.Fatalf("array schema marshal failed: %v", err)
	}
	props := arraySchema["properties"].(map[string]any)
	if _, ok := props["members"].(map[string]any)["items"]; !ok {
		t.Error("array field schema missing items")
	}

	scalarSchema := bulkTCSchema("layer_id", "integer", "desc", map[string]any{"type": "integer"})
	props = scalarSchema["properties"].(map[string]any)
	if _, ok := props["layer_id"].(map[string]any)["items"]; ok {
		t.Error("scalar field schema should not have items")
	}
}
