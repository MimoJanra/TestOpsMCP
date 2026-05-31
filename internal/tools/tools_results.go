package tools

import (
	"context"
	"fmt"
)

func (r *Registry) registerResultTools() {
	r.register(&Tool{
		Name:        "list_test_results",
		Description: "List test results in a launch with optional status filter",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "Filter by status (PASSED, FAILED, BROKEN, SKIPPED)",
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
		Handler: Typed(r.listTestResults),
	})

	r.register(&Tool{
		Name:        "get_test_result",
		Description: "Get detailed information about a single test result",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_result_id": map[string]any{
					"type":        "integer",
					"description": "Allure test result ID",
				},
			},
			"required": []string{"test_result_id"},
		},
		Handler: Typed(r.getTestResult),
	})

	r.register(&Tool{
		Name:        "assign_test_result",
		Description: "Assign a test result to a team member",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_result_id": map[string]any{
					"type":        "integer",
					"description": "Allure test result ID",
				},
				"username": map[string]any{
					"type":        "string",
					"description": "Username to assign to",
				},
			},
			"required": []string{"test_result_id", "username"},
		},
		Handler: Typed(r.assignTestResult),
	})

	r.register(&Tool{
		Name:        "mute_test_result",
		Description: "Mute a failing test result (mark as known issue)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_result_id": map[string]any{
					"type":        "integer",
					"description": "Allure test result ID",
				},
				"reason": map[string]any{
					"type":        "string",
					"description": "Reason for muting (optional)",
				},
			},
			"required": []string{"test_result_id"},
		},
		Handler: Typed(r.muteTestResult),
	})

	r.register(&Tool{
		Name:        "resolve_test_result",
		Description: "Resolve a test result (mark as resolved/fixed)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_result_id": map[string]any{
					"type":        "integer",
					"description": "Allure test result ID",
				},
			},
			"required": []string{"test_result_id"},
		},
		Handler: Typed(r.resolveTestResult),
	})

	r.register(&Tool{
		Name:        "unmute_test_result",
		Description: "Unmute a test result",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_result_id": map[string]any{
					"type":        "integer",
					"description": "Allure test result ID",
				},
			},
			"required": []string{"test_result_id"},
		},
		Handler: Typed(r.unmuteTestResult),
	})
}

type listTestResultsArgs struct {
	LaunchID int64  `json:"launch_id"`
	Status   string `json:"status"`
	Page     int    `json:"page"`
	Size     int    `json:"size"`
}

func (r *Registry) listTestResults(ctx context.Context, args listTestResultsArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	if args.Size <= 0 {
		args.Size = 10
	}
	if args.Size > 100 {
		args.Size = 100
	}

	r.logger.Info("listing test results", map[string]any{
		"launch_id": args.LaunchID,
		"status":    args.Status,
		"page":      args.Page,
		"size":      args.Size,
	})

	results, err := r.allure.ListTestResults(ctx, args.LaunchID, args.Status, args.Page, args.Size)
	if err != nil {
		r.logger.Error("list test results", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("list test results: %w", err)
	}

	items := make([]map[string]any, len(results.Content))
	for i, result := range results.Content {
		items[i] = map[string]any{
			"id":           result.ID,
			"name":         result.Name,
			"status":       result.Status,
			"launch_id":    result.LaunchID,
			"test_case_id": result.TestCaseID,
			"start_time":   result.StartTime,
			"end_time":     result.EndTime,
			"duration":     result.Duration,
			"assignee":     result.Assignee,
			"muted":        result.Muted,
			"flaky":        result.Flaky,
		}
	}

	return map[string]any{
		"test_results": items,
		"page":         results.Number,
		"size":         results.Size,
		"total":        results.Total,
		"is_last":      results.Last,
	}, nil
}

type getTestResultArgs struct {
	TestResultID int64 `json:"test_result_id"`
}

func (r *Registry) getTestResult(ctx context.Context, args getTestResultArgs) (any, error) {
	if args.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}

	r.logger.Info("fetching test result", map[string]any{"test_result_id": args.TestResultID})

	result, err := r.allure.GetTestResult(ctx, args.TestResultID)
	if err != nil {
		r.logger.Error("get test result", err, map[string]any{"test_result_id": args.TestResultID})
		return nil, fmt.Errorf("get test result: %w", err)
	}

	paramsList := make([]map[string]any, len(result.Parameters))
	for i, p := range result.Parameters {
		paramsList[i] = map[string]any{
			"name":     p.Name,
			"value":    p.Value,
			"hidden":   p.Hidden,
			"excluded": p.Excluded,
		}
	}

	tagsList := make([]map[string]any, len(result.Tags))
	for i, t := range result.Tags {
		tagsList[i] = map[string]any{"id": t.ID, "name": t.Name}
	}

	return map[string]any{
		"id":           result.ID,
		"name":         result.Name,
		"status":       result.Status,
		"launch_id":    result.LaunchID,
		"test_case_id": result.TestCaseID,
		"start_time":   result.StartTime,
		"end_time":     result.EndTime,
		"duration":     result.Duration,
		"full_name":    result.FullName,
		"description":  result.Description,
		"message":      result.Message,
		"trace":        result.Trace,
		"parameters":   paramsList,
		"assignee":     result.Assignee,
		"muted":        result.Muted,
		"flaky":        result.Flaky,
		"known":        result.Known,
		"tags":         tagsList,
	}, nil
}

type assignTestResultArgs struct {
	TestResultID int64  `json:"test_result_id"`
	Username     string `json:"username"`
}

func (r *Registry) assignTestResult(ctx context.Context, args assignTestResultArgs) (any, error) {
	if args.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}
	if args.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	r.logger.Info("assigning test result", map[string]any{
		"test_result_id": args.TestResultID,
		"username":       args.Username,
	})

	if err := r.allure.AssignTestResult(ctx, args.TestResultID, args.Username); err != nil {
		r.logger.Error("assign test result", err, map[string]any{"test_result_id": args.TestResultID})
		return nil, fmt.Errorf("assign test result: %w", err)
	}

	return map[string]any{"status": "assigned"}, nil
}

type muteTestResultArgs struct {
	TestResultID int64  `json:"test_result_id"`
	Reason       string `json:"reason"`
}

func (r *Registry) muteTestResult(ctx context.Context, args muteTestResultArgs) (any, error) {
	if args.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}

	r.logger.Info("muting test result", map[string]any{
		"test_result_id": args.TestResultID,
		"reason":         args.Reason,
	})

	if err := r.allure.MuteTestResult(ctx, args.TestResultID, args.Reason); err != nil {
		r.logger.Error("mute test result", err, map[string]any{"test_result_id": args.TestResultID})
		return nil, fmt.Errorf("mute test result: %w", err)
	}

	return map[string]any{"status": "muted"}, nil
}

type resolveTestResultArgs struct {
	TestResultID int64 `json:"test_result_id"`
}

func (r *Registry) resolveTestResult(ctx context.Context, args resolveTestResultArgs) (any, error) {
	if args.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}

	r.logger.Info("resolving test result", map[string]any{"test_result_id": args.TestResultID})

	if err := r.allure.ResolveTestResult(ctx, args.TestResultID); err != nil {
		r.logger.Error("resolve test result", err, map[string]any{"test_result_id": args.TestResultID})
		return nil, fmt.Errorf("resolve test result: %w", err)
	}

	return map[string]any{"status": "resolved"}, nil
}

type unmuteTestResultArgs struct {
	TestResultID int64 `json:"test_result_id"`
}

func (r *Registry) unmuteTestResult(ctx context.Context, args unmuteTestResultArgs) (any, error) {
	if args.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}

	r.logger.Info("unmuting test result", map[string]any{"test_result_id": args.TestResultID})

	if err := r.allure.UnmuteTestResult(ctx, args.TestResultID); err != nil {
		r.logger.Error("unmute test result", err, map[string]any{"test_result_id": args.TestResultID})
		return nil, fmt.Errorf("unmute test result: %w", err)
	}

	return map[string]any{"status": "unmuted"}, nil
}
