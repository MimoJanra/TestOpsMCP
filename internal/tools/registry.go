package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
)

type HandlerFunc func(ctx context.Context, input json.RawMessage) (any, error)

type Tool struct {
	Name        string
	Description string
	InputSchema any
	Handler     HandlerFunc
}

// Registry holds the set of tools exposed by the MCP server.
// Tools are registered once during NewRegistry and must not be mutated
// afterwards; ListTools therefore returns shared pointers without copying.
type Registry struct {
	tools  map[string]*Tool
	allure *allure.Client
	logger *core.Logger
	mu     sync.RWMutex
}

func NewRegistry(allureClient *allure.Client, logger *core.Logger) *Registry {
	r := &Registry{
		tools:  make(map[string]*Tool),
		allure: allureClient,
		logger: logger,
	}
	r.registerTools()
	return r
}

func (r *Registry) registerTools() {
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
		Name:        "list_test_cases",
		Description: "List test cases in a project",
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
		Handler: r.listTestCases,
	})

	r.register(&Tool{
		Name:        "get_test_case",
		Description: "Get test case details and steps",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.getTestCase,
	})

	r.register(&Tool{
		Name:        "run_test_case",
		Description: "Start a test run for a specific test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID to run in",
				},
			},
			"required": []string{"test_case_id", "launch_id"},
		},
		Handler: r.runTestCase,
	})

	r.register(&Tool{
		Name:        "create_test_case",
		Description: "Create a new test case in a project",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Test case name",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Test case description (optional)",
				},
			},
			"required": []string{"project_id", "name"},
		},
		Handler: r.createTestCase,
	})

	r.register(&Tool{
		Name:        "update_test_case",
		Description: "Update an existing test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "New test case name (optional)",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "New description (optional)",
				},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.updateTestCase,
	})

	r.register(&Tool{
		Name:        "delete_test_case",
		Description: "Delete a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.deleteTestCase,
	})

	r.register(&Tool{
		Name:        "list_projects",
		Description: "List all accessible projects",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
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
		},
		Handler: r.listProjects,
	})

	r.register(&Tool{
		Name:        "get_project",
		Description: "Get project details and settings",
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
		Handler: r.getProject,
	})

	r.register(&Tool{
		Name:        "get_project_stats",
		Description: "Get project statistics (test count, runs)",
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
		Handler: r.getProjectStats,
	})

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

func (r *Registry) register(tool *Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
}

func (r *Registry) GetTool(name string) *Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tools[name]
}

func (r *Registry) ListTools() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
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

func (r *Registry) listTestCases(ctx context.Context, input json.RawMessage) (any, error) {
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

	r.logger.Info("listing test cases", map[string]any{
		"project_id": params.ProjectID,
		"page":       params.Page,
		"size":       params.Size,
	})

	cases, err := r.allure.ListTestCases(ctx, params.ProjectID, params.Page, params.Size)
	if err != nil {
		r.logger.Error("list test cases", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("list test cases: %w", err)
	}

	items := make([]map[string]any, len(cases.Content))
	for i, tc := range cases.Content {
		items[i] = map[string]any{
			"id":                tc.ID,
			"name":              tc.Name,
			"project_id":        tc.ProjectID,
			"status":            tc.Status,
			"automation_status": tc.AutomationStatus,
		}
	}

	return map[string]any{
		"test_cases": items,
		"page":       cases.Number,
		"size":       cases.Size,
		"total":      cases.Total,
		"is_last":    cases.Last,
	}, nil
}

func (r *Registry) getTestCase(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("fetching test case", map[string]any{"test_case_id": params.TestCaseID})

	tc, err := r.allure.GetTestCase(ctx, params.TestCaseID)
	if err != nil {
		r.logger.Error("get test case", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("get test case: %w", err)
	}

	return map[string]any{
		"id":                tc.ID,
		"uuid":              tc.UUID,
		"name":              tc.Name,
		"project_id":        tc.ProjectID,
		"description":       tc.Description,
		"status":            tc.Status,
		"automation_status": tc.AutomationStatus,
		"full_name":         tc.FullName,
	}, nil
}

func (r *Registry) runTestCase(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
		LaunchID   int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("running test case", map[string]any{
		"test_case_id": params.TestCaseID,
		"launch_id":    params.LaunchID,
	})

	if err := r.allure.RunTestCase(ctx, params.TestCaseID, params.LaunchID); err != nil {
		r.logger.Error("run test case", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("run test case: %w", err)
	}

	return map[string]any{"status": "started"}, nil
}

func (r *Registry) listProjects(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		Page int `json:"page"`
		Size int `json:"size"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.Size == 0 {
		params.Size = 10
	}
	if params.Size > 100 {
		params.Size = 100
	}

	r.logger.Info("listing projects", map[string]any{
		"page": params.Page,
		"size": params.Size,
	})

	projects, err := r.allure.ListProjects(ctx, params.Page, params.Size)
	if err != nil {
		r.logger.Error("list projects", err, map[string]any{})
		return nil, fmt.Errorf("list projects: %w", err)
	}

	items := make([]map[string]any, len(projects.Content))
	for i, p := range projects.Content {
		items[i] = map[string]any{
			"id":   p.ID,
			"name": p.Name,
			"code": p.Code,
		}
	}

	return map[string]any{
		"projects": items,
		"page":     projects.Number,
		"size":     projects.Size,
		"total":    projects.Total,
		"is_last":  projects.Last,
	}, nil
}

func (r *Registry) getProject(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID int64 `json:"project_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching project", map[string]any{"project_id": params.ProjectID})

	project, err := r.allure.GetProject(ctx, params.ProjectID)
	if err != nil {
		r.logger.Error("get project", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("get project: %w", err)
	}

	return map[string]any{
		"id":          project.ID,
		"name":        project.Name,
		"code":        project.Code,
		"description": project.Description,
	}, nil
}

func (r *Registry) getProjectStats(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID int64 `json:"project_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching project stats", map[string]any{"project_id": params.ProjectID})

	stats, err := r.allure.GetProjectStats(ctx, params.ProjectID)
	if err != nil {
		r.logger.Error("get project stats", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("get project stats: %w", err)
	}

	return map[string]any{
		"project_id":           stats.ID,
		"automated_test_cases": stats.AutomatedTestCases,
		"manual_test_cases":    stats.ManualTestCases,
		"automation_percent":   stats.AutomationPercent,
		"launches":             stats.Launches,
	}, nil
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

func (r *Registry) createTestCase(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID   int64  `json:"project_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	if params.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	r.logger.Info("creating test case", map[string]any{
		"project_id":  params.ProjectID,
		"name":        params.Name,
		"description": params.Description,
	})

	tc, err := r.allure.CreateTestCase(ctx, params.ProjectID, params.Name, params.Description)
	if err != nil {
		r.logger.Error("create test case", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("create test case: %w", err)
	}

	return map[string]any{
		"id":                tc.ID,
		"uuid":              tc.UUID,
		"name":              tc.Name,
		"project_id":        tc.ProjectID,
		"description":       tc.Description,
		"status":            tc.Status,
		"automation_status": tc.AutomationStatus,
		"full_name":         tc.FullName,
	}, nil
}

func (r *Registry) updateTestCase(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID  int64  `json:"test_case_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if params.Name == "" && params.Description == "" {
		return nil, fmt.Errorf("at least one field (name or description) must be provided")
	}

	r.logger.Info("updating test case", map[string]any{
		"test_case_id": params.TestCaseID,
		"name":         params.Name,
		"description":  params.Description,
	})

	if err := r.allure.UpdateTestCase(ctx, params.TestCaseID, params.Name, params.Description); err != nil {
		r.logger.Error("update test case", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("update test case: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

func (r *Registry) deleteTestCase(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("deleting test case", map[string]any{"test_case_id": params.TestCaseID})

	if err := r.allure.DeleteTestCase(ctx, params.TestCaseID); err != nil {
		r.logger.Error("delete test case", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("delete test case: %w", err)
	}

	return map[string]any{"status": "deleted"}, nil
}
