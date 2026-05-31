package tools

import (
	"context"
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
		Handler: Typed(r.listTestCases),
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
		Handler: Typed(r.getTestCase),
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
		Handler: Typed(r.runTestCase),
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
		Handler: Typed(r.createTestCase),
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
		Handler: Typed(r.updateTestCase),
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
		Handler: Typed(r.deleteTestCase),
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
		Handler: Typed(r.cloneTestCase),
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
		Handler: Typed(r.restoreTestCase),
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
		Handler: Typed(r.createTestCaseStep),
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
		Handler: Typed(r.updateTestCaseStep),
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
		Handler: Typed(r.deleteTestCaseStep),
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
		Handler: Typed(r.getTestCaseCustomFields),
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
		Handler: Typed(r.updateTestCaseCustomFields),
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
		Handler: Typed(r.getTestCaseHistory),
	})
}

type listTestCasesArgs struct {
	ProjectID int64 `json:"project_id"`
	Page      int   `json:"page"`
	Size      int   `json:"size"`
}

func (r *Registry) listTestCases(ctx context.Context, args listTestCasesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	if args.Size <= 0 {
		args.Size = 10
	}
	if args.Size > 100 {
		args.Size = 100
	}

	r.logger.Info("listing test cases", map[string]any{
		"project_id": args.ProjectID,
		"page":       args.Page,
		"size":       args.Size,
	})

	cases, err := r.allure.ListTestCases(ctx, args.ProjectID, args.Page, args.Size)
	if err != nil {
		r.logger.Error("list test cases", err, map[string]any{"project_id": args.ProjectID})
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

type getTestCaseArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCase(ctx context.Context, args getTestCaseArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("fetching test case", map[string]any{"test_case_id": args.TestCaseID})

	tc, err := r.allure.GetTestCaseOverview(ctx, args.TestCaseID)
	if err != nil {
		r.logger.Error("get test case", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("get test case: %w", err)
	}

	scenario, err := r.allure.GetTestCaseScenario(ctx, args.TestCaseID)
	if err != nil {
		r.logger.Info("scenario not available", map[string]any{"test_case_id": args.TestCaseID})
	} else if scenario != nil {
		tc["manual_scenario"] = scenario
	}

	return tc, nil
}

type runTestCaseArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	LaunchID   int64 `json:"launch_id"`
}

func (r *Registry) runTestCase(ctx context.Context, args runTestCaseArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("running test case", map[string]any{
		"test_case_id": args.TestCaseID,
		"launch_id":    args.LaunchID,
	})

	if err := r.allure.RunTestCase(ctx, args.TestCaseID, args.LaunchID); err != nil {
		r.logger.Error("run test case", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("run test case: %w", err)
	}

	return map[string]any{"status": "started"}, nil
}

type createTestCaseArgs struct {
	ProjectID   int64  `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *Registry) createTestCase(ctx context.Context, args createTestCaseArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	r.logger.Info("creating test case", map[string]any{
		"project_id":  args.ProjectID,
		"name":        args.Name,
		"description": args.Description,
	})

	tc, err := r.allure.CreateTestCase(ctx, args.ProjectID, args.Name, args.Description)
	if err != nil {
		r.logger.Error("create test case", err, map[string]any{"project_id": args.ProjectID})
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

type updateTestCaseArgs struct {
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

func (r *Registry) updateTestCase(ctx context.Context, args updateTestCaseArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	hasFields := args.Name != "" || args.Description != "" || args.FullName != "" ||
		args.Precondition != "" || args.ExpectedResult != "" ||
		args.Automated != nil || args.External != nil || args.Deleted != nil ||
		args.StatusID != nil || args.TestLayerID != nil || args.WorkflowID != nil ||
		len(args.Tags) > 0 || len(args.Members) > 0 || len(args.Links) > 0 ||
		len(args.ManualScenario) > 0

	if !hasFields {
		return nil, fmt.Errorf("at least one field must be provided")
	}

	req := allure.UpdateTestCaseRequest{
		Name:           args.Name,
		Description:    args.Description,
		FullName:       args.FullName,
		Precondition:   args.Precondition,
		ExpectedResult: args.ExpectedResult,
		Automated:      args.Automated,
		External:       args.External,
		Deleted:        args.Deleted,
		StatusID:       args.StatusID,
		TestLayerID:    args.TestLayerID,
		WorkflowID:     args.WorkflowID,
		Tags:           args.Tags,
		Members:        args.Members,
		Links:          args.Links,
		Scenario:       args.ManualScenario,
	}

	r.logger.Info("updating test case", map[string]any{"test_case_id": args.TestCaseID})

	if err := r.allure.UpdateTestCase(ctx, args.TestCaseID, req); err != nil {
		r.logger.Error("update test case", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("update test case: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

type deleteTestCaseArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) deleteTestCase(ctx context.Context, args deleteTestCaseArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("deleting test case", map[string]any{"test_case_id": args.TestCaseID})

	if err := r.allure.DeleteTestCase(ctx, args.TestCaseID); err != nil {
		r.logger.Error("delete test case", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("delete test case: %w", err)
	}

	return map[string]any{"status": "deleted"}, nil
}

type cloneTestCaseArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) cloneTestCase(ctx context.Context, args cloneTestCaseArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("cloning test case", map[string]any{"test_case_id": args.TestCaseID})

	newID, err := r.allure.CloneTestCase(ctx, args.TestCaseID)
	if err != nil {
		r.logger.Error("clone test case", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("clone test case: %w", err)
	}

	return map[string]any{"cloned_test_case_id": newID, "status": "cloned"}, nil
}

type restoreTestCaseArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) restoreTestCase(ctx context.Context, args restoreTestCaseArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("restoring test case", map[string]any{"test_case_id": args.TestCaseID})

	if err := r.allure.RestoreTestCase(ctx, args.TestCaseID); err != nil {
		r.logger.Error("restore test case", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("restore test case: %w", err)
	}

	return map[string]any{"status": "restored"}, nil
}

type createTestCaseStepArgs struct {
	TestCaseID int64  `json:"test_case_id"`
	Body       string `json:"body"`
	AfterID    int64  `json:"after_id"`
	ParentID   int64  `json:"parent_id"`
}

func (r *Registry) createTestCaseStep(ctx context.Context, args createTestCaseStepArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.Body == "" {
		return nil, fmt.Errorf("body must be provided")
	}

	req := allure.ScenarioStepCreateRequest{
		TestCaseID: args.TestCaseID,
		Body:       args.Body,
		ParentID:   args.ParentID,
	}

	r.logger.Info("creating test case step", map[string]any{
		"test_case_id": args.TestCaseID,
		"body":         args.Body,
	})

	stepID, err := r.allure.CreateTestCaseStep(ctx, req, args.AfterID)
	if err != nil {
		r.logger.Error("create test case step", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("create test case step: %w", err)
	}

	return map[string]any{"step_id": stepID}, nil
}

type updateTestCaseStepArgs struct {
	StepID         int64  `json:"step_id"`
	Body           string `json:"body"`
	ExpectedResult string `json:"expected_result"`
}

func (r *Registry) updateTestCaseStep(ctx context.Context, args updateTestCaseStepArgs) (any, error) {
	if args.StepID <= 0 {
		return nil, fmt.Errorf("step_id must be positive")
	}
	if args.Body == "" && args.ExpectedResult == "" {
		return nil, fmt.Errorf("at least one field (body or expected_result) must be provided")
	}

	req := allure.ScenarioStepPatchRequest{
		Body:           args.Body,
		ExpectedResult: args.ExpectedResult,
	}

	r.logger.Info("updating test case step", map[string]any{"step_id": args.StepID})

	if err := r.allure.UpdateTestCaseStep(ctx, args.StepID, req); err != nil {
		r.logger.Error("update test case step", err, map[string]any{"step_id": args.StepID})
		return nil, fmt.Errorf("update test case step: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

type deleteTestCaseStepArgs struct {
	StepID int64 `json:"step_id"`
}

func (r *Registry) deleteTestCaseStep(ctx context.Context, args deleteTestCaseStepArgs) (any, error) {
	if args.StepID <= 0 {
		return nil, fmt.Errorf("step_id must be positive")
	}

	r.logger.Info("deleting test case step", map[string]any{"step_id": args.StepID})

	if err := r.allure.DeleteTestCaseStep(ctx, args.StepID); err != nil {
		r.logger.Error("delete test case step", err, map[string]any{"step_id": args.StepID})
		return nil, fmt.Errorf("delete test case step: %w", err)
	}

	return map[string]any{"status": "deleted"}, nil
}

type getTestCaseCustomFieldsArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseCustomFields(ctx context.Context, args getTestCaseCustomFieldsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("fetching test case custom fields", map[string]any{"test_case_id": args.TestCaseID})

	fields, err := r.allure.GetTestCaseCustomFields(ctx, args.TestCaseID)
	if err != nil {
		r.logger.Error("get test case custom fields", err, map[string]any{"test_case_id": args.TestCaseID})
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

type updateTestCaseCustomFieldsArgs struct {
	TestCaseID   int64 `json:"test_case_id"`
	CustomFields []struct {
		CustomFieldID int64   `json:"custom_field_id"`
		ValueIDs      []int64 `json:"value_ids"`
	} `json:"custom_fields"`
}

func (r *Registry) updateTestCaseCustomFields(ctx context.Context, args updateTestCaseCustomFieldsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if len(args.CustomFields) == 0 {
		return nil, fmt.Errorf("custom_fields must contain at least one entry")
	}

	fields := make([]allure.CustomFieldWithValuesDto, len(args.CustomFields))
	for i, cf := range args.CustomFields {
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
		"test_case_id": args.TestCaseID,
		"fields_count": len(fields),
	})

	if err := r.allure.UpdateTestCaseCustomFields(ctx, args.TestCaseID, fields); err != nil {
		r.logger.Error("update test case custom fields", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("update test case custom fields: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

type getTestCaseHistoryArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
}

func (r *Registry) getTestCaseHistory(ctx context.Context, args getTestCaseHistoryArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if args.Size <= 0 {
		args.Size = 10
	}
	if args.Size > 100 {
		args.Size = 100
	}

	r.logger.Info("fetching test case history", map[string]any{
		"test_case_id": args.TestCaseID,
		"page":         args.Page,
		"size":         args.Size,
	})

	history, err := r.allure.GetTestCaseHistory(ctx, args.TestCaseID, args.Page, args.Size)
	if err != nil {
		r.logger.Error("get test case history", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("get test case history: %w", err)
	}

	return history, nil
}
