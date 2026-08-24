package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
	"github.com/MimoJanra/TestOpsMCP/internal/tasks"
)

func newLaunchesTestRegistry(t *testing.T, handler http.HandlerFunc) *Registry {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/uaa/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "fake-jwt", "expires_in": 3600})
			return
		}
		handler(w, req)
	}))
	t.Cleanup(server.Close)
	client := allure.NewClient(server.URL, "test-token", 5*time.Second)
	return NewRegistry(client, core.NewLogger(core.LevelError))
}

// waitForTask is defined in tools_bulk_extra_test.go and reused here.

func TestGetLaunchEnvironment_Success(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/launch/5/env" {
			t.Errorf("path = %q", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"OS": "linux", "BRANCH": "main"})
	})
	result, err := r.getLaunchEnvironment(context.Background(), getLaunchEnvironmentArgs{LaunchID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	env := result.(map[string]any)["environment"].(map[string]any)
	if env["OS"] != "linux" {
		t.Errorf("environment = %+v", env)
	}
}

func TestGetLaunchEnvironment_ValidatesInput(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API for invalid launch_id")
	})
	if _, err := r.getLaunchEnvironment(context.Background(), getLaunchEnvironmentArgs{LaunchID: 0}); err == nil {
		t.Error("expected error for non-positive launch_id")
	}
}

func TestCopyLaunch_Success(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/launch/5/copy" {
			t.Errorf("path = %q", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 6, "name": "copy of 5"})
	})
	result, err := r.copyLaunch(context.Background(), copyLaunchArgs{LaunchID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	taskID := result.(map[string]any)["task_id"].(string)
	task := waitForTask(t, r, taskID)
	if task.Status != tasks.StatusSucceeded {
		t.Fatalf("task status = %s, error = %s", task.Status, task.Error)
	}
	res := task.Result.(map[string]any)
	if res["launch_id"] != int64(6) {
		t.Errorf("result = %+v", res)
	}
}

func TestCopyLaunch_ValidatesInput(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API for invalid launch_id")
	})
	if _, err := r.copyLaunch(context.Background(), copyLaunchArgs{LaunchID: -1}); err == nil {
		t.Error("expected error for non-positive launch_id")
	}
}

func TestCopyLaunch_TaskFailsOnAPIError(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	result, err := r.copyLaunch(context.Background(), copyLaunchArgs{LaunchID: 5})
	if err != nil {
		t.Fatalf("unexpected synchronous error: %v", err)
	}
	taskID := result.(map[string]any)["task_id"].(string)
	task := waitForTask(t, r, taskID)
	if task.Status != tasks.StatusFailed {
		t.Fatalf("task status = %s, want failed", task.Status)
	}
}

func TestMergeLaunches_Success(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/launch/merge" {
			t.Errorf("path = %q", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 42})
	})
	result, err := r.mergeLaunches(context.Background(), mergeLaunchesArgs{LaunchIDs: []int64{1, 2}, LaunchName: "merged"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	taskID := result.(map[string]any)["task_id"].(string)
	task := waitForTask(t, r, taskID)
	if task.Status != tasks.StatusSucceeded {
		t.Fatalf("task status = %s, error = %s", task.Status, task.Error)
	}
}

func TestMergeLaunches_ValidatesInput(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API")
	})
	if _, err := r.mergeLaunches(context.Background(), mergeLaunchesArgs{LaunchIDs: nil, LaunchName: "x"}); err == nil {
		t.Error("expected error for empty launch_ids")
	}
	if _, err := r.mergeLaunches(context.Background(), mergeLaunchesArgs{LaunchIDs: []int64{1}, LaunchName: ""}); err == nil {
		t.Error("expected error for empty launch_name")
	}
}

func TestAddTestCasesToLaunch_Success(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/launch/5/testcase/add" {
			t.Errorf("path = %q", req.URL.Path)
		}
		w.WriteHeader(http.StatusAccepted)
	})
	result, err := r.addTestCasesToLaunch(context.Background(), addTestCasesToLaunchArgs{
		LaunchID: 5, ProjectID: 1, TestCaseIDs: []int64{10, 11},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(map[string]any)["count"] != 2 {
		t.Errorf("result = %+v", result)
	}
}

func TestAddTestCasesToLaunch_NoJobAssignedGivesActionableError(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "no-job-assigned"})
	})
	_, err := r.addTestCasesToLaunch(context.Background(), addTestCasesToLaunchArgs{
		LaunchID: 5, ProjectID: 1, TestCaseIDs: []int64{10},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "automation_status") {
		t.Errorf("expected actionable no-job-assigned message, got: %v", err)
	}
}

func TestAddTestCasesToLaunch_ValidatesInput(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API")
	})
	cases := []addTestCasesToLaunchArgs{
		{LaunchID: 0, ProjectID: 1, TestCaseIDs: []int64{1}},
		{LaunchID: 1, ProjectID: 0, TestCaseIDs: []int64{1}},
		{LaunchID: 1, ProjectID: 1, TestCaseIDs: nil},
	}
	for _, args := range cases {
		if _, err := r.addTestCasesToLaunch(context.Background(), args); err == nil {
			t.Errorf("args=%+v: expected error", args)
		}
	}
}

func TestAddTestPlanToLaunch_Success(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/launch/5/testplan/add" {
			t.Errorf("path = %q", req.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	result, err := r.addTestPlanToLaunch(context.Background(), addTestPlanToLaunchArgs{LaunchID: 5, TestPlanID: 9})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(map[string]any)["status"] != "success" {
		t.Errorf("result = %+v", result)
	}
}

func TestAddTestPlanToLaunch_ValidatesInput(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API")
	})
	if _, err := r.addTestPlanToLaunch(context.Background(), addTestPlanToLaunchArgs{LaunchID: 1, TestPlanID: 0}); err == nil {
		t.Error("expected error for non-positive test_plan_id")
	}
}

func TestGetLaunchDefects_Success(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/launch/5/defect" {
			t.Errorf("path = %q", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"content": []any{}})
	})
	if _, err := r.getLaunchDefects(context.Background(), getLaunchDefectsArgs{LaunchID: 5, Size: 1000}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetLaunchDefects_ValidatesInput(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API")
	})
	if _, err := r.getLaunchDefects(context.Background(), getLaunchDefectsArgs{LaunchID: 0}); err == nil {
		t.Error("expected error for non-positive launch_id")
	}
}

func TestRemoveTestCasesFromLaunch_ValidatesMode(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API")
	})
	if _, err := r.removeTestCasesFromLaunch(context.Background(), removeTestCasesFromLaunchArgs{
		LaunchID: 1, TestCaseIDs: []int64{1}, Mode: "bogus",
	}); err == nil {
		t.Error("expected error for invalid mode")
	}
}

func TestRemoveTestCasesFromLaunch_HideMode(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/api/testresult":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"content": []map[string]any{
					{"id": 100, "testCaseId": 10, "name": "t1"},
					{"id": 101, "testCaseId": 20, "name": "t2"},
				},
				"last": true,
			})
		default:
			w.WriteHeader(http.StatusOK)
		}
	})
	result, err := r.removeTestCasesFromLaunch(context.Background(), removeTestCasesFromLaunchArgs{
		LaunchID: 1, TestCaseIDs: []int64{10},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["removed_count"] != 1 {
		t.Errorf("result = %+v", m)
	}
}

func TestRemoveTestCasesFromLaunch_NoMatches(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"content": []map[string]any{}, "last": true})
	})
	result, err := r.removeTestCasesFromLaunch(context.Background(), removeTestCasesFromLaunchArgs{
		LaunchID: 1, TestCaseIDs: []int64{999},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.(map[string]any)["removed_count"] != 0 {
		t.Errorf("result = %+v", result)
	}
}

func TestRemoveTestCasesFromLaunch_DeleteModeRequiresElicit(t *testing.T) {
	r := newLaunchesTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{{"id": 100, "testCaseId": 10, "name": "t1"}},
			"last":    true,
		})
	})
	_, err := r.removeTestCasesFromLaunch(context.Background(), removeTestCasesFromLaunchArgs{
		LaunchID: 1, TestCaseIDs: []int64{10}, Mode: "delete",
	})
	if err == nil {
		t.Error("expected error when mode=delete without an interactive session")
	}
}
