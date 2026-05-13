package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

func (r *Registry) registerAnalyticsTools() {
	r.register(&Tool{
		Name:        "get_launch_trend_analytics",
		Description: "Get launch trend data over time (passed/failed/broken)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
			},
			"required": []string{"project_id"},
		},
		Handler: r.getLaunchTrendAnalytics,
	})

	r.register(&Tool{
		Name:        "get_launch_duration_analytics",
		Description: "Get launch execution time distribution",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
			},
			"required": []string{"project_id"},
		},
		Handler: r.getLaunchDurationAnalytics,
	})

	r.register(&Tool{
		Name:        "get_test_success_rate",
		Description: "Get test case success rate metrics",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
			},
			"required": []string{"project_id"},
		},
		Handler: r.getTestSuccessRate,
	})
}

func (r *Registry) getLaunchTrendAnalytics(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID int64 `json:"project_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching launch trend analytics", map[string]any{"project_id": params.ProjectID})

	trends, err := r.allure.GetLaunchTrendAnalytics(ctx, params.ProjectID)
	if err != nil {
		r.logger.Error("get launch trend analytics", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("get launch trend analytics: %w", err)
	}

	trendItems := make([]map[string]any, len(trends))
	for i, t := range trends {
		trendItems[i] = map[string]any{
			"passed":  t.Passed,
			"failed":  t.Failed,
			"broken":  t.Broken,
			"skipped": t.Skipped,
		}
	}

	return map[string]any{"trends": trendItems}, nil
}

func (r *Registry) getLaunchDurationAnalytics(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID int64 `json:"project_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching launch duration analytics", map[string]any{"project_id": params.ProjectID})

	data, err := r.allure.GetLaunchDurationAnalytics(ctx, params.ProjectID)
	if err != nil {
		r.logger.Error("get launch duration analytics", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("get launch duration analytics: %w", err)
	}

	return map[string]any{"duration_histogram": data}, nil
}

func (r *Registry) getTestSuccessRate(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID int64 `json:"project_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching test success rate", map[string]any{"project_id": params.ProjectID})

	data, err := r.allure.GetTestSuccessRateAnalytics(ctx, params.ProjectID)
	if err != nil {
		r.logger.Error("get test success rate", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("get test success rate: %w", err)
	}

	return map[string]any{"success_rate": data}, nil
}
