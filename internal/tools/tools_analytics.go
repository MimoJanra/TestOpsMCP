package tools

import (
	"context"
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
		Handler: Typed(r.getLaunchTrendAnalytics),
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
		Handler: Typed(r.getLaunchDurationAnalytics),
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
		Handler: Typed(r.getTestSuccessRate),
	})
}

type getLaunchTrendAnalyticsArgs struct {
	ProjectID int64 `json:"project_id"`
}

func (r *Registry) getLaunchTrendAnalytics(ctx context.Context, args getLaunchTrendAnalyticsArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching launch trend analytics", map[string]any{"project_id": args.ProjectID})

	trends, err := r.allure.GetLaunchTrendAnalytics(ctx, args.ProjectID)
	if err != nil {
		r.logger.Error("get launch trend analytics", err, map[string]any{"project_id": args.ProjectID})
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

type getLaunchDurationAnalyticsArgs struct {
	ProjectID int64 `json:"project_id"`
}

func (r *Registry) getLaunchDurationAnalytics(ctx context.Context, args getLaunchDurationAnalyticsArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching launch duration analytics", map[string]any{"project_id": args.ProjectID})

	data, err := r.allure.GetLaunchDurationAnalytics(ctx, args.ProjectID)
	if err != nil {
		r.logger.Error("get launch duration analytics", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("get launch duration analytics: %w", err)
	}

	return map[string]any{"duration_histogram": data}, nil
}

type getTestSuccessRateArgs struct {
	ProjectID int64 `json:"project_id"`
}

func (r *Registry) getTestSuccessRate(ctx context.Context, args getTestSuccessRateArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching test success rate", map[string]any{"project_id": args.ProjectID})

	data, err := r.allure.GetTestSuccessRateAnalytics(ctx, args.ProjectID)
	if err != nil {
		r.logger.Error("get test success rate", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("get test success rate: %w", err)
	}

	return map[string]any{"success_rate": data}, nil
}
