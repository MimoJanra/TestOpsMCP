package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
)

// treeIDSchema is shared by every folder tool: in Allure a folder only exists
// within a named tree (built from custom fields), and tree positions are node
// ids, not paths.
var treeIDSchema = map[string]any{
	"type": "integer",
	"description": "Tree ID from list_test_case_trees (e.g. the \"Suites\" tree). Optional when the project has " +
		"exactly one tree.",
}

func (r *Registry) registerTreeTools() {
	r.register(&Tool{
		Name: "list_test_case_trees",
		Description: "List the project's test case trees (e.g. \"Suites\", \"Features\"). Folders in Allure are values of " +
			"custom fields and only exist within a tree, so every other folder tool needs a tree_id from here.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.listTestCaseTrees),
	})

	r.register(&Tool{
		Name: "browse_test_case_tree",
		Description: "Browse one level of a test case tree: the folders and test cases directly inside a folder " +
			"(or the tree root when parent_node_id is omitted). Use a folder's id as the next parent_node_id to go deeper.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":     map[string]any{"type": "integer", "description": "Allure project ID"},
				"tree_id":        treeIDSchema,
				"parent_node_id": map[string]any{"type": "integer", "description": "Folder node ID to list (omit for the tree root)"},
				"page":           map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":           map[string]any{"type": "integer", "description": "Items per page", "default": 50},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.browseTestCaseTree),
	})

	r.register(&Tool{
		Name: "get_test_case_tree_folders",
		Description: "List only the folders directly inside a folder of a test case tree (or the tree root when " +
			"parent_node_id is omitted). Returns folder node IDs for create_test_case_folder / move_test_cases_to_folder. " +
			"page/size apply to the folder list (the API interleaves folders with test cases, so the tool collects them first).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":     map[string]any{"type": "integer", "description": "Allure project ID"},
				"tree_id":        treeIDSchema,
				"parent_node_id": map[string]any{"type": "integer", "description": "Folder node ID to list (omit for the tree root)"},
				"page":           map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":           map[string]any{"type": "integer", "description": "Items per page", "default": 50},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.getTestCaseTreeFolders),
	})

	r.register(&Tool{
		Name: "move_test_cases_to_folder",
		Description: "Move test cases into a folder of a test case tree (omit node_id to move them to the tree root). " +
			"Moving sets the test case's custom field value for that folder and replaces every folder the case had in this tree. " +
			"node_id must be a folder of this tree (checked — a leaf id or another tree's folder would silently strip the folder). " +
			"Asynchronous — the move lands shortly after the call returns.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":    map[string]any{"type": "integer", "description": "Allure project ID"},
				"tree_id":       treeIDSchema,
				"test_case_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Test case IDs to move"},
				"node_id":       map[string]any{"type": "integer", "description": "Destination folder node ID (omit for the tree root)"},
			},
			"required": []string{"project_id", "test_case_ids"},
		},
		Handler: Typed(r.moveTestCasesToFolder),
	})

	r.register(&Tool{
		Name: "create_test_case_folder",
		Description: "Create a folder in a test case tree (at the tree root when parent_node_id is omitted). " +
			"The folder is a new value of the tree's custom field at that level.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":     map[string]any{"type": "integer", "description": "Allure project ID"},
				"tree_id":        treeIDSchema,
				"name":           map[string]any{"type": "string", "description": "Folder name"},
				"parent_node_id": map[string]any{"type": "integer", "description": "Parent folder node ID (omit for the tree root)"},
			},
			"required": []string{"project_id", "name"},
		},
		Handler: Typed(r.createTestCaseFolder),
	})

	r.register(&Tool{
		Name: "rename_test_case_folder",
		Description: "Rename a folder of a test case tree in place: it keeps its node id and its test cases follow. " +
			"Other folders with the same name elsewhere in the tree are not affected.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"tree_id":    treeIDSchema,
				"node_id":    map[string]any{"type": "integer", "description": "Folder node ID"},
				"name":       map[string]any{"type": "string", "description": "New folder name"},
			},
			"required": []string{"project_id", "node_id", "name"},
		},
		Handler: Typed(r.renameTestCaseFolder),
	})

	r.register(&Tool{
		Name: "delete_test_case_folder",
		Description: "Delete a folder of a test case tree. Its test cases are not deleted — they lose this folder assignment. " +
			"Do NOT use delete_custom_field_value for this: one value can back several same-named folders, and deleting it removes all of them.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"tree_id":    treeIDSchema,
				"node_id":    map[string]any{"type": "integer", "description": "Folder node ID"},
			},
			"required": []string{"project_id", "node_id"},
		},
		Handler: Typed(r.deleteTestCaseFolder),
	})
}

// ── handlers ─────────────────────────────────────────────────────────────────

const treePageSize = 100

// resolveTreeID returns treeID when set, otherwise the project's only tree.
// With several trees there is no safe default, so it errors listing them.
func (r *Registry) resolveTreeID(ctx context.Context, projectID, treeID int64) (int64, error) {
	if treeID > 0 {
		return treeID, nil
	}
	trees, err := r.allure.ListTestCaseTrees(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("list test case trees: %w", err)
	}
	switch len(trees) {
	case 0:
		return 0, fmt.Errorf("project %d has no test case trees", projectID)
	case 1:
		return trees[0].ID, nil
	}
	names := make([]string, len(trees))
	for i, t := range trees {
		names[i] = fmt.Sprintf("%d (%s)", t.ID, t.Name)
	}
	return 0, fmt.Errorf("project %d has %d trees, pass tree_id: %s", projectID, len(trees), strings.Join(names, ", "))
}

type listTestCaseTreesArgs struct {
	ProjectID int64 `json:"project_id"`
}

func (r *Registry) listTestCaseTrees(ctx context.Context, args listTestCaseTreesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	trees, err := r.allure.ListTestCaseTrees(ctx, args.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("list test case trees: %w", err)
	}
	items := make([]map[string]any, len(trees))
	for i, t := range trees {
		items[i] = map[string]any{"id": t.ID, "name": t.Name}
	}
	return map[string]any{"trees": items}, nil
}

type browseTestCaseTreeArgs struct {
	ProjectID    int64 `json:"project_id"`
	TreeID       int64 `json:"tree_id"`
	ParentNodeID int64 `json:"parent_node_id"`
	Page         int   `json:"page"`
	Size         int   `json:"size"`
}

// listTreeLevel fetches one level of a tree and splits it into folders and
// test cases.
func (r *Registry) listTreeLevel(ctx context.Context, args browseTestCaseTreeArgs) (treeID int64, node *allure.TestCaseTreeNodeResponse, folders, testCases []map[string]any, err error) {
	if args.ProjectID <= 0 {
		return 0, nil, nil, nil, fmt.Errorf("project_id must be positive")
	}
	treeID, err = r.resolveTreeID(ctx, args.ProjectID, args.TreeID)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	size := args.Size
	if size <= 0 {
		size = 50
	}
	node, err = r.allure.GetTestCaseTreeNode(ctx, args.ProjectID, treeID, args.ParentNodeID, "", args.Page, size)
	if err != nil {
		return 0, nil, nil, nil, err
	}
	folders = []map[string]any{}
	testCases = []map[string]any{}
	for _, c := range node.Children.Content {
		if c.Type == "GROUP" {
			folders = append(folders, map[string]any{"id": c.ID, "name": c.Name, "test_case_count": c.Count, "custom_field_value_id": c.CustomFieldValueID})
			continue
		}
		tc := map[string]any{"test_case_id": c.TestCaseID, "leaf_id": c.ID, "name": c.Name, "automated": c.Automated}
		if c.Status != nil {
			tc["status"] = c.Status.Name
		}
		testCases = append(testCases, tc)
	}
	return treeID, node, folders, testCases, nil
}

func (r *Registry) browseTestCaseTree(ctx context.Context, args browseTestCaseTreeArgs) (any, error) {
	treeID, node, folders, testCases, err := r.listTreeLevel(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("browse test case tree: %w", err)
	}
	return map[string]any{
		"tree_id":    treeID,
		"node_id":    node.ID,
		"name":       node.Name,
		"folders":    folders,
		"test_cases": testCases,
		"page":       node.Children.Number,
		"total":      node.Children.Total,
		"is_last":    node.Children.Last,
	}, nil
}

type getTestCaseTreeFoldersArgs = browseTestCaseTreeArgs

func (r *Registry) getTestCaseTreeFolders(ctx context.Context, args getTestCaseTreeFoldersArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	treeID, err := r.resolveTreeID(ctx, args.ProjectID, args.TreeID)
	if err != nil {
		return nil, err
	}
	// Folders and test cases share one name-sorted listing, so a page of it
	// can hold no folders at all even though later pages do (confirmed live).
	// Collect every folder of the level, then page over those.
	folders := []map[string]any{}
	var nodeID int64
	var nodeName string
	for page := 0; page < 200; page++ {
		node, err := r.allure.GetTestCaseTreeNode(ctx, args.ProjectID, treeID, args.ParentNodeID, "", page, treePageSize)
		if err != nil {
			return nil, fmt.Errorf("get tree folders: %w", err)
		}
		nodeID, nodeName = node.ID, node.Name
		for _, c := range node.Children.Content {
			if c.Type == "GROUP" {
				folders = append(folders, map[string]any{"id": c.ID, "name": c.Name, "test_case_count": c.Count, "custom_field_value_id": c.CustomFieldValueID})
			}
		}
		if node.Children.Last || len(node.Children.Content) == 0 {
			break
		}
	}
	size := args.Size
	if size <= 0 {
		size = 50
	}
	start := min(args.Page*size, len(folders))
	end := min(start+size, len(folders))
	return map[string]any{
		"tree_id": treeID,
		"node_id": nodeID,
		"name":    nodeName,
		"folders": folders[start:end],
		"page":    args.Page,
		"total":   len(folders),
		"is_last": end >= len(folders),
	}, nil
}

// requireTreeFolder checks that nodeID is a folder of this tree. The API
// accepts a leaf id, a bogus id or another tree's folder as a move target and
// silently strips the test cases' folder instead (confirmed live).
func (r *Registry) requireTreeFolder(ctx context.Context, projectID, treeID, nodeID int64) error {
	folders, err := r.treeFolders(ctx, projectID, treeID)
	if err != nil {
		return fmt.Errorf("look up folder %d: %w", nodeID, err)
	}
	for _, f := range folders {
		if f.ID == nodeID {
			return nil
		}
	}
	return fmt.Errorf("node %d is not a folder of tree %d — use a folder id from browse_test_case_tree / get_test_case_tree_folders for this tree", nodeID, treeID)
}

// treeFolders walks every folder of a tree (breadth-first).
func (r *Registry) treeFolders(ctx context.Context, projectID, treeID int64) ([]allure.TestCaseTreeNodeDto, error) {
	var out []allure.TestCaseTreeNodeDto
	queue := []int64{0}
	for visited := 0; len(queue) > 0; visited++ {
		if visited > 5000 {
			return nil, fmt.Errorf("tree %d has more than 5000 folders", treeID)
		}
		parent := queue[0]
		queue = queue[1:]
		for page := 0; ; page++ {
			node, err := r.allure.GetTestCaseTreeNode(ctx, projectID, treeID, parent, "", page, treePageSize)
			if err != nil {
				return nil, err
			}
			for _, c := range node.Children.Content {
				if c.Type == "GROUP" {
					out = append(out, c)
					queue = append(queue, c.ID)
				}
			}
			if node.Children.Last || len(node.Children.Content) == 0 {
				break
			}
		}
	}
	return out, nil
}

type folderNodeArgs struct {
	ProjectID int64  `json:"project_id"`
	TreeID    int64  `json:"tree_id"`
	NodeID    int64  `json:"node_id"`
	Name      string `json:"name"`
}

func (r *Registry) renameTestCaseFolder(ctx context.Context, args folderNodeArgs) (any, error) {
	if args.ProjectID <= 0 || args.NodeID == 0 {
		return nil, fmt.Errorf("project_id and node_id are required")
	}
	if strings.TrimSpace(args.Name) == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	treeID, err := r.resolveTreeID(ctx, args.ProjectID, args.TreeID)
	if err != nil {
		return nil, err
	}
	if err := r.requireTreeFolder(ctx, args.ProjectID, treeID, args.NodeID); err != nil {
		return nil, err
	}
	if err := r.allure.RenameTestCaseTreeGroup(ctx, args.ProjectID, args.NodeID, strings.TrimSpace(args.Name)); err != nil {
		return nil, fmt.Errorf("rename folder: %w", err)
	}
	return map[string]any{"status": "renamed", "node_id": args.NodeID, "tree_id": treeID}, nil
}

func (r *Registry) deleteTestCaseFolder(ctx context.Context, args folderNodeArgs) (any, error) {
	if args.ProjectID <= 0 || args.NodeID == 0 {
		return nil, fmt.Errorf("project_id and node_id are required")
	}
	treeID, err := r.resolveTreeID(ctx, args.ProjectID, args.TreeID)
	if err != nil {
		return nil, err
	}
	folders, err := r.treeFolders(ctx, args.ProjectID, treeID)
	if err != nil {
		return nil, fmt.Errorf("look up folder %d: %w", args.NodeID, err)
	}
	var target *allure.TestCaseTreeNodeDto
	sharing := 0
	for i := range folders {
		if folders[i].ID == args.NodeID {
			target = &folders[i]
		}
	}
	if target == nil {
		return nil, fmt.Errorf("node %d is not a folder of tree %d", args.NodeID, treeID)
	}
	for _, f := range folders {
		if f.ID != target.ID && f.CustomFieldValueID == target.CustomFieldValueID {
			sharing++
		}
	}

	// deleteGroup only unassigns the folder's test cases: the folder stays in
	// the tree, empty (confirmed live). The folder exists as long as its custom
	// field value does, so drop the value too — unless another folder (a
	// same-named one under a different parent) is backed by the same value.
	if err := r.allure.DeleteTestCaseTreeGroup(ctx, args.ProjectID, args.NodeID); err != nil {
		return nil, fmt.Errorf("delete folder: %w", err)
	}
	result := map[string]any{"status": "deleted", "node_id": args.NodeID, "tree_id": treeID}
	if target.CustomFieldValueID > 0 && sharing == 0 {
		if err := r.allure.DeleteCustomFieldValue(ctx, args.ProjectID, target.CustomFieldValueID); err != nil {
			result["status"] = "emptied"
			result["warning"] = fmt.Sprintf("test cases were unassigned, but removing the folder's value %d failed: %v", target.CustomFieldValueID, err)
		}
	} else if sharing > 0 {
		result["status"] = "emptied"
		result["warning"] = fmt.Sprintf("test cases were unassigned; the empty folder stays because its value %d also backs %d other folder(s)", target.CustomFieldValueID, sharing)
	}
	return result, nil
}

// resolveTreeLeaves finds the current leaf ids of the given test cases in a
// tree. Leaf ids are tree positions and change whenever a test case moves, so
// they are looked up fresh right before each move: the tree is filtered to
// branches containing these test cases (baseAql) and only those are walked.
// A test case can appear under several folders (multi-value custom field), so
// each maps to a list.
func (r *Registry) resolveTreeLeaves(ctx context.Context, projectID, treeID int64, testCaseIDs []int64) (map[int64][]int64, error) {
	want := make(map[int64]bool, len(testCaseIDs))
	ids := make([]string, len(testCaseIDs))
	for i, id := range testCaseIDs {
		want[id] = true
		ids[i] = strconv.FormatInt(id, 10)
	}
	aql := "id in [" + strings.Join(ids, ", ") + "]"
	found := make(map[int64][]int64)

	var walk func(parent int64, depth int) error
	walk = func(parent int64, depth int) error {
		if depth > 50 {
			return fmt.Errorf("test case tree is deeper than 50 levels")
		}
		for page := 0; ; page++ {
			node, err := r.allure.GetTestCaseTreeNode(ctx, projectID, treeID, parent, aql, page, treePageSize)
			if err != nil {
				return err
			}
			for _, c := range node.Children.Content {
				switch {
				case c.Type == "LEAF" && want[c.TestCaseID]:
					found[c.TestCaseID] = append(found[c.TestCaseID], c.ID)
				case c.Type == "GROUP" && c.Count > 0:
					if err := walk(c.ID, depth+1); err != nil {
						return err
					}
				}
			}
			if node.Children.Last || len(node.Children.Content) == 0 {
				return nil
			}
		}
	}
	if err := walk(0, 0); err != nil {
		return nil, err
	}
	return found, nil
}

type moveTestCasesToFolderArgs struct {
	ProjectID   int64   `json:"project_id"`
	TreeID      int64   `json:"tree_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	NodeID      int64   `json:"node_id"`
}

func (r *Registry) moveTestCasesToFolder(ctx context.Context, args moveTestCasesToFolderArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	treeID, err := r.resolveTreeID(ctx, args.ProjectID, args.TreeID)
	if err != nil {
		return nil, err
	}
	nodeID := args.NodeID
	if nodeID != 0 {
		if err := r.requireTreeFolder(ctx, args.ProjectID, treeID, nodeID); err != nil {
			return nil, err
		}
	} else {
		root, err := r.allure.GetTestCaseTreeNode(ctx, args.ProjectID, treeID, 0, "", 0, 1)
		if err != nil {
			return nil, fmt.Errorf("look up tree root: %w", err)
		}
		nodeID = root.ID
	}

	leaves, err := r.resolveTreeLeaves(ctx, args.ProjectID, treeID, args.TestCaseIDs)
	if err != nil {
		return nil, fmt.Errorf("locate test cases in tree %d: %w", treeID, err)
	}
	var leafIDs []int64
	notFound := make([]int64, 0)
	for _, id := range args.TestCaseIDs {
		if l, ok := leaves[id]; ok {
			leafIDs = append(leafIDs, l...)
		} else {
			notFound = append(notFound, id)
		}
	}
	if len(leafIDs) == 0 {
		return nil, fmt.Errorf("none of the test cases were found in tree %d: %v", treeID, notFound)
	}

	if err := r.allure.MoveTreeLeaves(ctx, args.ProjectID, treeID, leafIDs, nodeID); err != nil {
		return nil, fmt.Errorf("move test cases to folder: %w", err)
	}
	return map[string]any{
		"status":                  "moving",
		"tree_id":                 treeID,
		"node_id":                 nodeID,
		"count":                   len(args.TestCaseIDs) - len(notFound),
		"not_found_test_case_ids": notFound,
	}, nil
}

type createTestCaseFolderArgs struct {
	ProjectID    int64  `json:"project_id"`
	TreeID       int64  `json:"tree_id"`
	Name         string `json:"name"`
	ParentNodeID int64  `json:"parent_node_id"`
}

func (r *Registry) createTestCaseFolder(ctx context.Context, args createTestCaseFolderArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if strings.TrimSpace(args.Name) == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	treeID, err := r.resolveTreeID(ctx, args.ProjectID, args.TreeID)
	if err != nil {
		return nil, err
	}
	if args.ParentNodeID != 0 {
		if err := r.requireTreeFolder(ctx, args.ProjectID, treeID, args.ParentNodeID); err != nil {
			return nil, err
		}
	}
	group, err := r.allure.CreateTestCaseTreeGroup(ctx, args.ProjectID, treeID, args.ParentNodeID, args.Name)
	if err != nil {
		// Each tree level is one custom field; a tree with N fields is N folders
		// deep (confirmed live: "Suites" = Suite only, "Features" = Feature → Story).
		if strings.Contains(err.Error(), "tree.no-cf-in-node") {
			return nil, fmt.Errorf("tree %d has no folder level below node %d — each level of a tree is one custom field, and this tree has no field left at that depth; create the folder higher up or use a tree with more levels: %w", treeID, args.ParentNodeID, err)
		}
		return nil, fmt.Errorf("create test case folder: %w", err)
	}
	return map[string]any{
		"folder_id":             group.ID,
		"name":                  group.Name,
		"tree_id":               treeID,
		"parent_node_id":        group.ParentNodeID,
		"custom_field_id":       group.CustomFieldID,
		"custom_field_value_id": group.CustomFieldValueID,
	}, nil
}
