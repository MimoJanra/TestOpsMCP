package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/MimoJanra/TestOpsMCP/internal/session"
)

func (r *Registry) registerAnalysisTools() {
	r.register(&Tool{
		Name: "analyze_launch_failures",
		Description: "Analyze failed tests in a launch using AI. " +
			"Fetches failed test results and asks Claude to identify root causes and suggest fixes.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID to analyze",
				},
				"max_failures": map[string]any{
					"type":        "integer",
					"description": "Maximum number of failures to analyze (default 20, max 50)",
					"default":     20,
				},
			},
			"required": []string{"launch_id"},
		},
		Annotations: map[string]any{
			"readOnlyHint":    true,
			"destructiveHint": false,
		},
		Handler: Typed(r.analyzeLaunchFailures),
	})
}

type analyzeLaunchFailuresArgs struct {
	LaunchID    int64 `json:"launch_id"`
	MaxFailures int   `json:"max_failures"`
}

func (r *Registry) analyzeLaunchFailures(ctx context.Context, args analyzeLaunchFailuresArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if args.MaxFailures <= 0 {
		args.MaxFailures = 20
	}
	if args.MaxFailures > 50 {
		args.MaxFailures = 50
	}

	sample, ok := session.SamplingFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("sampling not available in this session")
	}

	r.logger.Info("analyzing launch failures", map[string]any{"launch_id": args.LaunchID})

	results, err := r.allure.ListTestResults(ctx, args.LaunchID, "FAILED", 0, args.MaxFailures)
	if err != nil {
		return nil, fmt.Errorf("list failed results: %w", err)
	}

	if len(results.Content) == 0 {
		return map[string]any{
			"launch_id": args.LaunchID,
			"failures":  0,
			"analysis":  "No failed tests found in this launch.",
		}, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Launch #%d — %d failed tests:\n\n", args.LaunchID, len(results.Content))
	for i, res := range results.Content {
		fmt.Fprintf(&sb, "%d. %s\n", i+1, res.Name)
		if res.Message != "" {
			msg := res.Message
			if len(msg) > 300 {
				msg = msg[:300] + "..."
			}
			fmt.Fprintf(&sb, "   Error: %s\n", msg)
		}
		if res.Trace != "" {
			trace := res.Trace
			if len(trace) > 200 {
				trace = trace[:200] + "..."
			}
			fmt.Fprintf(&sb, "   Trace: %s\n", trace)
		}
		fmt.Fprintln(&sb)
	}

	sampResult, err := sample(ctx,
		"You are a test analytics expert. Analyze the test failures below. "+
			"Identify common root causes, group related failures, and suggest specific fixes. "+
			"Be concise and actionable.",
		[]session.SamplingMessage{{
			Role: "user",
			Text: sb.String(),
		}},
		2000,
	)
	if err != nil {
		r.logger.Error("sampling failed", err, map[string]any{"launch_id": args.LaunchID})
		return map[string]any{
			"launch_id": args.LaunchID,
			"failures":  len(results.Content),
			"summary":   sb.String(),
			"error":     "AI analysis unavailable: " + err.Error(),
		}, nil
	}

	return map[string]any{
		"launch_id": args.LaunchID,
		"failures":  len(results.Content),
		"analysis":  sampResult.Text,
	}, nil
}
