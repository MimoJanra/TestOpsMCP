package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindSpecFile_Found(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repoRoot := filepath.Join(cwd, "..", "..")
	if _, err := os.Stat(filepath.Join(repoRoot, "spec", "testops.json")); err != nil {
		t.Skip("spec/testops.json not present at repo root in this checkout")
	}

	if err := os.Chdir(repoRoot); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	path, err := FindSpecFile()
	if err != nil {
		t.Fatalf("FindSpecFile() from repo root: unexpected error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("FindSpecFile() returned %q which does not exist: %v", path, err)
	}
}

func TestFindSpecFile_NotFound(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	empty := t.TempDir()
	if err := os.Chdir(empty); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(cwd) }()

	if _, err := FindSpecFile(); err == nil {
		t.Error("expected error when testops.json is nowhere to be found")
	}
}

func TestLoadOpenAPI_RealSpec(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(cwd, "..", "..", "spec", "testops.json")
	if _, err := os.Stat(specPath); err != nil {
		t.Skip("spec/testops.json not present at repo root in this checkout")
	}

	spec, err := LoadOpenAPI(specPath)
	if err != nil {
		t.Fatalf("LoadOpenAPI: unexpected error: %v", err)
	}
	if len(spec.Paths) == 0 {
		t.Error("expected the real spec to have at least one path")
	}

	idx, err := BuildOperationsIndex(spec)
	if err != nil {
		t.Fatalf("BuildOperationsIndex: unexpected error: %v", err)
	}
	if len(idx.ListAll()) == 0 {
		t.Error("expected the real spec to yield at least one operation")
	}
}

func TestLoadOpenAPI_MissingFile(t *testing.T) {
	if _, err := LoadOpenAPI(filepath.Join(t.TempDir(), "does-not-exist.json")); err == nil {
		t.Error("expected error for a missing spec file")
	}
}

func TestLoadOpenAPI_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadOpenAPI(path); err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestBuildOperationsIndex_NilPaths(t *testing.T) {
	idx, err := BuildOperationsIndex(&OpenAPISpec{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(idx.ListAll()) != 0 {
		t.Errorf("expected empty index, got %d operations", len(idx.ListAll()))
	}
	if idx.Get("anything") != nil {
		t.Error("Get on empty index should return nil")
	}
}

func TestBuildOperationsIndex_SkipsMalformedEntries(t *testing.T) {
	spec := &OpenAPISpec{
		Paths: map[string]map[string]interface{}{
			"/valid/{id}": {
				"get": map[string]interface{}{
					"operationId": "get_valid",
					"summary":     "Get a valid thing",
					"description": "Fetches a thing by id",
					"tags":        []interface{}{"things", "read"},
					"parameters": []interface{}{
						map[string]interface{}{
							"name":        "id",
							"in":          "path",
							"description": "thing id",
							"required":    true,
							"schema":      map[string]interface{}{"type": "integer"},
						},
						// Malformed: not a map -> parseParameter returns nil, dropped.
						"not-a-param-map",
						// Malformed: missing name -> parseParameter returns nil, dropped.
						map[string]interface{}{"in": "query"},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content":  map[string]interface{}{"application/json": map[string]interface{}{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "ok"},
					},
				},
				// nil method spec -> skipped entirely.
				"delete": nil,
			},
			// nil methods map for this path -> skipped entirely.
			"/nothing": nil,
			"/bad": {
				// methodSpec not a map -> parseOperation returns nil.
				"post": "not-a-map",
				// missing operationId -> parseOperation returns nil.
				"put": map[string]interface{}{"summary": "no id"},
			},
			"/badbody": {
				"get": map[string]interface{}{
					"operationId": "get_badbody",
					// requestBody present but not a map -> parseRequestBody returns nil.
					"requestBody": "not-a-map",
					// tags present but wrong element type -> getStringArray returns nil.
					"tags": []interface{}{123, true},
				},
			},
		},
	}

	idx, err := BuildOperationsIndex(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := idx.ListAll()
	if len(all) != 2 {
		names := make([]string, len(all))
		for i, op := range all {
			names[i] = op.OperationID
		}
		t.Fatalf("expected exactly 2 valid operations, got %d: %v", len(all), names)
	}

	op := idx.Get("get_valid")
	if op == nil {
		t.Fatal("expected get_valid to be indexed")
	}
	if op.Method != "GET" {
		t.Errorf("Method = %q, want GET (should be upper-cased)", op.Method)
	}
	if op.Summary != "Get a valid thing" || op.Description != "Fetches a thing by id" {
		t.Errorf("summary/description not parsed: %+v", op)
	}
	if len(op.Tags) != 2 || op.Tags[0] != "things" {
		t.Errorf("tags not parsed: %v", op.Tags)
	}
	if len(op.Parameters) != 1 || op.Parameters[0].Name != "id" || !op.Parameters[0].Required {
		t.Errorf("expected exactly 1 valid parameter (malformed ones dropped), got %+v", op.Parameters)
	}
	if op.RequestBody == nil || !op.RequestBody.Required {
		t.Errorf("expected requestBody.required=true, got %+v", op.RequestBody)
	}
	if byTag := idx.byTag["things"]; len(byTag) != 1 {
		t.Errorf("expected 1 operation tagged 'things', got %d", len(byTag))
	}

	badbody := idx.Get("get_badbody")
	if badbody == nil {
		t.Fatal("expected get_badbody to be indexed despite malformed requestBody/tags")
	}
	if badbody.RequestBody != nil {
		t.Errorf("expected nil RequestBody for non-map requestBody, got %+v", badbody.RequestBody)
	}
	if badbody.Tags != nil {
		t.Errorf("expected nil Tags for wrong-typed tags array, got %v", badbody.Tags)
	}

	if idx.Get("no_such_op") != nil {
		t.Error("Get for unknown operation id should return nil")
	}
}

func TestOperationsIndex_SearchAndScore(t *testing.T) {
	spec := &OpenAPISpec{
		Paths: map[string]map[string]interface{}{
			"/launch": {
				"post": map[string]interface{}{
					"operationId": "create_launch",
					"summary":     "Create a new launch",
					"description": "Starts a new test launch",
					"tags":        []interface{}{"launches"},
				},
			},
			"/launch/{id}": {
				"get": map[string]interface{}{
					"operationId": "get_launch",
					"summary":     "Get launch details",
					"description": "Fetches details for a launch",
					"tags":        []interface{}{"launches"},
				},
			},
			"/project": {
				"get": map[string]interface{}{
					"operationId": "list_projects",
					"summary":     "List projects",
					"description": "Lists all projects",
					"tags":        []interface{}{"projects"},
				},
			},
		},
	}
	idx, err := BuildOperationsIndex(spec)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("matches by operation id", func(t *testing.T) {
		results := idx.Search("create_launch")
		if len(results) != 1 || results[0].OperationID != "create_launch" {
			t.Fatalf("got %v", results)
		}
	})

	t.Run("matches by tag ranks below operation id/summary match but is still returned", func(t *testing.T) {
		results := idx.Search("launch")
		if len(results) != 2 {
			t.Fatalf("expected 2 launch operations, got %d: %v", len(results), results)
		}
		// create_launch matches operationId (score 100 + summary prefix nothing) and
		// get_launch matches operationId too; both should be present regardless of order.
		ids := map[string]bool{results[0].OperationID: true, results[1].OperationID: true}
		if !ids["create_launch"] || !ids["get_launch"] {
			t.Errorf("expected both launch operations, got %v", results)
		}
	})

	t.Run("no match returns empty", func(t *testing.T) {
		if results := idx.Search("nonexistent-keyword-xyz"); len(results) != 0 {
			t.Errorf("expected no results, got %v", results)
		}
	})

	t.Run("case-insensitive", func(t *testing.T) {
		if results := idx.Search("PROJECTS"); len(results) != 1 || results[0].OperationID != "list_projects" {
			t.Errorf("expected case-insensitive match, got %v", results)
		}
	})

	t.Run("ListAll is sorted by operation id", func(t *testing.T) {
		all := idx.ListAll()
		for i := 1; i < len(all); i++ {
			if all[i-1].OperationID > all[i].OperationID {
				t.Fatalf("ListAll not sorted: %q before %q", all[i-1].OperationID, all[i].OperationID)
			}
		}
	})
}

func TestMatchesQueryAndScoreMatch(t *testing.T) {
	op := &Operation{
		OperationID: "create_launch",
		Summary:     "Create a new launch",
		Description: "Starts a new test launch run",
		Path:        "/api/launch",
		Tags:        []string{"launches", "write"},
	}

	if !matchesQuery(op, "launch") {
		t.Error("expected match on operation id substring")
	}
	if !matchesQuery(op, "starts a new") {
		t.Error("expected match on description substring")
	}
	if !matchesQuery(op, "write") {
		t.Error("expected match on tag")
	}
	if matchesQuery(op, "totally-unrelated") {
		t.Error("expected no match")
	}

	if score := scoreMatch(op, "create_launch"); score < 100 {
		t.Errorf("exact operation id match score = %d, want >= 100", score)
	}
	if score := scoreMatch(op, "create"); score < 50 {
		t.Errorf("summary-prefix match score = %d, want >= 50 (prefix+contains bonus)", score)
	}
	if score := scoreMatch(op, "/api/launch"); score < 10 {
		t.Errorf("path match score = %d, want >= 10", score)
	}
	if score := scoreMatch(op, "nope"); score != 0 {
		t.Errorf("no-match score = %d, want 0", score)
	}
}

func TestGetStringValueAndArray(t *testing.T) {
	m := map[string]interface{}{
		"name":     "hello",
		"wrongTyp": 42,
		"tags":     []interface{}{"a", "b", 3},
		"notArr":   "nope",
	}

	if v := getStringValue(m, "name"); v != "hello" {
		t.Errorf("getStringValue(name) = %q", v)
	}
	if v := getStringValue(m, "wrongTyp"); v != "" {
		t.Errorf("getStringValue(wrongTyp) = %q, want empty for non-string", v)
	}
	if v := getStringValue(m, "missing"); v != "" {
		t.Errorf("getStringValue(missing) = %q, want empty", v)
	}

	if arr := getStringArray(m, "tags"); len(arr) != 2 || arr[0] != "a" || arr[1] != "b" {
		t.Errorf("getStringArray(tags) = %v, want [a b] (non-string elements dropped)", arr)
	}
	if arr := getStringArray(m, "notArr"); arr != nil {
		t.Errorf("getStringArray(notArr) = %v, want nil for non-array value", arr)
	}
	if arr := getStringArray(m, "missing"); arr != nil {
		t.Errorf("getStringArray(missing) = %v, want nil", arr)
	}
}
