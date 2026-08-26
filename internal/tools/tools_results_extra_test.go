package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListTestResults(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testresult", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content":       []map[string]any{{"id": 1, "name": "t1", "status": "PASSED"}},
			"number":        0,
			"size":          10,
			"totalElements": 1,
			"last":          true,
		})
	})
	r := newRelationsTestRegistry(t, mux)

	result, err := r.listTestResults(context.Background(), listTestResultsArgs{LaunchID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := result.(map[string]any)
	if out["total"] != 1 {
		t.Errorf("total = %v, want 1", out["total"])
	}

	if _, err := r.listTestResults(context.Background(), listTestResultsArgs{LaunchID: 0}); err == nil {
		t.Error("expected error for non-positive launch_id")
	}
}

func TestGetTestResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testresult/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "t1", "status": "FAILED"})
	})
	r := newRelationsTestRegistry(t, mux)

	result, err := r.getTestResult(context.Background(), getTestResultArgs{TestResultID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := result.(map[string]any)
	if out["status"] != "FAILED" {
		t.Errorf("status = %v, want FAILED", out["status"])
	}

	if _, err := r.getTestResult(context.Background(), getTestResultArgs{TestResultID: 0}); err == nil {
		t.Error("expected error for non-positive test_result_id")
	}
}

func TestAssignTestResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testresult/1/assign", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r := newRelationsTestRegistry(t, mux)

	if _, err := r.assignTestResult(context.Background(), assignTestResultArgs{TestResultID: 1, Username: "alice"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.assignTestResult(context.Background(), assignTestResultArgs{TestResultID: 0, Username: "alice"}); err == nil {
		t.Error("expected error for non-positive test_result_id")
	}
	if _, err := r.assignTestResult(context.Background(), assignTestResultArgs{TestResultID: 1}); err == nil {
		t.Error("expected error for empty username")
	}
}

func TestMuteUnmuteTestResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testresult/1/mute", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/testresult/1/unmute", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r := newRelationsTestRegistry(t, mux)

	if _, err := r.muteTestResult(context.Background(), muteTestResultArgs{TestResultID: 1, Reason: "known issue"}); err != nil {
		t.Fatalf("mute: unexpected error: %v", err)
	}
	if _, err := r.muteTestResult(context.Background(), muteTestResultArgs{TestResultID: 0}); err == nil {
		t.Error("mute: expected error for non-positive test_result_id")
	}

	if _, err := r.unmuteTestResult(context.Background(), unmuteTestResultArgs{TestResultID: 1}); err != nil {
		t.Fatalf("unmute: unexpected error: %v", err)
	}
	if _, err := r.unmuteTestResult(context.Background(), unmuteTestResultArgs{TestResultID: 0}); err == nil {
		t.Error("unmute: expected error for non-positive test_result_id")
	}
}

func TestResolveTestResult(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testresult/1/resolve", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r := newRelationsTestRegistry(t, mux)

	if _, err := r.resolveTestResult(context.Background(), resolveTestResultArgs{TestResultID: 1, Status: "passed"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.resolveTestResult(context.Background(), resolveTestResultArgs{TestResultID: 0, Status: "passed"}); err == nil {
		t.Error("expected error for non-positive test_result_id")
	}
	if _, err := r.resolveTestResult(context.Background(), resolveTestResultArgs{TestResultID: 1}); err == nil {
		t.Error("expected error for empty status")
	}
}
