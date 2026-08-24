package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestBrowseTestCaseTree(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcasetree/leaf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"content": []any{}})
	})
	r := newRelationsTestRegistry(t, mux)

	if _, err := r.browseTestCaseTree(context.Background(), browseTestCaseTreeArgs{ProjectID: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.browseTestCaseTree(context.Background(), browseTestCaseTreeArgs{ProjectID: 0}); err == nil {
		t.Error("expected error for non-positive project_id")
	}
}

func TestGetTestCaseTreeFolders(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcasetree/group", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"content": []any{}})
	})
	r := newRelationsTestRegistry(t, mux)

	if _, err := r.getTestCaseTreeFolders(context.Background(), getTestCaseTreeFoldersArgs{ProjectID: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.getTestCaseTreeFolders(context.Background(), getTestCaseTreeFoldersArgs{ProjectID: 0}); err == nil {
		t.Error("expected error for non-positive project_id")
	}
}

func TestMoveTestCasesToFolder(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcase/bulk/draganddrop", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r := newRelationsTestRegistry(t, mux)

	args := moveTestCasesToFolderArgs{ProjectID: 1, TestCaseIDs: []int64{1, 2}}
	if _, err := r.moveTestCasesToFolder(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := r.moveTestCasesToFolder(context.Background(), moveTestCasesToFolderArgs{TestCaseIDs: []int64{1}}); err == nil {
		t.Error("expected error for non-positive project_id")
	}
	if _, err := r.moveTestCasesToFolder(context.Background(), moveTestCasesToFolderArgs{ProjectID: 1}); err == nil {
		t.Error("expected error for empty test_case_ids")
	}
}

func TestCreateTestCaseFolder(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/testcasetree/group", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 9, "name": "New Folder"})
	})
	r := newRelationsTestRegistry(t, mux)

	result, err := r.createTestCaseFolder(context.Background(), createTestCaseFolderArgs{ProjectID: 1, Name: "New Folder"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := result.(map[string]any)
	if out["name"] != "New Folder" {
		t.Errorf("name = %v, want New Folder", out["name"])
	}

	if _, err := r.createTestCaseFolder(context.Background(), createTestCaseFolderArgs{Name: "x"}); err == nil {
		t.Error("expected error for non-positive project_id")
	}
	if _, err := r.createTestCaseFolder(context.Background(), createTestCaseFolderArgs{ProjectID: 1}); err == nil {
		t.Error("expected error for empty name")
	}
}
