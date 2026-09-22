package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
)

func (r *Registry) registerResultTools() {
	r.register(&Tool{
		Name: "list_test_results",
		Description: "List test results in a launch with optional status filter (passed, failed, broken, skipped, unknown). " +
			"The underlying API has no server-side status filter, so a filtered request scans the launch's results " +
			"client-side (capped; see the truncated field) and paginates over the matches — page/size apply to the " +
			"filtered list, not the launch's raw result order.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "Filter by status (passed, failed, broken, skipped, unknown)",
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "Page number (0-based)",
					"default":     0,
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "Items per page (default 10, max 1000)",
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
				"status": map[string]any{
					"type":        "string",
					"enum":        []string{"failed", "broken", "passed", "skipped", "unknown"},
					"description": "Resolution status to set",
				},
			},
			"required": []string{"test_result_id", "status"},
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

const (
	// listTestResultsScanPageSize is the underlying page size used while
	// scanning for status matches (confirmed live the API honors page sizes
	// well above 100 with no server-side clamp; kept well below that to leave
	// headroom rather than push the true limit).
	listTestResultsScanPageSize = 500
	// listTestResultsMaxScanPages caps a single filtered list_test_results call
	// at 20,000 scanned results, so a huge launch can't make one call scan forever.
	listTestResultsMaxScanPages = 40
)

// filterTestResultsByStatus scans launchID's test results (via the unfiltered
// Client.ListTestResults — the API has no server-side status filter) and
// returns up to `size` results whose status matches `status`
// (case-insensitive), skipping the first `page*size` matches. isLast reports
// whether there are no more matching results after this page; truncated
// reports whether the scan cap (listTestResultsMaxScanPages pages) was hit
// before that could be determined.
func (r *Registry) filterTestResultsByStatus(ctx context.Context, launchID int64, status string, page, size int) (matched []allure.TestResultItem, isLast bool, truncated bool, err error) {
	start := page * size
	end := start + size
	matchedCount := 0

	for p := 0; p < listTestResultsMaxScanPages; p++ {
		resp, ferr := r.allure.ListTestResults(ctx, launchID, p, listTestResultsScanPageSize)
		if ferr != nil {
			return nil, false, false, ferr
		}

		for _, tr := range resp.Content {
			if !strings.EqualFold(tr.Status, status) {
				continue
			}
			if matchedCount >= end {
				// Proof there's at least one more match beyond the requested page.
				return matched, false, false, nil
			}
			if matchedCount >= start {
				matched = append(matched, tr)
			}
			matchedCount++
		}

		if resp.Last || len(resp.Content) == 0 {
			return matched, true, false, nil
		}
		if p == listTestResultsMaxScanPages-1 {
			return matched, false, true, nil
		}
	}
	return matched, true, false, nil
}

func (r *Registry) listTestResults(ctx context.Context, args listTestResultsArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if args.Page < 0 {
		args.Page = 0
	}

	if args.Size <= 0 {
		args.Size = 10
	}
	// The API itself accepts far larger page sizes (confirmed live: size=300
	// returns a full 300-item page, no server-side clamp) — 100 was an
	// arbitrary client-side cap that silently truncated a caller's requested
	// size with no indication in the tool description. 1000 is a generous
	// ceiling to keep a single response reasonably sized.
	if args.Size > 1000 {
		args.Size = 1000
	}

	r.logger.Info("listing test results", map[string]any{
		"launch_id": args.LaunchID,
		"status":    args.Status,
		"page":      args.Page,
		"size":      args.Size,
	})

	buildItem := func(result allure.TestResultItem) map[string]any {
		return map[string]any{
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

	if args.Status != "" {
		matched, isLast, truncated, err := r.filterTestResultsByStatus(ctx, args.LaunchID, args.Status, args.Page, args.Size)
		if err != nil {
			r.logger.Error("list test results", err, map[string]any{"launch_id": args.LaunchID})
			return nil, fmt.Errorf("list test results: %w", err)
		}
		items := make([]map[string]any, len(matched))
		for i, result := range matched {
			items[i] = buildItem(result)
		}
		return map[string]any{
			"test_results": items,
			"page":         args.Page,
			"size":         args.Size,
			"is_last":      isLast,
			"truncated":    truncated,
		}, nil
	}

	results, err := r.allure.ListTestResults(ctx, args.LaunchID, args.Page, args.Size)
	if err != nil {
		r.logger.Error("list test results", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("list test results: %w", err)
	}

	items := make([]map[string]any, len(results.Content))
	for i, result := range results.Content {
		items[i] = buildItem(result)
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
	TestResultID int64  `json:"test_result_id"`
	Status       string `json:"status"`
}

func (r *Registry) resolveTestResult(ctx context.Context, args resolveTestResultArgs) (any, error) {
	if args.TestResultID <= 0 {
		return nil, fmt.Errorf("test_result_id must be positive")
	}
	if args.Status == "" {
		return nil, fmt.Errorf("status is required")
	}

	r.logger.Info("resolving test result", map[string]any{"test_result_id": args.TestResultID})

	if err := r.allure.ResolveTestResult(ctx, args.TestResultID, args.Status); err != nil {
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
