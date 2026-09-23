package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// fakeTreeServer models the v2 tree as observed live on project 408: tree 353
// ("Suites") with root node 93990 holding test case 19067 (leaf 61267) and
// folder 99820, which holds test case 19068 under a different leaf id (61661) —
// leaf ids are tree positions, not test case ids.
type fakeTreeServer struct {
	trees       []map[string]any
	moveBody    map[string]any
	createQuery string
	baseAqls    []string
}

func (f *fakeTreeServer) mux() *http.ServeMux {
	leaf := func(id, tc int64, name string) map[string]any {
		return map[string]any{"id": id, "type": "LEAF", "testCaseId": tc, "name": name, "status": map[string]any{"id": -1, "name": "Draft"}}
	}
	page := func(content ...map[string]any) map[string]any {
		if content == nil {
			content = []map[string]any{}
		}
		return map[string]any{"content": content, "last": true, "number": 0, "totalElements": len(content)}
	}
	m := http.NewServeMux()
	m.HandleFunc("/api/v2/tree", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"content": f.trees})
	})
	m.HandleFunc("/api/v2/project/1/test-case/tree/tree-node", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if aql := q.Get("baseAql"); aql != "" {
			f.baseAqls = append(f.baseAqls, aql)
		}
		switch q.Get("parentNodeId") {
		case "":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 93990, "children": page(
				leaf(61267, 19067, "case one"),
				map[string]any{"id": 99820, "type": "GROUP", "name": "folder", "count": 1, "parentNodeId": 93990},
			)})
		case "99820":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 99820, "name": "folder", "children": page(leaf(61661, 19068, "case two"))})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"children": page()})
		}
	})
	m.HandleFunc("/api/v2/project/1/test-case/tree/group", func(w http.ResponseWriter, r *http.Request) {
		f.createQuery = r.URL.RawQuery
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 5, "name": body["name"], "type": "GROUP", "parentNodeId": 93990, "customFieldId": -5, "customFieldValueId": 1215})
	})
	m.HandleFunc("/api/v2/test-case/tree/bulk/drag-and-drop", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&f.moveBody)
		w.WriteHeader(http.StatusAccepted)
	})
	return m
}

func newTreeTestRegistry(t *testing.T) (*Registry, *fakeTreeServer) {
	f := &fakeTreeServer{trees: []map[string]any{{"id": 353, "name": "Suites"}}}
	return newRelationsTestRegistry(t, f.mux()), f
}

func TestListTestCaseTrees(t *testing.T) {
	r, _ := newTreeTestRegistry(t)
	res, err := r.listTestCaseTrees(context.Background(), listTestCaseTreesArgs{ProjectID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	trees := res.(map[string]any)["trees"].([]map[string]any)
	if len(trees) != 1 || trees[0]["id"] != int64(353) {
		t.Errorf("unexpected trees: %+v", trees)
	}
}

func TestBrowseTestCaseTree(t *testing.T) {
	r, _ := newTreeTestRegistry(t)
	res, err := r.browseTestCaseTree(context.Background(), browseTestCaseTreeArgs{ProjectID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := res.(map[string]any)
	if out["tree_id"] != int64(353) || out["node_id"] != int64(93990) {
		t.Errorf("tree/node = %v/%v, want 353/93990 (the only tree, its root)", out["tree_id"], out["node_id"])
	}
	folders := out["folders"].([]map[string]any)
	cases := out["test_cases"].([]map[string]any)
	if len(folders) != 1 || folders[0]["id"] != int64(99820) {
		t.Errorf("unexpected folders: %+v", folders)
	}
	if len(cases) != 1 || cases[0]["test_case_id"] != int64(19067) || cases[0]["leaf_id"] != int64(61267) {
		t.Errorf("unexpected test cases: %+v", cases)
	}

	if _, err := r.browseTestCaseTree(context.Background(), browseTestCaseTreeArgs{ProjectID: 0}); err == nil {
		t.Error("expected error for non-positive project_id")
	}
}

func TestGetTestCaseTreeFolders(t *testing.T) {
	r, _ := newTreeTestRegistry(t)
	res, err := r.getTestCaseTreeFolders(context.Background(), getTestCaseTreeFoldersArgs{ProjectID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := res.(map[string]any)
	if _, has := out["test_cases"]; has {
		t.Error("folders listing should not include test cases")
	}
	if folders := out["folders"].([]map[string]any); len(folders) != 1 {
		t.Errorf("unexpected folders: %+v", folders)
	}
}

// TestMoveTestCasesToFolder guards the v2 drag-and-drop contract: it takes
// tree leaf ids (which change on every move), not test case ids, so the
// handler must resolve them — including test cases nested inside folders.
func TestMoveTestCasesToFolder(t *testing.T) {
	r, f := newTreeTestRegistry(t)
	res, err := r.moveTestCasesToFolder(context.Background(), moveTestCasesToFolderArgs{
		ProjectID: 1, TestCaseIDs: []int64{19067, 19068, 555}, NodeID: 99820,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := res.(map[string]any)
	if nf := out["not_found_test_case_ids"].([]int64); len(nf) != 1 || nf[0] != 555 {
		t.Errorf("not_found_test_case_ids = %v, want [555]", nf)
	}
	if f.moveBody["nodeId"] != float64(99820) {
		t.Errorf("nodeId = %v, want 99820", f.moveBody["nodeId"])
	}
	sel := f.moveBody["selection"].(map[string]any)
	leaves := sel["leavesInclude"].([]any)
	if len(leaves) != 2 || leaves[0] != float64(61267) || leaves[1] != float64(61661) {
		t.Errorf("leavesInclude = %v, want the resolved leaf ids [61267 61661], not test case ids", leaves)
	}
	if sel["treeId"] != float64(353) {
		t.Errorf("treeId = %v, want 353", sel["treeId"])
	}
	if len(f.baseAqls) == 0 || !strings.Contains(f.baseAqls[0], "19067") {
		t.Errorf("leaf lookup should filter the tree by the requested ids, got baseAql %v", f.baseAqls)
	}

	for name, args := range map[string]moveTestCasesToFolderArgs{
		"no project":    {TestCaseIDs: []int64{1}},
		"no test cases": {ProjectID: 1},
		"none in tree":  {ProjectID: 1, TestCaseIDs: []int64{555}},
	} {
		if _, err := r.moveTestCasesToFolder(context.Background(), args); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestMoveTestCasesToFolder_RootWhenNodeOmitted(t *testing.T) {
	r, f := newTreeTestRegistry(t)
	if _, err := r.moveTestCasesToFolder(context.Background(), moveTestCasesToFolderArgs{ProjectID: 1, TestCaseIDs: []int64{19068}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.moveBody["nodeId"] != float64(93990) {
		t.Errorf("nodeId = %v, want the tree root 93990", f.moveBody["nodeId"])
	}
}

func TestCreateTestCaseFolder(t *testing.T) {
	r, f := newTreeTestRegistry(t)
	res, err := r.createTestCaseFolder(context.Background(), createTestCaseFolderArgs{ProjectID: 1, Name: "New Folder", ParentNodeID: 99820})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := res.(map[string]any)
	if out["name"] != "New Folder" || out["folder_id"] != int64(5) {
		t.Errorf("unexpected result: %+v", out)
	}
	if !strings.Contains(f.createQuery, "treeId=353") || !strings.Contains(f.createQuery, "parentNodeId=99820") {
		t.Errorf("create query = %q, want treeId=353 and parentNodeId=99820", f.createQuery)
	}

	if _, err := r.createTestCaseFolder(context.Background(), createTestCaseFolderArgs{Name: "x"}); err == nil {
		t.Error("expected error for non-positive project_id")
	}
	if _, err := r.createTestCaseFolder(context.Background(), createTestCaseFolderArgs{ProjectID: 1}); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestResolveTreeID_MultipleTreesNeedExplicitID(t *testing.T) {
	r, f := newTreeTestRegistry(t)
	f.trees = []map[string]any{{"id": 352, "name": "Features"}, {"id": 353, "name": "Suites"}}
	_, err := r.createTestCaseFolder(context.Background(), createTestCaseFolderArgs{ProjectID: 1, Name: "x"})
	if err == nil || !strings.Contains(err.Error(), "352 (Features)") {
		t.Errorf("expected an error listing the trees to choose from, got %v", err)
	}
	if _, err := r.createTestCaseFolder(context.Background(), createTestCaseFolderArgs{ProjectID: 1, TreeID: 352, Name: "x"}); err != nil {
		t.Errorf("explicit tree_id should work: %v", err)
	}
}
