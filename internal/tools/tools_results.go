package tools

import (
	"context"
	"encoding/json"
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
		Handler: r.listTestResults,
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
		Handler: r.getTestResult,
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
		Handler: r.assignTestResult,
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
		Handler: r.muteTestResult,
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
		Handler: r.resolveTestResult,
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
		Handler: r.unmuteTestResult,
	})
}

func (r *Registry) listTestResults(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID int64  `json:"launch_id"`
		Status   string `json:"status"`
		Page     int    `json:"page"`
		Size     int    `json:"size"`
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

	r.logger.Info("listing test results", map[string]any{
		"launch_id": params.LaunchID,
		"status":    params.Status,
		"page":      params.Page,
		"size":      params.Size,
	})

	results, err := r.allure.ListTestResults(ctx, params.LaunchID, params.Status, params.Page, params.Size)
	if err != nil {
		r.logger.Error("list test results", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("list test results: %w", err)
	}

	items := make([]map[string]any, len(results.Content))
	for i, result := range results.Content {
		items[i] = map[string]any{
			"id":         result.ID,
			"name":       result.Name,
			"status":     result.Status,
			"launch_id":  result.LaunchID,
			"start_time": result.StartTime,
			"end_time":   result.EndTime,
			"duration":   result.Duration,
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

func (r *Registry) getTestResult(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestResultID int64 `json:"test_result_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}

	r.logger.Info("fetching test result", map[string]any{"test_result_id": params.TestResultID})

	result, err := r.allure.GetTestResult(ctx, params.TestResultID)
	if err != nil {
		r.logger.Error("get test result", err, map[string]any{"test_result_id": params.TestResultID})
		return nil, fmt.Errorf("get test result: %w", err)
	}

	return map[string]any{
		"id":          result.ID,
		"uuid":        result.UUID,
		"name":        result.Name,
		"status":      result.Status,
		"launch_id":   result.LaunchID,
		"start_time":  result.StartTime,
		"end_time":    result.EndTime,
		"duration":    result.Duration,
		"full_name":   result.FullName,
		"description": result.Description,
		"parameters":  result.Parameters,
	}, nil
}

func (r *Registry) assignTestResult(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestResultID int64  `json:"test_result_id"`
		Username     string `json:"username"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}

	if params.Username == "" {
		return nil, fmt.Errorf("username is required")
	}

	r.logger.Info("assigning test result", map[string]any{
		"test_result_id": params.TestResultID,
		"username":       params.Username,
	})

	if err := r.allure.AssignTestResult(ctx, params.TestResultID, params.Username); err != nil {
		r.logger.Error("assign test result", err, map[string]any{"test_result_id": params.TestResultID})
		return nil, fmt.Errorf("assign test result: %w", err)
	}

	return map[string]any{"status": "assigned"}, nil
}

func (r *Registry) muteTestResult(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestResultID int64  `json:"test_result_id"`
		Reason       string `json:"reason"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}

	r.logger.Info("muting test result", map[string]any{
		"test_result_id": params.TestResultID,
		"reason":         params.Reason,
	})

	if err := r.allure.MuteTestResult(ctx, params.TestResultID, params.Reason); err != nil {
		r.logger.Error("mute test result", err, map[string]any{"test_result_id": params.TestResultID})
		return nil, fmt.Errorf("mute test result: %w", err)
	}

	return map[string]any{"status": "muted"}, nil
}

func (r *Registry) resolveTestResult(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestResultID int64 `json:"test_result_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}

	r.logger.Info("resolving test result", map[string]any{"test_result_id": params.TestResultID})

	if err := r.allure.ResolveTestResult(ctx, params.TestResultID); err != nil {
		r.logger.Error("resolve test result", err, map[string]any{"test_result_id": params.TestResultID})
		return nil, fmt.Errorf("resolve test result: %w", err)
	}

	return map[string]any{"status": "resolved"}, nil
}

func (r *Registry) unmuteTestResult(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestResultID int64 `json:"test_result_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}

	r.logger.Info("unmuting test result", map[string]any{"test_result_id": params.TestResultID})

	if err := r.allure.UnmuteTestResult(ctx, params.TestResultID); err != nil {
		r.logger.Error("unmute test result", err, map[string]any{"test_result_id": params.TestResultID})
		return nil, fmt.Errorf("unmute test result: %w", err)
	}

	return map[string]any{"status": "unmuted"}, nil
}
