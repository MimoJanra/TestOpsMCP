package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/session"
	"github.com/MimoJanra/TestOpsMCP/internal/tasks"
)

func (r *Registry) registerLaunchTools() {
	r.register(&Tool{
		Name:        "run_allure_launch",
		Description: "Create and start a new test launch in a project. Long-running — returns a task_id, use get_task_status to track progress. Check the result with get_launch_status once done.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"launch_name": map[string]any{
					"type":        "string",
					"description": "Name of the launch",
				},
			},
			"required": []string{"project_id", "launch_name"},
		},
		Handler: Typed(r.runAllureLaunch),
	})

	r.register(&Tool{
		Name:        "get_launch_status",
		Description: "Get the current status of a launch (running, done, closed, etc.). Use to poll after run_allure_launch or to check whether a launch is still active.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: Typed(r.getLaunchStatus),
	})

	r.register(&Tool{
		Name:        "get_launch_report",
		Description: "Get a summary report for a launch: total tests run, passed/failed/broken/skipped counts, and duration.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: Typed(r.getLaunchReport),
	})

	r.register(&Tool{
		Name:        "close_launch",
		Description: "Close an active launch to finalize its results. Closed launches are locked and their metrics count toward project statistics.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: Typed(r.closeLaunch),
	})

	r.register(&Tool{
		Name:        "reopen_launch",
		Description: "Reopen a closed launch to allow adding more test results. Use when a CI run was closed before all results were submitted.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: Typed(r.reopenLaunch),
	})

	r.register(&Tool{
		Name:        "list_launches",
		Description: "List test launches in a project, most recent first. Returns launch ID, name, status, and result counts. Use get_launch_details for full info on a single launch.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "Page number (0-based)",
					"default":     0,
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "Items per page",
					"default":     10,
				},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.listLaunches),
	})

	r.register(&Tool{
		Name:        "get_launch_details",
		Description: "Get full details of a launch: status, pass/fail/broken/skipped breakdown, environment, timing, and linked test plan.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: Typed(r.getLaunchDetails),
	})

	r.register(&Tool{
		Name:        "get_launch_environment",
		Description: "Get launch environment variables",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: Typed(r.getLaunchEnvironment),
	})

	r.register(&Tool{
		Name:        "copy_launch",
		Description: "Copy/duplicate a launch",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID to copy",
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: Typed(r.copyLaunch),
	})

	r.register(&Tool{
		Name:        "merge_launches",
		Description: "Merge multiple launches into a single launch",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "IDs of launches to merge",
				},
				"launch_name": map[string]any{
					"type":        "string",
					"description": "Name for the merged launch",
				},
			},
			"required": []string{"launch_ids", "launch_name"},
		},
		Handler: Typed(r.mergeLaunches),
	})

	r.register(&Tool{
		Name: "add_test_cases_to_launch",
		Description: "Add test cases to an existing launch so they appear in its scope and can receive results. " +
			"Only works for test cases with automation_status \"manual\" — cases with automation_status \"automated\" require a CI job assignment in TestOps " +
			"and will fail with a \"no-job-assigned\" error if added this way. Check automation_status via get_test_case or search_test_cases first.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"test_case_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test case IDs to add",
				},
				"assignees": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
					"description": "Usernames to assign test cases to (optional)",
				},
			},
			"required": []string{"launch_id", "project_id", "test_case_ids"},
		},
		Handler: Typed(r.addTestCasesToLaunch),
	})

	r.register(&Tool{
		Name:        "add_test_plan_to_launch",
		Description: "Attach a test plan to a launch, adding all of the plan's test cases into the launch scope.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"test_plan_id": map[string]any{
					"type":        "integer",
					"description": "Test plan ID to add",
				},
			},
			"required": []string{"launch_id", "test_plan_id"},
		},
		Handler: Typed(r.addTestPlanToLaunch),
	})

	r.register(&Tool{
		Name: "remove_test_cases_from_launch",
		Description: "Remove test cases from a launch. Resolves each test_case_id to its test result(s) in the launch (including retries) and removes them. " +
			"mode=\"hide\" (default) keeps the data but excludes it from the report (safer, reversible); " +
			"mode=\"delete\" permanently deletes the test results (irreversible — including execution history). " +
			"Use this to trim a launch that has too many cases or duplicates, e.g. after add_test_cases_to_launch.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID to remove test cases from",
				},
				"test_case_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test case IDs to remove from the launch",
				},
				"mode": map[string]any{
					"type":        "string",
					"enum":        []string{"hide", "delete"},
					"description": "\"hide\" (default) excludes results from the report but keeps the data; \"delete\" permanently deletes them",
					"default":     "hide",
				},
			},
			"required": []string{"launch_id", "test_case_ids"},
		},
		// Can permanently delete data when mode=delete, so flag as destructive.
		Annotations: map[string]any{
			"readOnlyHint":    false,
			"destructiveHint": true,
		},
		Handler: Typed(r.removeTestCasesFromLaunch),
	})

	r.register(&Tool{
		Name:        "get_launch_defects",
		Description: "Get defects (bugs/issues) linked to test results within a launch — useful for a post-run defect summary.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "Page number (0-based)",
					"default":     0,
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "Items per page",
					"default":     10,
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: Typed(r.getLaunchDefects),
	})
}

type runAllureLaunchArgs struct {
	ProjectID  int64  `json:"project_id"`
	LaunchName string `json:"launch_name"`
}

func (r *Registry) runAllureLaunch(ctx context.Context, args runAllureLaunchArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.LaunchName == "" {
		return nil, fmt.Errorf("launch_name is required")
	}

	r.logger.Info("starting Allure launch async", map[string]any{
		"project_id":  args.ProjectID,
		"launch_name": args.LaunchName,
	})

	task, taskCtx := r.taskStore.Create("run_allure_launch", ctx)
	r.taskStore.Run(task.ID, taskCtx, func(taskCtx context.Context) {
		launch, err := r.allure.CreateLaunch(taskCtx, args.ProjectID, args.LaunchName)
		if err != nil {
			r.logger.Error("create launch", err, map[string]any{"project_id": args.ProjectID})
			r.taskStore.Update(task.ID, tasks.StatusFailed, "", nil, err)
			return
		}
		r.logger.Info("launch created", map[string]any{"launch_id": launch.ID})
		r.taskStore.Update(task.ID, tasks.StatusSucceeded, "", map[string]any{
			"launch_id": launch.ID,
			"status":    "started",
		}, nil)
	})

	return map[string]any{
		"task_id": task.ID,
		"message": "Launch creation started. Use get_task_status to track progress.",
	}, nil
}

type getLaunchStatusArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) getLaunchStatus(ctx context.Context, args getLaunchStatusArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch status", map[string]any{"launch_id": args.LaunchID})

	status, err := r.allure.GetLaunchStatus(ctx, args.LaunchID)
	if err != nil {
		r.logger.Error("get launch status", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("get launch status: %w", err)
	}

	return map[string]any{"status": status}, nil
}

type getLaunchReportArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) getLaunchReport(ctx context.Context, args getLaunchReportArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch report", map[string]any{"launch_id": args.LaunchID})

	stats, err := r.allure.GetLaunchStatistics(ctx, args.LaunchID)
	if err != nil {
		r.logger.Error("get launch statistics", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("get launch statistics: %w", err)
	}

	return map[string]any{
		"total":  stats.Total,
		"passed": stats.Passed,
		"failed": stats.Failed,
		"broken": stats.Broken,
	}, nil
}

type closeLaunchArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) closeLaunch(ctx context.Context, args closeLaunchArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("closing launch", map[string]any{"launch_id": args.LaunchID})

	if err := r.allure.CloseLaunch(ctx, args.LaunchID); err != nil {
		r.logger.Error("close launch", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("close launch: %w", err)
	}

	return map[string]any{"status": "closed"}, nil
}

type reopenLaunchArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) reopenLaunch(ctx context.Context, args reopenLaunchArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("reopening launch", map[string]any{"launch_id": args.LaunchID})

	if err := r.allure.ReopenLaunch(ctx, args.LaunchID); err != nil {
		r.logger.Error("reopen launch", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("reopen launch: %w", err)
	}

	return map[string]any{"status": "reopened"}, nil
}

type listLaunchesArgs struct {
	ProjectID int64 `json:"project_id"`
	Page      int   `json:"page"`
	Size      int   `json:"size"`
}

func (r *Registry) listLaunches(ctx context.Context, args listLaunchesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	if args.Size <= 0 {
		args.Size = 10
	}
	if args.Size > 100 {
		args.Size = 100
	}

	r.logger.Info("listing launches", map[string]any{
		"project_id": args.ProjectID,
		"page":       args.Page,
		"size":       args.Size,
	})

	launches, err := r.allure.ListLaunches(ctx, args.ProjectID, args.Page, args.Size)
	if err != nil {
		r.logger.Error("list launches", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("list launches: %w", err)
	}

	items := make([]map[string]any, len(launches.Content))
	for i, launch := range launches.Content {
		tags := make([]map[string]any, len(launch.Tags))
		for j, tag := range launch.Tags {
			tags[j] = map[string]any{
				"id":   tag.ID,
				"name": tag.Name,
			}
		}
		items[i] = map[string]any{
			"id":          launch.ID,
			"name":        launch.Name,
			"status":      launch.Status,
			"project_id":  launch.ProjectID,
			"start_time":  launch.StartTime,
			"end_time":    launch.EndTime,
			"environment": launch.Environment,
			"tags":        tags,
		}
	}

	return map[string]any{
		"launches": items,
		"page":     launches.Number,
		"size":     launches.Size,
		"total":    launches.Total,
		"is_last":  launches.Last,
	}, nil
}

type getLaunchDetailsArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) getLaunchDetails(ctx context.Context, args getLaunchDetailsArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch details", map[string]any{"launch_id": args.LaunchID})

	details, err := r.allure.GetLaunchDetails(ctx, args.LaunchID)
	if err != nil {
		r.logger.Error("get launch details", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("get launch details: %w", err)
	}

	tags := make([]map[string]any, len(details.Tags))
	for i, tag := range details.Tags {
		tags[i] = map[string]any{
			"id":   tag.ID,
			"name": tag.Name,
		}
	}

	return map[string]any{
		"id":             details.ID,
		"uuid":           details.UUID,
		"name":           details.Name,
		"status":         normalizeLaunchStatus(details.Status),
		"project_id":     details.ProjectID,
		"start_time":     details.StartTime,
		"end_time":       details.EndTime,
		"environment":    details.Environment,
		"tags":           tags,
		"description":    details.Description,
		"report_web_url": details.ReportWebUrl,
	}, nil
}

type getLaunchEnvironmentArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) getLaunchEnvironment(ctx context.Context, args getLaunchEnvironmentArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch environment", map[string]any{"launch_id": args.LaunchID})

	env, err := r.allure.GetLaunchEnvironment(ctx, args.LaunchID)
	if err != nil {
		r.logger.Error("get launch environment", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("get launch environment: %w", err)
	}

	return map[string]any{"environment": env}, nil
}

type copyLaunchArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) copyLaunch(ctx context.Context, args copyLaunchArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("copying launch async", map[string]any{"launch_id": args.LaunchID})

	task, taskCtx := r.taskStore.Create("copy_launch", ctx)
	r.taskStore.Run(task.ID, taskCtx, func(taskCtx context.Context) {
		if err := r.allure.CopyLaunch(taskCtx, args.LaunchID); err != nil {
			r.logger.Error("copy launch", err, map[string]any{"launch_id": args.LaunchID})
			r.taskStore.Update(task.ID, tasks.StatusFailed, "", nil, err)
			return
		}
		// The API returns 202 Accepted with no body — the new launch's ID isn't
		// known here; the caller can find it via list_launches.
		r.taskStore.Update(task.ID, tasks.StatusSucceeded, "", map[string]any{
			"status":  "copied",
			"message": "Copy accepted. The new launch isn't returned by this endpoint — use list_launches to find it.",
		}, nil)
	})

	return map[string]any{
		"task_id": task.ID,
		"message": "Launch copy started. Use get_task_status to track progress.",
	}, nil
}

type mergeLaunchesArgs struct {
	LaunchIDs  []int64 `json:"launch_ids"`
	LaunchName string  `json:"launch_name"`
}

func (r *Registry) mergeLaunches(ctx context.Context, args mergeLaunchesArgs) (any, error) {
	if len(args.LaunchIDs) == 0 {
		return nil, fmt.Errorf("launch_ids must not be empty")
	}
	if args.LaunchName == "" {
		return nil, fmt.Errorf("launch_name is required")
	}

	r.logger.Info("merging launches async", map[string]any{
		"count": len(args.LaunchIDs),
		"name":  args.LaunchName,
	})

	task, taskCtx := r.taskStore.Create("merge_launches", ctx)
	r.taskStore.Run(task.ID, taskCtx, func(taskCtx context.Context) {
		launchID, err := r.allure.MergeLaunches(taskCtx, args.LaunchIDs, args.LaunchName)
		if err != nil {
			r.logger.Error("merge launches", err, map[string]any{"count": len(args.LaunchIDs)})
			r.taskStore.Update(task.ID, tasks.StatusFailed, "", nil, err)
			return
		}
		r.taskStore.Update(task.ID, tasks.StatusSucceeded, "", map[string]any{
			"merged_launch_id": launchID,
			"status":           "merged",
		}, nil)
	})

	return map[string]any{
		"task_id": task.ID,
		"message": "Launch merge started. Use get_task_status to track progress.",
	}, nil
}

type addTestCasesToLaunchArgs struct {
	LaunchID    int64    `json:"launch_id"`
	ProjectID   int64    `json:"project_id"`
	TestCaseIDs []int64  `json:"test_case_ids"`
	Assignees   []string `json:"assignees"`
}

func (r *Registry) addTestCasesToLaunch(ctx context.Context, args addTestCasesToLaunchArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}

	r.logger.Info("adding test cases to launch", map[string]any{"launch_id": args.LaunchID, "count": len(args.TestCaseIDs)})

	if err := r.allure.AddTestCasesToLaunch(ctx, args.LaunchID, args.ProjectID, args.TestCaseIDs, args.Assignees); err != nil {
		r.logger.Error("add test cases to launch", err, map[string]any{"launch_id": args.LaunchID})
		var apiErr *allure.APIError
		isNoJobAssigned := errors.As(err, &apiErr) &&
			(apiErr.Code == "no-job-assigned" || strings.Contains(apiErr.Message, "no-job-assigned"))
		if isNoJobAssigned {
			return nil, fmt.Errorf("add test cases: one or more of these test cases has automation_status \"automated\" and TestOps requires a CI job assignment before it can be added to a launch this way; either add only cases with automation_status \"manual\", or assign an automation job to the automated cases in TestOps first: %w", err)
		}
		return nil, fmt.Errorf("add test cases: %w", err)
	}

	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type addTestPlanToLaunchArgs struct {
	LaunchID   int64 `json:"launch_id"`
	TestPlanID int64 `json:"test_plan_id"`
}

func (r *Registry) addTestPlanToLaunch(ctx context.Context, args addTestPlanToLaunchArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if args.TestPlanID <= 0 {
		return nil, fmt.Errorf("test_plan_id must be positive")
	}

	r.logger.Info("adding test plan to launch", map[string]any{"launch_id": args.LaunchID, "test_plan_id": args.TestPlanID})

	if err := r.allure.AddTestPlanToLaunch(ctx, args.LaunchID, args.TestPlanID); err != nil {
		r.logger.Error("add test plan to launch", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("add test plan: %w", err)
	}

	return map[string]any{"status": "success"}, nil
}

// remove_test_cases_from_launch scans the launch's test results client-side to
// map test_case_id → test result IDs, because the Allure API offers no
// "remove test case from launch" endpoint and no bulk test-result delete.
const (
	removeFromLaunchPageSize = 100
	removeFromLaunchMaxPages = 200 // scan at most 20000 results per launch
)

type removeTestCasesFromLaunchArgs struct {
	LaunchID    int64   `json:"launch_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	Mode        string  `json:"mode"`
}

func (r *Registry) removeTestCasesFromLaunch(ctx context.Context, args removeTestCasesFromLaunchArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}

	mode := args.Mode
	if mode == "" {
		mode = "hide"
	}
	if mode != "hide" && mode != "delete" {
		return nil, fmt.Errorf("mode must be \"hide\" or \"delete\"")
	}

	// Build a lookup set of the requested test case IDs.
	wanted := make(map[int64]bool, len(args.TestCaseIDs))
	for _, id := range args.TestCaseIDs {
		if id > 0 {
			wanted[id] = true
		}
	}
	if len(wanted) == 0 {
		return nil, fmt.Errorf("test_case_ids must contain a positive ID")
	}

	r.logger.Info("removing test cases from launch", map[string]any{
		"launch_id": args.LaunchID,
		"count":     len(wanted),
		"mode":      mode,
	})

	// Scan the launch's results and collect the result IDs that belong to the
	// requested test cases. A test case may have several results (retries).
	var resultIDs []int64
	found := make(map[int64]bool) // test case IDs with at least one result
	truncated := false

	for page := 0; page < removeFromLaunchMaxPages; page++ {
		resp, err := r.allure.ListTestResults(ctx, args.LaunchID, "", page, removeFromLaunchPageSize)
		if err != nil {
			r.logger.Error("remove test cases from launch: list results", err, map[string]any{"launch_id": args.LaunchID})
			return nil, fmt.Errorf("list test results: %w", err)
		}

		for _, tr := range resp.Content {
			if wanted[tr.TestCaseID] {
				resultIDs = append(resultIDs, tr.ID)
				found[tr.TestCaseID] = true
			}
		}

		if resp.Last || len(resp.Content) == 0 {
			break
		}
		if page == removeFromLaunchMaxPages-1 {
			truncated = true
		}
	}

	// Requested test case IDs that had no result in this launch.
	notFound := make([]int64, 0)
	for _, id := range args.TestCaseIDs {
		if id > 0 && !found[id] {
			notFound = append(notFound, id)
		}
	}

	if len(resultIDs) == 0 {
		return map[string]any{
			"launch_id":               args.LaunchID,
			"mode":                    mode,
			"removed_count":           0,
			"removed_result_ids":      []int64{},
			"not_found_test_case_ids": notFound,
			"truncated":               truncated,
			"message":                 "no matching test results found in the launch",
		}, nil
	}

	// Apply the removal.
	succeeded := make([]int64, 0, len(resultIDs))
	var failed []map[string]any

	switch mode {
	case "hide":
		if err := r.allure.BulkHideTestResults(ctx, args.LaunchID, resultIDs); err != nil {
			r.logger.Error("remove test cases from launch: hide", err, map[string]any{"launch_id": args.LaunchID})
			return nil, fmt.Errorf("hide test results: %w", err)
		}
		succeeded = resultIDs
	case "delete":
		elicit, ok := session.ElicitFromContext(ctx)
		if !ok {
			return nil, fmt.Errorf("mode=\"delete\" requires user confirmation but no interactive session is available; use mode=\"hide\" instead")
		}
		schema, _ := json.Marshal(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"confirmed": map[string]any{"type": "boolean", "description": "Confirm permanent deletion"},
			},
		})
		confirmResult, err := elicit(ctx, fmt.Sprintf("Permanently delete %d test result(s) from launch #%d? This cannot be undone.", len(resultIDs), args.LaunchID), schema)
		if err != nil {
			return nil, fmt.Errorf("confirmation failed: %w", err)
		}
		if confirmResult.Action != "accept" {
			return map[string]any{"cancelled": true, "message": "Deletion cancelled."}, nil
		}

		for _, id := range resultIDs {
			if err := r.allure.DeleteTestResult(ctx, id); err != nil {
				failed = append(failed, map[string]any{"test_result_id": id, "error": err.Error()})
				continue
			}
			succeeded = append(succeeded, id)
		}
	}

	result := map[string]any{
		"launch_id":               args.LaunchID,
		"mode":                    mode,
		"removed_count":           len(succeeded),
		"removed_result_ids":      succeeded,
		"not_found_test_case_ids": notFound,
		"truncated":               truncated,
	}
	if len(failed) > 0 {
		result["failed"] = failed
	}
	return result, nil
}

type getLaunchDefectsArgs struct {
	LaunchID int64 `json:"launch_id"`
	Page     int   `json:"page"`
	Size     int   `json:"size"`
}

func (r *Registry) getLaunchDefects(ctx context.Context, args getLaunchDefectsArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	if args.Size <= 0 {
		args.Size = 10
	}
	if args.Size > 100 {
		args.Size = 100
	}

	r.logger.Info("fetching launch defects", map[string]any{
		"launch_id": args.LaunchID,
		"page":      args.Page,
		"size":      args.Size,
	})

	defects, err := r.allure.GetLaunchDefects(ctx, args.LaunchID, args.Page, args.Size)
	if err != nil {
		r.logger.Error("get launch defects", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("get launch defects: %w", err)
	}

	return defects, nil
}
