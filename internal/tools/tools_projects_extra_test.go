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
)

func newProjectsTestRegistry(t *testing.T, handler http.HandlerFunc) *Registry {
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

func TestListProjects_Success(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/project" {
			t.Errorf("path = %q", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{
				{"id": 1, "name": "Alpha", "code": "ALPHA"},
				{"id": 2, "name": "Beta", "code": "BETA"},
			},
			"number": 0, "size": 10, "totalElements": 2, "last": true,
		})
	})

	result, err := r.listProjects(context.Background(), listProjectsArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["total"] != 2 {
		t.Errorf("total = %v, want 2", m["total"])
	}
	items := m["projects"].([]map[string]any)
	if len(items) != 2 || items[0]["code"] != "ALPHA" {
		t.Errorf("projects = %+v", items)
	}
}

func TestListProjects_SizeClamped(t *testing.T) {
	var gotQuery string
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		gotQuery = req.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"content": []map[string]any{}, "last": true})
	})
	if _, err := r.listProjects(context.Background(), listProjectsArgs{Size: 1000}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "page=0&size=100" {
		t.Errorf("query = %q, want size clamped to 100", gotQuery)
	}
}

func TestListProjects_APIError(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	if _, err := r.listProjects(context.Background(), listProjectsArgs{}); err == nil {
		t.Error("expected error on 500 response")
	}
}

func TestFindProject_ValidatesQuery(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API for empty query")
	})
	if _, err := r.findProject(context.Background(), findProjectArgs{Query: "  "}); err == nil {
		t.Error("expected error for empty query")
	}
}

func TestFindProject_MatchesNameAndCode(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{
				{"id": 1, "name": "Payments Service", "code": "PAY"},
				{"id": 2, "name": "Other", "code": "OTHER"},
			},
			"last": true,
		})
	})
	result, err := r.findProject(context.Background(), findProjectArgs{Query: "pay"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["count"] != 1 {
		t.Fatalf("count = %v, want 1", m["count"])
	}
	matches := m["matches"].([]map[string]any)
	if matches[0]["id"] != int64(1) {
		t.Errorf("matched wrong project: %+v", matches[0])
	}
}

func TestFindProject_PaginatesUntilLimitOrLastPage(t *testing.T) {
	pages := 0
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		pages++
		w.Header().Set("Content-Type", "application/json")
		last := pages >= 2
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{{"id": pages, "name": "match", "code": "m"}},
			"last":    last,
		})
	})
	result, err := r.findProject(context.Background(), findProjectArgs{Query: "match", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["scanned"] != 2 {
		t.Errorf("scanned = %v, want 2 (paginated across 2 pages)", m["scanned"])
	}
	if m["truncated"] != false {
		t.Errorf("truncated = %v, want false (exhausted before limit)", m["truncated"])
	}
}

func TestFindProject_APIError(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	if _, err := r.findProject(context.Background(), findProjectArgs{Query: "x"}); err == nil {
		t.Error("expected error on upstream failure")
	}
}

func TestGetProject_Success(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/project/7" {
			t.Errorf("path = %q", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 7, "name": "Alpha", "code": "ALPHA", "description": "desc"})
	})
	result, err := r.getProject(context.Background(), getProjectArgs{ProjectID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["name"] != "Alpha" || m["description"] != "desc" {
		t.Errorf("project = %+v", m)
	}
}

func TestGetProject_ValidatesID(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API for invalid id")
	})
	if _, err := r.getProject(context.Background(), getProjectArgs{ProjectID: 0}); err == nil {
		t.Error("expected error for non-positive project_id")
	}
}

func TestGetProject_NotFound(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	if _, err := r.getProject(context.Background(), getProjectArgs{ProjectID: 999}); err == nil {
		t.Error("expected error on 404")
	}
}

func TestGetProjectStats_Success(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/project/7/stats" {
			t.Errorf("path = %q", req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": 7, "automatedTestCases": 40, "manualTestCases": 10, "automationPercent": 80.0, "launches": 5,
		})
	})
	result, err := r.getProjectStats(context.Background(), getProjectStatsArgs{ProjectID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["automated_test_cases"] != int64(40) || m["automation_percent"] != 80.0 {
		t.Errorf("stats = %+v", m)
	}
}

func TestGetProjectStats_ValidatesID(t *testing.T) {
	r := newProjectsTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not call API for invalid id")
	})
	if _, err := r.getProjectStats(context.Background(), getProjectStatsArgs{ProjectID: -1}); err == nil {
		t.Error("expected error for non-positive project_id")
	}
}
