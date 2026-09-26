package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MimoJanra/TestOpsMCP/internal/session"
)

// registerTestPlanTools registers the test plan lifecycle (bulk_create_test_plan
// creates plans, add_test_plan_to_launch uses them) and delete_launch.
func (r *Registry) registerTestPlanTools() {
	r.register(&Tool{
		Name:        "list_test_plans",
		Description: "List a project's test plans (newest first), optionally filtered by name. Returns id, name and test case count.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"name":       map[string]any{"type": "string", "description": "Name filter (optional)"},
				"page":       map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":       map[string]any{"type": "integer", "description": "Items per page (default 20, max 100)", "default": 20},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.listTestPlans),
	})

	r.register(&Tool{
		Name:        "get_test_plan",
		Description: "Get a test plan: name, project, test case count, and the AQL it selects by (base_rql).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_plan_id": map[string]any{"type": "integer", "description": "Test plan ID"},
			},
			"required": []string{"test_plan_id"},
		},
		Handler: Typed(r.getTestPlan),
	})

	r.register(&Tool{
		Name:        "run_test_plan",
		Description: "Start a new launch from a test plan's test cases; returns the new launch_id. To add a plan to an existing launch use add_test_plan_to_launch.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_plan_id": map[string]any{"type": "integer", "description": "Test plan ID"},
				"launch_name":  map[string]any{"type": "string", "description": "Name for the new launch"},
			},
			"required": []string{"test_plan_id", "launch_name"},
		},
		Handler: Typed(r.runTestPlan),
	})

	r.register(&Tool{
		Name:        "rename_test_plan",
		Description: "Rename a test plan.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_plan_id": map[string]any{"type": "integer", "description": "Test plan ID"},
				"name":         map[string]any{"type": "string", "description": "New name"},
			},
			"required": []string{"test_plan_id", "name"},
		},
		Handler: Typed(r.renameTestPlan),
	})

	r.register(&Tool{
		Name:        "delete_test_plan",
		Description: "Delete a test plan. Its test cases and past launches are not affected. Asks the user to confirm.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_plan_id": map[string]any{"type": "integer", "description": "Test plan ID"},
			},
			"required": []string{"test_plan_id"},
		},
		Handler: Typed(r.deleteTestPlan),
	})

	r.register(&Tool{
		Name:        "delete_launch",
		Description: "Permanently delete a launch and all its test results. Cannot be undone. Asks the user to confirm.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{"type": "integer", "description": "Allure launch ID"},
			},
			"required": []string{"launch_id"},
		},
		Handler: Typed(r.deleteLaunch),
	})
}

type listTestPlansArgs struct {
	ProjectID int64  `json:"project_id"`
	Name      string `json:"name"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

func (r *Registry) listTestPlans(ctx context.Context, args listTestPlansArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.Size <= 0 {
		args.Size = 20
	}
	if args.Size > 100 {
		args.Size = 100
	}
	plans, err := r.allure.ListTestPlans(ctx, args.ProjectID, args.Name, args.Page, args.Size)
	if err != nil {
		return nil, fmt.Errorf("list test plans: %w", err)
	}
	items := make([]map[string]any, len(plans.Content))
	for i, p := range plans.Content {
		items[i] = map[string]any{
			"id":               p.ID,
			"name":             p.Name,
			"test_cases_count": p.TestCasesCount,
			"created_by":       p.CreatedBy,
			"created_date":     p.CreatedDate,
		}
	}
	return map[string]any{
		"test_plans": items,
		"page":       plans.Number,
		"size":       plans.Size,
		"total":      plans.Total,
		"is_last":    plans.Last,
	}, nil
}

type testPlanIDArgs struct {
	TestPlanID int64 `json:"test_plan_id"`
}

func (r *Registry) getTestPlan(ctx context.Context, args testPlanIDArgs) (any, error) {
	if args.TestPlanID <= 0 {
		return nil, fmt.Errorf("test_plan_id must be positive")
	}
	p, err := r.allure.GetTestPlan(ctx, args.TestPlanID)
	if err != nil {
		return nil, fmt.Errorf("get test plan: %w", err)
	}
	return map[string]any{
		"id":                 p.ID,
		"name":               p.Name,
		"project_id":         p.ProjectID,
		"test_cases_count":   p.TestCasesCount,
		"base_rql":           p.BaseRql,
		"created_by":         p.CreatedBy,
		"created_date":       p.CreatedDate,
		"last_modified_date": p.LastModifiedDate,
	}, nil
}

type runTestPlanArgs struct {
	TestPlanID int64  `json:"test_plan_id"`
	LaunchName string `json:"launch_name"`
}

func (r *Registry) runTestPlan(ctx context.Context, args runTestPlanArgs) (any, error) {
	if args.TestPlanID <= 0 {
		return nil, fmt.Errorf("test_plan_id must be positive")
	}
	if strings.TrimSpace(args.LaunchName) == "" {
		return nil, fmt.Errorf("launch_name is required")
	}
	launchID, err := r.allure.RunTestPlan(ctx, args.TestPlanID, args.LaunchName)
	if err != nil {
		return nil, fmt.Errorf("run test plan: %w", err)
	}
	return map[string]any{"status": "started", "launch_id": launchID}, nil
}

type renameTestPlanArgs struct {
	TestPlanID int64  `json:"test_plan_id"`
	Name       string `json:"name"`
}

func (r *Registry) renameTestPlan(ctx context.Context, args renameTestPlanArgs) (any, error) {
	if args.TestPlanID <= 0 {
		return nil, fmt.Errorf("test_plan_id must be positive")
	}
	if strings.TrimSpace(args.Name) == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	if err := r.allure.RenameTestPlan(ctx, args.TestPlanID, args.Name); err != nil {
		return nil, fmt.Errorf("rename test plan: %w", err)
	}
	return map[string]any{"status": "renamed"}, nil
}

func (r *Registry) deleteTestPlan(ctx context.Context, args testPlanIDArgs) (any, error) {
	if args.TestPlanID <= 0 {
		return nil, fmt.Errorf("test_plan_id must be positive")
	}
	// Check it exists first: deleting a missing plan would otherwise ask the
	// user to confirm something that isn't there.
	plan, err := r.allure.GetTestPlan(ctx, args.TestPlanID)
	if err != nil {
		return nil, fmt.Errorf("look up test plan: %w", err)
	}
	if ok, resp, err := confirmDestructive(ctx, fmt.Sprintf("Delete test plan #%d %q?", plan.ID, plan.Name)); !ok {
		return resp, err
	}
	if err := r.allure.DeleteTestPlan(ctx, args.TestPlanID); err != nil {
		return nil, fmt.Errorf("delete test plan: %w", err)
	}
	return map[string]any{"status": "deleted"}, nil
}

type deleteLaunchArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) deleteLaunch(ctx context.Context, args deleteLaunchArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	launch, err := r.requireLaunch(ctx, args.LaunchID)
	if err != nil {
		return nil, err
	}
	if ok, resp, err := confirmDestructive(ctx, fmt.Sprintf("Permanently delete launch #%d %q and all its results? This cannot be undone.", launch.ID, launch.Name)); !ok {
		return resp, err
	}
	if err := r.allure.DeleteLaunch(ctx, args.LaunchID); err != nil {
		return nil, fmt.Errorf("delete launch: %w", err)
	}
	return map[string]any{"status": "deleted"}, nil
}

// confirmDestructive asks the user via elicitation. ok is false when the
// action must not proceed; resp/err then hold what the tool should return.
func confirmDestructive(ctx context.Context, message string) (ok bool, resp any, err error) {
	elicit, has := session.ElicitFromContext(ctx)
	if !has {
		return false, nil, fmt.Errorf("this action requires user confirmation but no interactive session is available")
	}
	schema, _ := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"confirmed": map[string]any{"type": "boolean", "description": "Confirm"},
		},
	})
	result, err := elicit(ctx, message, schema)
	if err != nil {
		return false, nil, fmt.Errorf("confirmation failed: %w", err)
	}
	if result.Action != "accept" {
		return false, map[string]any{"cancelled": true, "message": "Cancelled."}, nil
	}
	return true, nil, nil
}
