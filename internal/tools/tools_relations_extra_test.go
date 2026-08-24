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

// newRelationsTestRegistry starts a fake Allure server that understands the
// handful of endpoints tools_relations.go's handlers call, and returns a
// Registry wired to it.
func newRelationsTestRegistry(t *testing.T, mux *http.ServeMux) *Registry {
	t.Helper()
	mux.HandleFunc("/api/uaa/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "fake-jwt", "expires_in": 3600})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	client := allure.NewClient(server.URL, "test-token", 5*time.Second)
	return NewRegistry(client, core.NewLogger(core.LevelError))
}

func TestGetTestCaseDefects(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcase/1/defect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"content": []any{}})
	})
	r := newRelationsTestRegistry(t, mux)

	if _, err := r.getTestCaseDefects(context.Background(), getTestCaseDefectsArgs{TestCaseID: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.getTestCaseDefects(context.Background(), getTestCaseDefectsArgs{TestCaseID: 0}); err == nil {
		t.Error("expected error for non-positive test_case_id")
	}
}

func TestAddRemoveTestCaseDefect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcase/1/defect/2", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r := newRelationsTestRegistry(t, mux)

	if _, err := r.addTestCaseDefect(context.Background(), addTestCaseDefectArgs{TestCaseID: 1, DefectID: 2}); err != nil {
		t.Fatalf("add: unexpected error: %v", err)
	}
	if _, err := r.addTestCaseDefect(context.Background(), addTestCaseDefectArgs{TestCaseID: 0, DefectID: 2}); err == nil {
		t.Error("add: expected error for non-positive test_case_id")
	}
	if _, err := r.addTestCaseDefect(context.Background(), addTestCaseDefectArgs{TestCaseID: 1, DefectID: 0}); err == nil {
		t.Error("add: expected error for non-positive defect_id")
	}

	if _, err := r.removeTestCaseDefect(context.Background(), removeTestCaseDefectArgs{TestCaseID: 1, DefectID: 2}); err != nil {
		t.Fatalf("remove: unexpected error: %v", err)
	}
	if _, err := r.removeTestCaseDefect(context.Background(), removeTestCaseDefectArgs{TestCaseID: 0, DefectID: 2}); err == nil {
		t.Error("remove: expected error for non-positive test_case_id")
	}
}

func TestGetTestCaseMembers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcase/1/members", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 5, "name": "Alice"}})
	})
	r := newRelationsTestRegistry(t, mux)

	result, err := r.getTestCaseMembers(context.Background(), getTestCaseMembersArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	members, _ := out["members"].([]map[string]any)
	if len(members) != 1 || members[0]["name"] != "Alice" {
		t.Errorf("members = %+v, want one member named Alice", members)
	}

	if _, err := r.getTestCaseMembers(context.Background(), getTestCaseMembersArgs{TestCaseID: 0}); err == nil {
		t.Error("expected error for non-positive test_case_id")
	}
}

func TestAddTestCaseMembers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcase/1/members", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	r := newRelationsTestRegistry(t, mux)

	args := addTestCaseMembersArgs{TestCaseID: 1, Members: []allure.MemberDto{{ID: 1, Name: "Alice"}}}
	if _, err := r.addTestCaseMembers(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.addTestCaseMembers(context.Background(), addTestCaseMembersArgs{TestCaseID: 0, Members: args.Members}); err == nil {
		t.Error("expected error for non-positive test_case_id")
	}
	if _, err := r.addTestCaseMembers(context.Background(), addTestCaseMembersArgs{TestCaseID: 1}); err == nil {
		t.Error("expected error for empty members")
	}
}

func TestRemoveTestCaseMembers(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcase/bulk/member/remove", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r := newRelationsTestRegistry(t, mux)

	args := removeTestCaseMembersArgs{ProjectID: 1, TestCaseID: 1, MemberIDs: []int64{5}}
	if _, err := r.removeTestCaseMembers(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.removeTestCaseMembers(context.Background(), removeTestCaseMembersArgs{TestCaseID: 1, MemberIDs: []int64{5}}); err == nil {
		t.Error("expected error for non-positive project_id")
	}
	if _, err := r.removeTestCaseMembers(context.Background(), removeTestCaseMembersArgs{ProjectID: 1, MemberIDs: []int64{5}}); err == nil {
		t.Error("expected error for non-positive test_case_id")
	}
	if _, err := r.removeTestCaseMembers(context.Background(), removeTestCaseMembersArgs{ProjectID: 1, TestCaseID: 1}); err == nil {
		t.Error("expected error for empty member_ids")
	}
}

func TestGetTestCaseExternalLinks(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcase/1/overview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"links": []map[string]any{{"name": "Bug", "type": "JIRA", "url": "https://example.com/BUG-1"}},
		})
	})
	r := newRelationsTestRegistry(t, mux)

	result, err := r.getTestCaseExternalLinks(context.Background(), getTestCaseExternalLinksArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := result.(map[string]any)
	links, _ := out["links"].([]map[string]any)
	if len(links) != 1 || links[0]["url"] != "https://example.com/BUG-1" {
		t.Errorf("links = %+v", links)
	}

	if _, err := r.getTestCaseExternalLinks(context.Background(), getTestCaseExternalLinksArgs{TestCaseID: 0}); err == nil {
		t.Error("expected error for non-positive test_case_id")
	}
}

func TestAddTestCaseExternalLink(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcase/1/overview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"links": []map[string]any{}})
	})
	var patchedBody map[string]any
	mux.HandleFunc("/api/testcase/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		_ = json.NewDecoder(r.Body).Decode(&patchedBody)
		w.WriteHeader(http.StatusOK)
	})
	r := newRelationsTestRegistry(t, mux)

	args := addTestCaseExternalLinkArgs{TestCaseID: 1, URL: "https://example.com/NEW-1", Name: "New", Type: "JIRA"}
	if _, err := r.addTestCaseExternalLink(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	links, _ := patchedBody["links"].([]any)
	if len(links) != 1 {
		t.Errorf("expected the PATCH body to carry 1 link, got %+v", patchedBody)
	}

	if _, err := r.addTestCaseExternalLink(context.Background(), addTestCaseExternalLinkArgs{TestCaseID: 0, URL: "x"}); err == nil {
		t.Error("expected error for non-positive test_case_id")
	}
	if _, err := r.addTestCaseExternalLink(context.Background(), addTestCaseExternalLinkArgs{TestCaseID: 1}); err == nil {
		t.Error("expected error for empty url")
	}
}

func TestDeleteTestCaseExternalLink(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcase/1/overview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"links": []map[string]any{{"name": "Bug", "type": "JIRA", "url": "https://example.com/BUG-1"}},
		})
	})
	mux.HandleFunc("/api/testcase/1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r := newRelationsTestRegistry(t, mux)

	if _, err := r.deleteTestCaseExternalLink(context.Background(), deleteTestCaseExternalLinkArgs{TestCaseID: 1, URL: "https://example.com/BUG-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.deleteTestCaseExternalLink(context.Background(), deleteTestCaseExternalLinkArgs{TestCaseID: 1, URL: "https://example.com/missing"}); err == nil {
		t.Error("expected error when the URL isn't among the current links")
	}
	if _, err := r.deleteTestCaseExternalLink(context.Background(), deleteTestCaseExternalLinkArgs{TestCaseID: 0, URL: "x"}); err == nil {
		t.Error("expected error for non-positive test_case_id")
	}
	if _, err := r.deleteTestCaseExternalLink(context.Background(), deleteTestCaseExternalLinkArgs{TestCaseID: 1}); err == nil {
		t.Error("expected error for empty url")
	}
}
