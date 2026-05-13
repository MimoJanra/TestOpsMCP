package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func (r *Registry) registerLaunchTools() {
	r.register(&Tool{
		Name:        "run_allure_launch",
		Description: "Start a test launch in Allure TestOps",
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
		Handler: r.runAllureLaunch,
	})

	r.register(&Tool{
		Name:        "get_launch_status",
		Description: "Get the status of a test launch",
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
		Handler: r.getLaunchStatus,
	})

	r.register(&Tool{
		Name:        "get_launch_report",
		Description: "Get the test report for a launch",
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
		Handler: r.getLaunchReport,
	})

	r.register(&Tool{
		Name:        "close_launch",
		Description: "Close/finish an active launch",
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
		Handler: r.closeLaunch,
	})

	r.register(&Tool{
		Name:        "reopen_launch",
		Description: "Reopen a closed launch for additional test results",
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
		Handler: r.reopenLaunch,
	})

	r.register(&Tool{
		Name:        "list_launches",
		Description: "List launches in a project with pagination",
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
		Handler: r.listLaunches,
	})

	r.register(&Tool{
		Name:        "get_launch_details",
		Description: "Get comprehensive launch information",
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
		Handler: r.getLaunchDetails,
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
		Handler: r.getLaunchEnvironment,
	})

	r.register(&Tool{
		Name:        "update_launch_environment",
		Description: "Update launch environment variables",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"environment": map[string]any{
					"type":        "object",
					"description": "Environment variables as key-value pairs",
				},
			},
			"required": []string{"launch_id", "environment"},
		},
		Handler: r.updateLaunchEnvironment,
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
		Handler: r.copyLaunch,
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
		Handler: r.mergeLaunches,
	})

	r.register(&Tool{
		Name:        "add_test_cases_to_launch",
		Description: "Add test cases to a launch",
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
		Handler: r.addTestCasesToLaunch,
	})

	r.register(&Tool{
		Name:        "add_test_plan_to_launch",
		Description: "Add a test plan to a launch",
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
		Handler: r.addTestPlanToLaunch,
	})

	r.register(&Tool{
		Name:        "get_launch_defects",
		Description: "Get defects linked to a launch",
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
		Handler: r.getLaunchDefects,
	})
}

func (r *Registry) runAllureLaunch(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID  int64  `json:"project_id"`
		LaunchName string `json:"launch_name"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	if params.LaunchName == "" {
		return nil, fmt.Errorf("launch_name is required")
	}

	r.logger.Info("starting Allure launch", map[string]any{
		"project_id":  params.ProjectID,
		"launch_name": params.LaunchName,
	})

	launch, err := r.allure.CreateLaunch(ctx, params.ProjectID, params.LaunchName)
	if err != nil {
		r.logger.Error("create launch", err, map[string]any{
			"project_id": params.ProjectID,
		})
		return nil, fmt.Errorf("create launch: %w", err)
	}

	r.logger.Info("launch created", map[string]any{"launch_id": launch.ID})

	return map[string]any{
		"launch_id": launch.ID,
		"status":    "started",
	}, nil
}

func (r *Registry) getLaunchStatus(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch status", map[string]any{"launch_id": params.LaunchID})

	status, err := r.allure.GetLaunchStatus(ctx, params.LaunchID)
	if err != nil {
		r.logger.Error("get launch status", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("get launch status: %w", err)
	}

	return map[string]any{"status": status}, nil
}

func (r *Registry) getLaunchReport(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch report", map[string]any{"launch_id": params.LaunchID})

	stats, err := r.allure.GetLaunchStatistics(ctx, params.LaunchID)
	if err != nil {
		r.logger.Error("get launch statistics", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("get launch statistics: %w", err)
	}

	return map[string]any{
		"total":  stats.Total,
		"passed": stats.Passed,
		"failed": stats.Failed,
		"broken": stats.Broken,
	}, nil
}

func (r *Registry) closeLaunch(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("closing launch", map[string]any{"launch_id": params.LaunchID})

	if err := r.allure.CloseLaunch(ctx, params.LaunchID); err != nil {
		r.logger.Error("close launch", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("close launch: %w", err)
	}

	return map[string]any{"status": "closed"}, nil
}

func (r *Registry) reopenLaunch(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("reopening launch", map[string]any{"launch_id": params.LaunchID})

	if err := r.allure.ReopenLaunch(ctx, params.LaunchID); err != nil {
		r.logger.Error("reopen launch", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("reopen launch: %w", err)
	}

	return map[string]any{"status": "reopened"}, nil
}

func (r *Registry) listLaunches(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID int64 `json:"project_id"`
		Page      int   `json:"page"`
		Size      int   `json:"size"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	if params.Size == 0 {
		params.Size = 10
	}
	if params.Size > 100 {
		params.Size = 100
	}

	r.logger.Info("listing launches", map[string]any{
		"project_id": params.ProjectID,
		"page":       params.Page,
		"size":       params.Size,
	})

	launches, err := r.allure.ListLaunches(ctx, params.ProjectID, params.Page, params.Size)
	if err != nil {
		r.logger.Error("list launches", err, map[string]any{"project_id": params.ProjectID})
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

func (r *Registry) getLaunchDetails(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch details", map[string]any{"launch_id": params.LaunchID})

	details, err := r.allure.GetLaunchDetails(ctx, params.LaunchID)
	if err != nil {
		r.logger.Error("get launch details", err, map[string]any{"launch_id": params.LaunchID})
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
		"status":         details.Status,
		"project_id":     details.ProjectID,
		"start_time":     details.StartTime,
		"end_time":       details.EndTime,
		"environment":    details.Environment,
		"tags":           tags,
		"description":    details.Description,
		"report_web_url": details.ReportWebUrl,
	}, nil
}

func (r *Registry) getLaunchEnvironment(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch environment", map[string]any{"launch_id": params.LaunchID})

	env, err := r.allure.GetLaunchEnvironment(ctx, params.LaunchID)
	if err != nil {
		r.logger.Error("get launch environment", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("get launch environment: %w", err)
	}

	return map[string]any{"environment": env}, nil
}

func (r *Registry) updateLaunchEnvironment(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID    int64          `json:"launch_id"`
		Environment map[string]any `json:"environment"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	if len(params.Environment) == 0 {
		return nil, fmt.Errorf("environment must not be empty")
	}

	r.logger.Info("updating launch environment", map[string]any{"launch_id": params.LaunchID})

	if err := r.allure.UpdateLaunchEnvironment(ctx, params.LaunchID, params.Environment); err != nil {
		r.logger.Error("update launch environment", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("update launch environment: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

func (r *Registry) copyLaunch(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("copying launch", map[string]any{"launch_id": params.LaunchID})

	launch, err := r.allure.CopyLaunch(ctx, params.LaunchID)
	if err != nil {
		r.logger.Error("copy launch", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("copy launch: %w", err)
	}

	return map[string]any{
		"launch_id": launch.ID,
		"name":      launch.Name,
		"status":    "copied",
	}, nil
}

func (r *Registry) mergeLaunches(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchIDs  []int64 `json:"launch_ids"`
		LaunchName string  `json:"launch_name"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if len(params.LaunchIDs) == 0 {
		return nil, fmt.Errorf("launch_ids must not be empty")
	}

	if params.LaunchName == "" {
		return nil, fmt.Errorf("launch_name is required")
	}

	r.logger.Info("merging launches", map[string]any{
		"count": len(params.LaunchIDs),
		"name":  params.LaunchName,
	})

	launchID, err := r.allure.MergeLaunches(ctx, params.LaunchIDs, params.LaunchName)
	if err != nil {
		r.logger.Error("merge launches", err, map[string]any{"count": len(params.LaunchIDs)})
		return nil, fmt.Errorf("merge launches: %w", err)
	}

	return map[string]any{
		"merged_launch_id": launchID,
		"status":           "merged",
	}, nil
}

func (r *Registry) addTestCasesToLaunch(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID    int64    `json:"launch_id"`
		ProjectID   int64    `json:"project_id"`
		TestCaseIDs []int64  `json:"test_case_ids"`
		Assignees   []string `json:"assignees"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(params.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}

	r.logger.Info("adding test cases to launch", map[string]any{"launch_id": params.LaunchID, "count": len(params.TestCaseIDs)})

	if err := r.allure.AddTestCasesToLaunch(ctx, params.LaunchID, params.ProjectID, params.TestCaseIDs, params.Assignees); err != nil {
		r.logger.Error("add test cases to launch", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("add test cases: %w", err)
	}

	return map[string]any{"status": "success", "count": len(params.TestCaseIDs)}, nil
}

func (r *Registry) addTestPlanToLaunch(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID   int64 `json:"launch_id"`
		TestPlanID int64 `json:"test_plan_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if params.TestPlanID <= 0 {
		return nil, fmt.Errorf("test_plan_id must be positive")
	}

	r.logger.Info("adding test plan to launch", map[string]any{"launch_id": params.LaunchID, "test_plan_id": params.TestPlanID})

	if err := r.allure.AddTestPlanToLaunch(ctx, params.LaunchID, params.TestPlanID); err != nil {
		r.logger.Error("add test plan to launch", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("add test plan: %w", err)
	}

	return map[string]any{"status": "success"}, nil
}

func (r *Registry) getLaunchDefects(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
		Page     int   `json:"page"`
		Size     int   `json:"size"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	if params.Size == 0 {
		params.Size = 10
	}
	if params.Size > 100 {
		params.Size = 100
	}

	r.logger.Info("fetching launch defects", map[string]any{
		"launch_id": params.LaunchID,
		"page":      params.Page,
		"size":      params.Size,
	})

	defects, err := r.allure.GetLaunchDefects(ctx, params.LaunchID, params.Page, params.Size)
	if err != nil {
		r.logger.Error("get launch defects", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("get launch defects: %w", err)
	}

	return defects, nil
}
