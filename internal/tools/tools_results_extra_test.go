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
	stored := "alice"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testresult/1/assign", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/testresult/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "assignee": stored})
	})
	r := newRelationsTestRegistry(t, mux)

	if _, err := r.assignTestResult(context.Background(), assignTestResultArgs{TestResultID: 1, Username: "alice"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A resolved result keeps its old assignee although the call succeeds.
	stored = "bob"
	if _, err := r.assignTestResult(context.Background(), assignTestResultArgs{TestResultID: 1, Username: "alice"}); err == nil {
		t.Error("expected error when the API keeps a different assignee")
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

// resolveMux serves the two calls resolve_test_result makes: the test result
// lookup (for its launch id) and the v2 bulk resolve with a one-item selection.
func resolveMux(body *map[string]any) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testresult/1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"launchId":7}`))
	})
	mux.HandleFunc("/api/v2/test-result/bulk/resolve", func(w http.ResponseWriter, req *http.Request) {
		if body != nil {
			_ = json.NewDecoder(req.Body).Decode(body)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return mux
}

func TestResolveTestResult(t *testing.T) {
	r := newRelationsTestRegistry(t, resolveMux(nil))

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

// TestResolveTestResult_SendsCategoryAndMessage guards against
// category_id/message being silently dropped — the web UI's "Change status"
// dialog has Status/Category/Details fields, but the tool previously only
// ever sent status, and the v1 single-result endpoint drops message anyway
// (confirmed live), so this must go through v2 with a one-result selection.
func TestResolveTestResult_SendsCategoryAndMessage(t *testing.T) {
	var body map[string]any
	r := newRelationsTestRegistry(t, resolveMux(&body))

	_, err := r.resolveTestResult(context.Background(), resolveTestResultArgs{
		TestResultID: 1, Status: "passed", CategoryID: 5, Message: "covered in launch #123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body["categoryId"] != float64(5) {
		t.Errorf("categoryId = %v, want 5", body["categoryId"])
	}
	if body["message"] != "covered in launch #123" {
		t.Errorf("message = %v, want %q", body["message"], "covered in launch #123")
	}
	sel, _ := body["selection"].(map[string]any)
	if sel["launchId"] != float64(7) {
		t.Errorf("selection.launchId = %v, want 7 (looked up from the test result)", sel["launchId"])
	}
	if ids, _ := sel["leafsInclude"].([]any); len(ids) != 1 || ids[0] != float64(1) {
		t.Errorf("selection.leafsInclude = %v, want [1]", sel["leafsInclude"])
	}
}
