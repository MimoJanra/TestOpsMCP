package tools

import (
	"context"
	"fmt"
)

func (r *Registry) registerTreeTools() {
	r.register(&Tool{
		Name: "browse_test_case_tree",
		Description: "Browse the test case folder tree at a given path — shows test cases at that level. " +
			"Start with an empty path to see the root. Use the folder IDs from the result as the next path to go deeper. " +
			"Use this to navigate folders before moving test cases or creating new ones in the right place.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"path": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "integer"},
					"description": "Tree path as a sequence of folder IDs (empty = root)",
					"default":     []int{},
				},
				"page": map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size": map[string]any{"type": "integer", "description": "Items per page", "default": 50},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.browseTestCaseTree),
	})

	r.register(&Tool{
		Name: "get_test_case_tree_folders",
		Description: "List subfolders (groups) inside a tree path. " +
			"Use this to see the folder structure before navigating or moving test cases. " +
			"Returns folder names and IDs needed for path-based operations.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"path": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "integer"},
					"description": "Tree path as a sequence of folder IDs (empty = root)",
					"default":     []int{},
				},
				"page": map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size": map[string]any{"type": "integer", "description": "Items per page", "default": 50},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.getTestCaseTreeFolders),
	})

	r.register(&Tool{
		Name: "move_test_cases_to_folder",
		Description: "Move test cases to a specific folder in the tree (drag-and-drop). " +
			"Use browse_test_case_tree or get_test_case_tree_folders first to find the destination folder path. " +
			"Pass an empty dest_path to move to the root.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"test_case_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "integer"},
					"description": "Test case IDs to move",
				},
				"dest_path": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "integer"},
					"description": "Destination folder path as a sequence of folder IDs (empty = root)",
					"default":     []int{},
				},
			},
			"required": []string{"project_id", "test_case_ids"},
		},
		Handler: Typed(r.moveTestCasesToFolder),
	})

	r.register(&Tool{
		Name: "create_test_case_folder",
		Description: "Create a new folder (group) in the test case tree. " +
			"Use get_test_case_tree_folders to find the parent path where the folder should be created.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Folder name",
				},
				"parent_path": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "integer"},
					"description": "Parent folder path as a sequence of folder IDs (empty = create at root)",
					"default":     []int{},
				},
			},
			"required": []string{"project_id", "name"},
		},
		Handler: Typed(r.createTestCaseFolder),
	})
}

// ── handlers ─────────────────────────────────────────────────────────────────

type browseTestCaseTreeArgs struct {
	ProjectID int64   `json:"project_id"`
	Path      []int64 `json:"path"`
	Page      int     `json:"page"`
	Size      int     `json:"size"`
}

func (r *Registry) browseTestCaseTree(ctx context.Context, args browseTestCaseTreeArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	size := args.Size
	if size <= 0 {
		size = 50
	}
	result, err := r.allure.BrowseTestCaseTree(ctx, args.ProjectID, args.Path, args.Page, size)
	if err != nil {
		return nil, fmt.Errorf("browse test case tree: %w", err)
	}
	return result, nil
}

type getTestCaseTreeFoldersArgs struct {
	ProjectID int64   `json:"project_id"`
	Path      []int64 `json:"path"`
	Page      int     `json:"page"`
	Size      int     `json:"size"`
}

func (r *Registry) getTestCaseTreeFolders(ctx context.Context, args getTestCaseTreeFoldersArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	size := args.Size
	if size <= 0 {
		size = 50
	}
	result, err := r.allure.GetTestCaseTreeGroups(ctx, args.ProjectID, args.Path, args.Page, size)
	if err != nil {
		return nil, fmt.Errorf("get tree folders: %w", err)
	}
	return result, nil
}

type moveTestCasesToFolderArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	DestPath    []int64 `json:"dest_path"`
}

func (r *Registry) moveTestCasesToFolder(ctx context.Context, args moveTestCasesToFolderArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if err := r.allure.MoveTestCasesToFolder(ctx, args.ProjectID, args.TestCaseIDs, args.DestPath); err != nil {
		return nil, fmt.Errorf("move test cases to folder: %w", err)
	}
	return map[string]any{"status": "moved", "count": len(args.TestCaseIDs)}, nil
}

type createTestCaseFolderArgs struct {
	ProjectID  int64   `json:"project_id"`
	Name       string  `json:"name"`
	ParentPath []int64 `json:"parent_path"`
}

func (r *Registry) createTestCaseFolder(ctx context.Context, args createTestCaseFolderArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.Name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	result, err := r.allure.CreateTestCaseFolder(ctx, args.ProjectID, args.ParentPath, args.Name)
	if err != nil {
		return nil, fmt.Errorf("create test case folder: %w", err)
	}
	return result, nil
}
