package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
)

func (r *Registry) registerTestCaseTools() {
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
		Description: "Get full test case details: metadata, manual_scenario (test execution steps), tags, members, custom fields, examples, and more",
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
					"description": "Test case name (optional)",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Description (optional)",
				},
				"full_name": map[string]any{
					"type":        "string",
					"description": "Full name (optional)",
				},
				"precondition": map[string]any{
					"type":        "string",
					"description": "Precondition (optional)",
				},
				"expected_result": map[string]any{
					"type":        "string",
					"description": "Expected result (optional)",
				},
				"automated": map[string]any{
					"type":        "boolean",
					"description": "Is automated (optional)",
				},
				"external": map[string]any{
					"type":        "boolean",
					"description": "Is external (optional)",
				},
				"deleted": map[string]any{
					"type":        "boolean",
					"description": "Mark as deleted (optional)",
				},
				"status_id": map[string]any{
					"type":        "integer",
					"description": "Status ID (optional)",
				},
				"test_layer_id": map[string]any{
					"type":        "integer",
					"description": "Test layer ID (optional)",
				},
				"workflow_id": map[string]any{
					"type":        "integer",
					"description": "Workflow ID (optional)",
				},
				"tags": map[string]any{
					"type":        "array",
					"description": "Tags (optional)",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id": map[string]any{"type": "integer"},
							"name": map[string]any{"type": "string"},
						},
					},
				},
				"members": map[string]any{
					"type":        "array",
					"description": "Members (optional)",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id": map[string]any{"type": "integer"},
							"name": map[string]any{"type": "string"},
						},
					},
				},
				"links": map[string]any{
					"type":        "array",
					"description": "External links (optional)",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{"type": "string"},
							"type": map[string]any{"type": "string"},
							"url": map[string]any{"type": "string"},
						},
					},
				},
				"manual_scenario": map[string]any{
					"type":        "object",
					"description": "Manual scenario with test execution steps (optional). Contains steps array with step definitions.",
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
		Name:        "clone_test_case",
		Description: "Clone an existing test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID to clone",
				},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.cloneTestCase,
	})

	r.register(&Tool{
		Name:        "restore_test_case",
		Description: "Restore a deleted test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID to restore",
				},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.restoreTestCase,
	})

	r.register(&Tool{
		Name:        "create_test_case_step",
		Description: "Create a new step in a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Test case ID",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Step body/description",
				},
				"after_id": map[string]any{
					"type":        "integer",
					"description": "Insert after step ID (optional)",
				},
				"parent_id": map[string]any{
					"type":        "integer",
					"description": "Parent step ID (optional)",
				},
			},
			"required": []string{"test_case_id", "body"},
		},
		Handler: r.createTestCaseStep,
	})

	r.register(&Tool{
		Name:        "update_test_case_step",
		Description: "Update an existing test case step",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"step_id": map[string]any{
					"type":        "integer",
					"description": "Step ID",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Step body/description (optional)",
				},
				"expected_result": map[string]any{
					"type":        "string",
					"description": "Expected result (optional)",
				},
			},
			"required": []string{"step_id"},
		},
		Handler: r.updateTestCaseStep,
	})

	r.register(&Tool{
		Name:        "delete_test_case_step",
		Description: "Delete a test case step",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"step_id": map[string]any{
					"type":        "integer",
					"description": "Step ID",
				},
			},
			"required": []string{"step_id"},
		},
		Handler: r.deleteTestCaseStep,
	})

	r.register(&Tool{
		Name:        "get_test_case_custom_fields",
		Description: "Get all custom field values for a test case",
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
		Handler: r.getTestCaseCustomFields,
	})

	r.register(&Tool{
		Name: "update_test_case_custom_fields",
		Description: "Update custom field values for a test case. " +
			"Each item must specify the custom field ID and the list of value IDs to set. " +
			"Use get_test_case_custom_fields first to discover available fields and their current values.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"custom_fields": map[string]any{
					"type":        "array",
					"description": "List of custom field values to set",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"custom_field_id": map[string]any{
								"type":        "integer",
								"description": "Custom field ID",
							},
							"value_ids": map[string]any{
								"type":        "array",
								"description": "List of value IDs to assign to this custom field",
								"items": map[string]any{
									"type": "integer",
								},
							},
						},
						"required": []string{"custom_field_id", "value_ids"},
					},
				},
			},
			"required": []string{"test_case_id", "custom_fields"},
		},
		Handler: r.updateTestCaseCustomFields,
	})

	r.register(&Tool{
		Name:        "get_test_case_history",
		Description: "Get test case change history and versions",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
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
			"required": []string{"test_case_id"},
		},
		Handler: r.getTestCaseHistory,
	})
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

	if params.Size <= 0 {
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

	tc, err := r.allure.GetTestCaseOverview(ctx, params.TestCaseID)
	if err != nil {
		r.logger.Error("get test case", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("get test case: %w", err)
	}

	scenario, err := r.allure.GetTestCaseScenario(ctx, params.TestCaseID)
	if err != nil {
		r.logger.Info("scenario not available", map[string]any{"test_case_id": params.TestCaseID})
	} else if scenario != nil {
		tc["manual_scenario"] = scenario
	}

	return tc, nil
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
		TestCaseID     int64                    `json:"test_case_id"`
		Name           string                   `json:"name"`
		Description    string                   `json:"description"`
		FullName       string                   `json:"full_name"`
		Precondition   string                   `json:"precondition"`
		ExpectedResult string                   `json:"expected_result"`
		Automated      *bool                    `json:"automated"`
		External       *bool                    `json:"external"`
		Deleted        *bool                    `json:"deleted"`
		StatusID       *int64                   `json:"status_id"`
		TestLayerID    *int64                   `json:"test_layer_id"`
		WorkflowID     *int64                   `json:"workflow_id"`
		Tags           []allure.TestTagDto      `json:"tags"`
		Members        []allure.MemberDto       `json:"members"`
		Links          []allure.ExternalLinkDto `json:"links"`
		ManualScenario map[string]any           `json:"manual_scenario"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	hasFields := params.Name != "" || params.Description != "" || params.FullName != "" ||
		params.Precondition != "" || params.ExpectedResult != "" ||
		params.Automated != nil || params.External != nil || params.Deleted != nil ||
		params.StatusID != nil || params.TestLayerID != nil || params.WorkflowID != nil ||
		len(params.Tags) > 0 || len(params.Members) > 0 || len(params.Links) > 0 ||
		len(params.ManualScenario) > 0

	if !hasFields {
		return nil, fmt.Errorf("at least one field must be provided")
	}

	req := allure.UpdateTestCaseRequest{
		Name:           params.Name,
		Description:    params.Description,
		FullName:       params.FullName,
		Precondition:   params.Precondition,
		ExpectedResult: params.ExpectedResult,
		Automated:      params.Automated,
		External:       params.External,
		Deleted:        params.Deleted,
		StatusID:       params.StatusID,
		TestLayerID:    params.TestLayerID,
		WorkflowID:     params.WorkflowID,
		Tags:           params.Tags,
		Members:        params.Members,
		Links:          params.Links,
		ManualScenario: params.ManualScenario,
	}

	r.logger.Info("updating test case", map[string]any{
		"test_case_id": params.TestCaseID,
	})

	if err := r.allure.UpdateTestCase(ctx, params.TestCaseID, req); err != nil {
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

func (r *Registry) cloneTestCase(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("cloning test case", map[string]any{"test_case_id": params.TestCaseID})

	newID, err := r.allure.CloneTestCase(ctx, params.TestCaseID)
	if err != nil {
		r.logger.Error("clone test case", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("clone test case: %w", err)
	}

	return map[string]any{"cloned_test_case_id": newID, "status": "cloned"}, nil
}

func (r *Registry) restoreTestCase(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("restoring test case", map[string]any{"test_case_id": params.TestCaseID})

	if err := r.allure.RestoreTestCase(ctx, params.TestCaseID); err != nil {
		r.logger.Error("restore test case", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("restore test case: %w", err)
	}

	return map[string]any{"status": "restored"}, nil
}

func (r *Registry) createTestCaseStep(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64  `json:"test_case_id"`
		Body       string `json:"body"`
		AfterID    int64  `json:"after_id"`
		ParentID   int64  `json:"parent_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if params.Body == "" {
		return nil, fmt.Errorf("body must be provided")
	}

	req := allure.ScenarioStepCreateRequest{
		TestCaseID: params.TestCaseID,
		Body:       params.Body,
		ParentID:   params.ParentID,
	}

	r.logger.Info("creating test case step", map[string]any{
		"test_case_id": params.TestCaseID,
		"body":         params.Body,
	})

	stepID, err := r.allure.CreateTestCaseStep(ctx, req, params.AfterID)
	if err != nil {
		r.logger.Error("create test case step", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("create test case step: %w", err)
	}

	return map[string]any{"step_id": stepID}, nil
}

func (r *Registry) updateTestCaseStep(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		StepID         int64  `json:"step_id"`
		Body           string `json:"body"`
		ExpectedResult string `json:"expected_result"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.StepID <= 0 {
		return nil, fmt.Errorf("step_id must be positive")
	}

	if params.Body == "" && params.ExpectedResult == "" {
		return nil, fmt.Errorf("at least one field (body or expected_result) must be provided")
	}

	req := allure.ScenarioStepPatchRequest{
		Body:           params.Body,
		ExpectedResult: params.ExpectedResult,
	}

	r.logger.Info("updating test case step", map[string]any{
		"step_id": params.StepID,
	})

	if err := r.allure.UpdateTestCaseStep(ctx, params.StepID, req); err != nil {
		r.logger.Error("update test case step", err, map[string]any{"step_id": params.StepID})
		return nil, fmt.Errorf("update test case step: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

func (r *Registry) deleteTestCaseStep(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		StepID int64 `json:"step_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.StepID <= 0 {
		return nil, fmt.Errorf("step_id must be positive")
	}

	r.logger.Info("deleting test case step", map[string]any{"step_id": params.StepID})

	if err := r.allure.DeleteTestCaseStep(ctx, params.StepID); err != nil {
		r.logger.Error("delete test case step", err, map[string]any{"step_id": params.StepID})
		return nil, fmt.Errorf("delete test case step: %w", err)
	}

	return map[string]any{"status": "deleted"}, nil
}

func (r *Registry) getTestCaseCustomFields(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("fetching test case custom fields", map[string]any{"test_case_id": params.TestCaseID})

	fields, err := r.allure.GetTestCaseCustomFields(ctx, params.TestCaseID)
	if err != nil {
		r.logger.Error("get test case custom fields", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("get test case custom fields: %w", err)
	}

	result := make([]map[string]any, len(fields))
	for i, f := range fields {
		values := make([]map[string]any, len(f.Values))
		for j, v := range f.Values {
			values[j] = map[string]any{
				"id":   v.ID,
				"name": v.Name,
			}
		}
		result[i] = map[string]any{
			"custom_field_id":   f.CustomField.ID,
			"custom_field_name": f.CustomField.Name,
			"values":            values,
		}
	}

	return map[string]any{"custom_fields": result}, nil
}

func (r *Registry) updateTestCaseCustomFields(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID   int64 `json:"test_case_id"`
		CustomFields []struct {
			CustomFieldID int64   `json:"custom_field_id"`
			ValueIDs      []int64 `json:"value_ids"`
		} `json:"custom_fields"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if len(params.CustomFields) == 0 {
		return nil, fmt.Errorf("custom_fields must contain at least one entry")
	}

	fields := make([]allure.CustomFieldWithValuesDto, len(params.CustomFields))
	for i, cf := range params.CustomFields {
		values := make([]allure.CustomFieldValueDto, len(cf.ValueIDs))
		for j, vid := range cf.ValueIDs {
			values[j] = allure.CustomFieldValueDto{ID: vid}
		}
		fields[i] = allure.CustomFieldWithValuesDto{
			CustomField: allure.CustomFieldDto{ID: cf.CustomFieldID},
			Values:      values,
		}
	}

	r.logger.Info("updating test case custom fields", map[string]any{
		"test_case_id":   params.TestCaseID,
		"fields_count":   len(fields),
	})

	if err := r.allure.UpdateTestCaseCustomFields(ctx, params.TestCaseID, fields); err != nil {
		r.logger.Error("update test case custom fields", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("update test case custom fields: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

func (r *Registry) getTestCaseHistory(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
		Page       int   `json:"page"`
		Size       int   `json:"size"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if params.Size <= 0 {
		params.Size = 10
	}
	if params.Size > 100 {
		params.Size = 100
	}

	r.logger.Info("fetching test case history", map[string]any{
		"test_case_id": params.TestCaseID,
		"page":         params.Page,
		"size":         params.Size,
	})

	history, err := r.allure.GetTestCaseHistory(ctx, params.TestCaseID, params.Page, params.Size)
	if err != nil {
		r.logger.Error("get test case history", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("get test case history: %w", err)
	}

	return history, nil
}
