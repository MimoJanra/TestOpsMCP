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

// TestListTestResults_SizeAboveOldCap guards against the size param silently
// clamping to 100 — the API accepts much larger pages (confirmed live:
// size=300 returns a full 300-item page, no server-side clamp).
func TestListTestResults_SizeAboveOldCap(t *testing.T) {
	var gotSize string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testresult", func(w http.ResponseWriter, r *http.Request) {
		gotSize = r.URL.Query().Get("size")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{}, "number": 0, "size": 300, "totalElements": 0, "last": true,
		})
	})
	r := newRelationsTestRegistry(t, mux)

	result, err := r.listTestResults(context.Background(), listTestResultsArgs{LaunchID: 1, Size: 300})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotSize != "300" {
		t.Errorf("request size = %q, want \"300\" (should not be clamped to 100)", gotSize)
	}
	if out := result.(map[string]any); out["size"] != 300 {
		t.Errorf("response size = %v, want 300", out["size"])
	}
}

// TestListTestResults_StatusFilter guards against #22: GET /api/testresult has
// no server-side status filter (a "status" query param is silently ignored),
// so a naive pass-through returned every status regardless of the requested
// filter. This verifies the client-side scan-and-paginate filtering actually
// excludes non-matching statuses and paginates correctly over the matches.
func TestListTestResults_StatusFilter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testresult", func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("status"); got != "" {
			t.Errorf("request should not send a status query param (API ignores it), got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{
				{"id": 1, "name": "t1", "status": "passed"},
				{"id": 2, "name": "t2", "status": "failed"},
				{"id": 3, "name": "t3", "status": "skipped"},
				{"id": 4, "name": "t4", "status": "failed"},
				{"id": 5, "name": "t5", "status": "passed"},
			},
			"number": 0, "size": 100, "totalElements": 5, "last": true,
		})
	})
	r := newRelationsTestRegistry(t, mux)

	result, err := r.listTestResults(context.Background(), listTestResultsArgs{LaunchID: 1, Status: "failed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := result.(map[string]any)
	items := out["test_results"].([]map[string]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 failed results, got %d: %+v", len(items), items)
	}
	if items[0]["id"] != int64(2) || items[1]["id"] != int64(4) {
		t.Errorf("unexpected failed result ids: %v, %v", items[0]["id"], items[1]["id"])
	}
	if out["is_last"] != true {
		t.Errorf("is_last = %v, want true", out["is_last"])
	}

	// Second failed item alone, via page/size over the filtered set.
	result, err = r.listTestResults(context.Background(), listTestResultsArgs{LaunchID: 1, Status: "FAILED", Page: 1, Size: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out = result.(map[string]any)
	items = out["test_results"].([]map[string]any)
	if len(items) != 1 || items[0]["id"] != int64(4) {
		t.Errorf("expected page 1 to return only id 4, got %+v", items)
	}
	if out["is_last"] != true {
		t.Errorf("is_last = %v, want true", out["is_last"])
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
