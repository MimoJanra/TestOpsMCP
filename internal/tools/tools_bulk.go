package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
)

func (r *Registry) registerBulkTools() {
	r.register(&Tool{
		Name:        "bulk_set_test_case_status",
		Description: "Bulk update status for multiple test cases",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"test_case_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test case IDs",
				},
				"status_id": map[string]any{
					"type":        "integer",
					"description": "Status ID to set",
				},
				"workflow_id": map[string]any{
					"type":        "integer",
					"description": "Workflow ID",
				},
			},
			"required": []string{"project_id", "test_case_ids", "status_id", "workflow_id"},
		},
		Handler: r.bulkSetTestCaseStatus,
	})

	r.register(&Tool{
		Name:        "bulk_add_test_case_tags",
		Description: "Bulk add tags to test cases",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"test_case_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test case IDs",
				},
				"tags": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id": map[string]any{
								"type": "integer",
							},
							"name": map[string]any{
								"type": "string",
							},
						},
					},
					"description": "Tags to add (with id and name)",
				},
			},
			"required": []string{"project_id", "test_case_ids", "tags"},
		},
		Handler: r.bulkAddTestCaseTags,
	})

	r.register(&Tool{
		Name:        "bulk_remove_test_case_tags",
		Description: "Bulk remove tags from test cases",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"test_case_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test case IDs",
				},
				"tags": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id": map[string]any{
								"type": "integer",
							},
							"name": map[string]any{
								"type": "string",
							},
						},
					},
					"description": "Tags to remove (with id and name)",
				},
			},
			"required": []string{"project_id", "test_case_ids", "tags"},
		},
		Handler: r.bulkRemoveTestCaseTags,
	})

	r.register(&Tool{
		Name:        "bulk_clone_test_cases",
		Description: "Bulk clone multiple test cases",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"test_case_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test case IDs to clone",
				},
			},
			"required": []string{"project_id", "test_case_ids"},
		},
		Handler: r.bulkCloneTestCases,
	})

	r.register(&Tool{
		Name:        "bulk_assign_test_results",
		Description: "Bulk assign test results to users",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"test_result_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test result IDs",
				},
				"assignees": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
					"description": "Usernames to assign to",
				},
			},
			"required": []string{"launch_id", "test_result_ids"},
		},
		Handler: r.bulkAssignTestResults,
	})

	r.register(&Tool{
		Name:        "bulk_mute_test_results",
		Description: "Bulk mute test results",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"test_result_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test result IDs",
				},
				"reason": map[string]any{
					"type":        "string",
					"description": "Reason for muting",
				},
			},
			"required": []string{"launch_id", "test_result_ids"},
		},
		Handler: r.bulkMuteTestResults,
	})

	r.register(&Tool{
		Name:        "bulk_unmute_test_results",
		Description: "Bulk unmute test results",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"test_result_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test result IDs",
				},
			},
			"required": []string{"launch_id", "test_result_ids"},
		},
		Handler: r.bulkUnmuteTestResults,
	})

	r.register(&Tool{
		Name:        "bulk_resolve_test_results",
		Description: "Bulk resolve test results",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
				"test_result_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Test result IDs",
				},
			},
			"required": []string{"launch_id", "test_result_ids"},
		},
		Handler: r.bulkResolveTestResults,
	})
}

func (r *Registry) bulkSetTestCaseStatus(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID   int64   `json:"project_id"`
		TestCaseIDs []int64 `json:"test_case_ids"`
		StatusID    int64   `json:"status_id"`
		WorkflowID  int64   `json:"workflow_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(params.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if params.StatusID <= 0 {
		return nil, fmt.Errorf("status_id must be positive")
	}
	if params.WorkflowID <= 0 {
		return nil, fmt.Errorf("workflow_id must be positive")
	}

	r.logger.Info("bulk setting test case status", map[string]any{"project_id": params.ProjectID, "count": len(params.TestCaseIDs)})

	if err := r.allure.BulkSetTestCaseStatus(ctx, params.ProjectID, params.StatusID, params.WorkflowID, params.TestCaseIDs); err != nil {
		r.logger.Error("bulk set test case status", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("bulk set status: %w", err)
	}

	return map[string]any{"status": "success", "count": len(params.TestCaseIDs)}, nil
}

func (r *Registry) bulkAddTestCaseTags(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID   int64               `json:"project_id"`
		TestCaseIDs []int64             `json:"test_case_ids"`
		Tags        []allure.TestTagDto `json:"tags"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(params.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if len(params.Tags) == 0 {
		return nil, fmt.Errorf("tags must not be empty")
	}

	r.logger.Info("bulk adding test case tags", map[string]any{"project_id": params.ProjectID, "count": len(params.TestCaseIDs)})

	if err := r.allure.BulkAddTestCaseTags(ctx, params.ProjectID, params.TestCaseIDs, params.Tags); err != nil {
		r.logger.Error("bulk add test case tags", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("bulk add tags: %w", err)
	}

	return map[string]any{"status": "success", "count": len(params.TestCaseIDs)}, nil
}

func (r *Registry) bulkRemoveTestCaseTags(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID   int64               `json:"project_id"`
		TestCaseIDs []int64             `json:"test_case_ids"`
		Tags        []allure.TestTagDto `json:"tags"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(params.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if len(params.Tags) == 0 {
		return nil, fmt.Errorf("tags must not be empty")
	}

	r.logger.Info("bulk removing test case tags", map[string]any{"project_id": params.ProjectID, "count": len(params.TestCaseIDs)})

	if err := r.allure.BulkRemoveTestCaseTags(ctx, params.ProjectID, params.TestCaseIDs, params.Tags); err != nil {
		r.logger.Error("bulk remove test case tags", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("bulk remove tags: %w", err)
	}

	return map[string]any{"status": "success", "count": len(params.TestCaseIDs)}, nil
}

func (r *Registry) bulkCloneTestCases(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID   int64   `json:"project_id"`
		TestCaseIDs []int64 `json:"test_case_ids"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	if len(params.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}

	r.logger.Info("bulk cloning test cases", map[string]any{
		"project_id": params.ProjectID,
		"count":      len(params.TestCaseIDs),
	})

	if err := r.allure.BulkCloneTestCases(ctx, params.ProjectID, params.TestCaseIDs); err != nil {
		r.logger.Error("bulk clone test cases", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("bulk clone test cases: %w", err)
	}

	return map[string]any{"status": "cloning_started", "count": len(params.TestCaseIDs)}, nil
}

func (r *Registry) bulkAssignTestResults(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID      int64    `json:"launch_id"`
		TestResultIDs []int64  `json:"test_result_ids"`
		Assignees     []string `json:"assignees"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if len(params.TestResultIDs) == 0 {
		return nil, fmt.Errorf("test_result_ids must not be empty")
	}

	r.logger.Info("bulk assigning test results", map[string]any{"launch_id": params.LaunchID, "count": len(params.TestResultIDs)})

	if err := r.allure.BulkAssignTestResults(ctx, params.LaunchID, params.TestResultIDs, params.Assignees); err != nil {
		r.logger.Error("bulk assign test results", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("bulk assign: %w", err)
	}

	return map[string]any{"status": "success", "count": len(params.TestResultIDs)}, nil
}

func (r *Registry) bulkMuteTestResults(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID      int64   `json:"launch_id"`
		TestResultIDs []int64 `json:"test_result_ids"`
		Reason        string  `json:"reason"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if len(params.TestResultIDs) == 0 {
		return nil, fmt.Errorf("test_result_ids must not be empty")
	}

	r.logger.Info("bulk muting test results", map[string]any{"launch_id": params.LaunchID, "count": len(params.TestResultIDs)})

	if err := r.allure.BulkMuteTestResults(ctx, params.LaunchID, params.TestResultIDs, params.Reason); err != nil {
		r.logger.Error("bulk mute test results", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("bulk mute: %w", err)
	}

	return map[string]any{"status": "success", "count": len(params.TestResultIDs)}, nil
}

func (r *Registry) bulkUnmuteTestResults(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID      int64   `json:"launch_id"`
		TestResultIDs []int64 `json:"test_result_ids"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if len(params.TestResultIDs) == 0 {
		return nil, fmt.Errorf("test_result_ids must not be empty")
	}

	r.logger.Info("bulk unmuting test results", map[string]any{"launch_id": params.LaunchID, "count": len(params.TestResultIDs)})

	if err := r.allure.BulkUnmuteTestResults(ctx, params.LaunchID, params.TestResultIDs); err != nil {
		r.logger.Error("bulk unmute test results", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("bulk unmute: %w", err)
	}

	return map[string]any{"status": "success", "count": len(params.TestResultIDs)}, nil
}

func (r *Registry) bulkResolveTestResults(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		LaunchID      int64   `json:"launch_id"`
		TestResultIDs []int64 `json:"test_result_ids"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if len(params.TestResultIDs) == 0 {
		return nil, fmt.Errorf("test_result_ids must not be empty")
	}

	r.logger.Info("bulk resolving test results", map[string]any{"launch_id": params.LaunchID, "count": len(params.TestResultIDs)})

	if err := r.allure.BulkResolveTestResults(ctx, params.LaunchID, params.TestResultIDs); err != nil {
		r.logger.Error("bulk resolve test results", err, map[string]any{"launch_id": params.LaunchID})
		return nil, fmt.Errorf("bulk resolve: %w", err)
	}

	return map[string]any{"status": "success", "count": len(params.TestResultIDs)}, nil
}
