package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/MimoJanra/TestOpsMCP/internal/session"
)

// truncateRunes truncates s to at most n runes, appending "…" if cut.
// Safe for any Unicode content including multi-byte sequences.
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func (r *Registry) registerAnalysisTools() {
	r.register(&Tool{
		Name: "analyze_launch_failures",
		Description: "Analyze failed and broken tests in a launch using AI. " +
			"Fetches the failing test results and asks the client's model (MCP sampling) to identify root causes and suggest fixes. " +
			"If the client doesn't support sampling, returns the collected failures without an analysis.",
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

	// A missing launch used to come back as "No failed tests found".
	if _, err := r.requireLaunch(ctx, args.LaunchID); err != nil {
		return nil, err
	}
	// The API has no server-side status filter (see Client.ListTestResults),
	// so this scans the launch's results client-side. Broken (errored) tests
	// are failures too and were previously skipped.
	results, _, _, err := r.filterTestResultsByStatus(ctx, args.LaunchID, "failed", 0, args.MaxFailures)
	if err != nil {
		return nil, fmt.Errorf("list failed results: %w", err)
	}
	if len(results) < args.MaxFailures {
		broken, _, _, err := r.filterTestResultsByStatus(ctx, args.LaunchID, "broken", 0, args.MaxFailures-len(results))
		if err != nil {
			return nil, fmt.Errorf("list broken results: %w", err)
		}
		results = append(results, broken...)
	}

	if len(results) == 0 {
		return map[string]any{
			"launch_id": args.LaunchID,
			"failures":  0,
			"analysis":  "No failed or broken tests found in this launch.",
		}, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Launch #%d — %d failed/broken tests:\n\n", args.LaunchID, len(results))
	for i, res := range results {
		fmt.Fprintf(&sb, "%d. [%s] %s\n", i+1, res.Status, res.Name)
		if res.Message != "" {
			fmt.Fprintf(&sb, "   Error: %s\n", truncateRunes(res.Message, 300))
		}
		if res.Trace != "" {
			fmt.Fprintf(&sb, "   Trace: %s\n", truncateRunes(res.Trace, 200))
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
			"failures":  len(results),
			"summary":   sb.String(),
			"error":     "AI analysis unavailable: " + err.Error(),
		}, nil
	}

	return map[string]any{
		"launch_id": args.LaunchID,
		"failures":  len(results),
		"analysis":  sampResult.Text,
	}, nil
}
