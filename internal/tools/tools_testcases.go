package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/session"
)

func (r *Registry) registerTestCaseTools() {
	r.register(&Tool{
		Name:        "list_test_cases",
		Description: "List test cases in a project (returns id, name, status). Paginated — default 10 per page. For filtering by status, tags, or text use search_test_cases instead.",
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
		Name: "get_test_case",
		Description: "Get full details of a test case: name, description, precondition, expected_result, status, tags, members, custom fields, and the manual scenario with its steps. " +
			"Two different scenario representations can appear: `manual_scenario` (the step tree that create_test_case_step/update_test_case_step/get_test_case_steps operate on) and a legacy `scenario` field (older, ID-less steps — present on test cases that predate or were imported outside the manual-scenario feature). " +
			"Check `hasManualScenario`: if it's false but `scenario.steps` is non-empty, this case's real content lives ONLY in the legacy field. " +
			"Adding even one step via create_test_case_step immediately switches the web UI to showing only the (mostly empty) modern tree — the legacy steps become invisible in the UI, though the API still returns them for a while. " +
			"Before adding any step to such a case, recreate ALL of its existing legacy steps (body via create_test_case_step, expected_result via update_test_case_step) in the same pass, so nothing appears to disappear from the UI. Confirmed live 2026-08-27 on tassta.testops.cloud project 170 case 13403.",
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
		Name: "run_test_case",
		Description: "Run a single test case within an existing launch. Both test_case_id and launch_id are required. " +
			"The test case must already be in the launch — call add_test_cases_to_launch first, or this fails with " +
			"409 (field 'selection' must not be null).",
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
		Description: "Create a new test case. Only name and project_id are required. Add description, precondition, and steps after creation with update_test_case and create_test_case_step.",
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
		Name: "update_test_case",
		Description: "Update any fields of an existing test case: name, description, precondition, " +
			"expected_result, status, tags, members, links, or test layer. " +
			"All fields are optional — only the ones you pass are changed. " +
			"WARNING: writing manual_scenario here has been observed to silently corrupt step text — the call " +
			"reports success but every step body is stored as the literal string \"<empty>\" instead of the text " +
			"you sent, with no error. Prefer building the scenario step by step instead: create_test_case_step " +
			"(and update_test_case_step to set each step's expected_result) reliably persists real text. " +
			"Only use manual_scenario here if you verify the result afterwards with get_test_case_steps.",
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
					"description": "The manual test scenario = the list of test steps. Pass {\"steps\": [{\"body\": \"...\"}]} to set all steps at once. This REPLACES the current steps. To add a single step use create_test_case_step.",
					"properties": map[string]any{
						"steps": map[string]any{
							"type":        "array",
							"description": "List of manual steps. Each step needs 'type' (required by API) and 'body'. Use type='body' for a regular step action, type='expected' for an expected-result entry.",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"type": map[string]any{
										"type":        "string",
										"enum":        []string{"body", "expected"},
										"description": "Step type: 'body' for action step, 'expected' for expected result",
									},
									"body": map[string]any{
										"type":        "string",
										"description": "Step text",
									},
								},
								"required": []string{"type", "body"},
							},
						},
					},
					"required": []string{"steps"},
				},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.updateTestCase),
	})

	r.register(&Tool{
		Name:        "delete_test_case",
		Description: "Soft-delete a test case (moves to trash, not permanently removed). Recover with restore_test_case. Use list_deleted_test_cases to browse the trash.",
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
		Description: "Duplicate a test case within the same project, copying all steps, tags, and custom fields. Returns the new test case ID.",
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
		Description: "Recover a soft-deleted test case from trash. Use list_deleted_test_cases first to find the ID of the test case you want to restore.",
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
		Name: "create_test_case_step",
		Description: "Add a step to a test case scenario. Appended to the end by default; use after_id to insert after a specific step, or parent_id to nest it inside another step. Get existing step IDs with get_test_case_steps. " +
			"IMPORTANT: check get_test_case's hasManualScenario field first. If it's false, this case's real steps may live only in the legacy `scenario` field (get_test_case), not in the tree this tool writes to. " +
			"Adding a step here immediately switches the web UI to showing only this tool's step tree — the legacy steps become invisible in the UI, even though they still exist server-side for a while. " +
			"If hasManualScenario is false and get_test_case's `scenario.steps` is non-empty, recreate ALL of those legacy steps here (with update_test_case_step for each expected_result) in the same pass, rather than adding just the one new step you actually wanted.",
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
		Name: "update_test_case_step",
		Description: "Edit a step's text (body) or expected result. Requires the step ID — get it from get_test_case_steps. " +
			"Setting expected_result always requires test_case_id: the API models expected results as a separate " +
			"linked step (not a plain field), and this tool finds or creates the entry that the web UI actually " +
			"displays there. Passing test_case_id on a body-only edit is also recommended so an existing expected " +
			"result isn't wiped. See github.com/MimoJanra/TestOpsMCP/issues/16 for the underlying API structure.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"step_id": map[string]any{
					"type":        "integer",
					"description": "Step ID",
				},
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Parent test case ID. Required when setting expected_result (the API needs it to locate/create the visible entry). Recommended on a body-only edit too, so an existing expected result isn't accidentally wiped.",
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
		Description: "Delete a single step from a test case scenario. Requires the step ID — use get_test_case_steps to find it.",
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
			"Each item must specify the custom field ID and the values to set — each value needs both id and " +
			"name (the API rejects id-only values with a not-null constraint on the value's name). " +
			"Use get_test_case_custom_fields first to discover available fields and their current values, " +
			"and list_custom_field_values to discover valid values (with id and name) for a field (e.g. Priority, Section) " +
			"before setting one, rather than guessing an ID.",
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
							"values": map[string]any{
								"type":        "array",
								"description": "Values to assign — each needs both id and name",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"id":   map[string]any{"type": "integer"},
										"name": map[string]any{"type": "string"},
									},
									"required": []string{"id", "name"},
								},
							},
						},
						"required": []string{"custom_field_id", "values"},
					},
				},
			},
			"required": []string{"test_case_id", "custom_fields"},
		},
		Handler: Typed(r.updateTestCaseCustomFields),
	})

	r.register(&Tool{
		Name: "list_custom_field_values",
		Description: "List the valid values defined for a custom field within a project (e.g. the allowed " +
			"Priority or Section options), so you can pick a real value ID instead of guessing one. " +
			"Get the custom_field_id from get_test_case_custom_fields on any test case in the project.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"custom_field_id": map[string]any{
					"type":        "integer",
					"description": "Custom field ID",
				},
				"query": map[string]any{
					"type":        "string",
					"description": "Optional filter on value name",
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "Zero-based page index (default 0)",
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "Page size (default 10)",
				},
			},
			"required": []string{"project_id", "custom_field_id"},
		},
		Handler: Typed(r.listCustomFieldValues),
	})

	r.register(&Tool{
		Name:        "get_test_case_history",
		Description: "Get the change log for a test case: who changed which field and when. Useful for auditing unexpected modifications.",
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

	// Uses the same normalized-step endpoint as get_test_case_steps. The old
	// dedicated scenario endpoint (GET /api/testcase/{id}/scenario) is marked
	// deprecated in the API spec and always returns an empty step list.
	scenario, err := r.allure.GetTestCaseSteps(ctx, args.TestCaseID)
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
	ManualScenario *allure.ScenarioDto      `json:"manual_scenario"`
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
		args.ManualScenario != nil

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

	elicit, ok := session.ElicitFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("deletion requires user confirmation but no interactive session is available")
	}
	schema, _ := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"confirmed": map[string]any{"type": "boolean", "description": "Confirm permanent deletion"},
		},
	})
	result, err := elicit(ctx, fmt.Sprintf("Permanently delete test case #%d? This cannot be undone.", args.TestCaseID), schema)
	if err != nil {
		return nil, fmt.Errorf("confirmation failed: %w", err)
	}
	if result.Action != "accept" {
		return map[string]any{"cancelled": true, "message": "Deletion cancelled."}, nil
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
	TestCaseID     int64  `json:"test_case_id"`
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

	if args.ExpectedResult != "" {
		// Setting expected_result needs test_case_id: the API models it as a
		// separate "container" step (linked via expectedResultId) whose own body
		// the web UI does not display — the UI instead renders a list of the
		// container's child steps. See github.com/MimoJanra/TestOpsMCP/issues/16.
		if args.TestCaseID <= 0 {
			return nil, fmt.Errorf("test_case_id is required to set expected_result: needed both to resend the current body (an expected_result-only PATCH is rejected) and to find or create the entry the web UI actually displays")
		}
		return r.setExpectedResult(ctx, args)
	}

	// Body-only edit.
	withExpectedResult := false
	if args.TestCaseID > 0 {
		tree, err := r.allure.GetTestCaseSteps(ctx, args.TestCaseID)
		if err != nil {
			return nil, fmt.Errorf("look up current step state: %w", err)
		}
		node := stepNodeFromTree(tree, args.StepID)
		if node == nil {
			return nil, fmt.Errorf("step %d not found under test case %d", args.StepID, args.TestCaseID)
		}
		// withExpectedResult=true is only safe to send when the step already has
		// an expected result to preserve — sending it on a step with none makes
		// the API spawn a new, empty expected-result container that never
		// existed before. See github.com/MimoJanra/TestOpsMCP/issues/16.
		if nodeInt64(node, "expectedResultId") > 0 {
			withExpectedResult = true
		}
	}

	r.logger.Info("updating test case step", map[string]any{"step_id": args.StepID})

	if err := r.allure.UpdateTestCaseStep(ctx, args.StepID, allure.ScenarioStepPatchRequest{Body: args.Body}, withExpectedResult); err != nil {
		r.logger.Error("update test case step", err, map[string]any{"step_id": args.StepID})
		return nil, fmt.Errorf("update test case step: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

// setExpectedResult sets a step's single visible expected-result entry.
//
// The API represents expected results as a separate "container" step, linked
// from the action step via expectedResultId. The container's own body is NOT
// what the web UI displays; the UI instead shows every one of the container's
// child steps as a separate visible expected-result entry (Allure supports
// multiple expected results per step). Verified live on 2026-08-26 against
// tassta.testops.cloud project 408: typing into the UI's Expected Result field
// added a new child under the container rather than editing the container or
// any existing child — see github.com/MimoJanra/TestOpsMCP/issues/16.
//
// To behave like "set the expected result" (one value, replaceable) rather
// than "append another one", this replaces the first existing child's text if
// the container already has children, and creates exactly one child otherwise.
func (r *Registry) setExpectedResult(ctx context.Context, args updateTestCaseStepArgs) (any, error) {
	tree, err := r.allure.GetTestCaseSteps(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("look up current step state: %w", err)
	}
	node := stepNodeFromTree(tree, args.StepID)
	if node == nil {
		return nil, fmt.Errorf("step %d not found under test case %d", args.StepID, args.TestCaseID)
	}

	body := args.Body
	if body == "" {
		body = nodeString(node, "body")
	}

	containerID := nodeInt64(node, "expectedResultId")
	if containerID <= 0 {
		// First expected result on this step: create the container.
		if err := r.allure.UpdateTestCaseStep(ctx, args.StepID, allure.ScenarioStepPatchRequest{
			Body:           body,
			ExpectedResult: args.ExpectedResult,
		}, true); err != nil {
			return nil, fmt.Errorf("create expected-result container: %w", err)
		}
		tree, err = r.allure.GetTestCaseSteps(ctx, args.TestCaseID)
		if err != nil {
			return nil, fmt.Errorf("re-fetch after creating expected-result container: %w", err)
		}
		node = stepNodeFromTree(tree, args.StepID)
		if node == nil {
			return nil, fmt.Errorf("step %d disappeared after update", args.StepID)
		}
		containerID = nodeInt64(node, "expectedResultId")
		if containerID <= 0 {
			return nil, fmt.Errorf("expected-result container was not created")
		}
	}

	children := nodeInt64Array(stepNodeFromTree(tree, containerID), "children")
	if len(children) == 0 {
		if _, err := r.allure.CreateTestCaseStep(ctx, allure.ScenarioStepCreateRequest{
			TestCaseID: args.TestCaseID,
			Body:       args.ExpectedResult,
			ParentID:   containerID,
		}, 0); err != nil {
			return nil, fmt.Errorf("create expected-result entry: %w", err)
		}
	} else {
		if err := r.allure.UpdateTestCaseStep(ctx, children[0], allure.ScenarioStepPatchRequest{Body: args.ExpectedResult}, false); err != nil {
			return nil, fmt.Errorf("update expected-result entry: %w", err)
		}
	}

	return map[string]any{"status": "updated"}, nil
}

// stepNodeFromTree looks up a step node by ID in a NormalizedScenarioDto tree
// (as returned by GetTestCaseSteps), checking both the scenarioSteps map and
// the root node.
func stepNodeFromTree(tree map[string]any, stepID int64) map[string]any {
	key := strconv.FormatInt(stepID, 10)
	if steps, ok := tree["scenarioSteps"].(map[string]any); ok {
		if node, ok := steps[key].(map[string]any); ok {
			return node
		}
	}
	if root, ok := tree["root"].(map[string]any); ok && int64(nodeFloat(root, "id")) == stepID {
		return root
	}
	return nil
}

func nodeFloat(node map[string]any, field string) float64 {
	v, _ := node[field].(float64)
	return v
}

func nodeInt64(node map[string]any, field string) int64 {
	return int64(nodeFloat(node, field))
}

func nodeInt64Array(node map[string]any, field string) []int64 {
	if node == nil {
		return nil
	}
	arr, _ := node[field].([]any)
	out := make([]int64, 0, len(arr))
	for _, v := range arr {
		if f, ok := v.(float64); ok {
			out = append(out, int64(f))
		}
	}
	return out
}

func nodeString(node map[string]any, field string) string {
	v, _ := node[field].(string)
	return v
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
		CustomFieldID int64                        `json:"custom_field_id"`
		Values        []allure.CustomFieldValueDto `json:"values"`
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
		for _, v := range cf.Values {
			if v.Name == "" {
				return nil, fmt.Errorf("custom_fields[%d].values: name must be set for value id %d (the API rejects id-only values) — get it via list_custom_field_values", i, v.ID)
			}
		}
		fields[i] = allure.CustomFieldWithValuesDto{
			CustomField: allure.CustomFieldDto{ID: cf.CustomFieldID},
			Values:      cf.Values,
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

type listCustomFieldValuesArgs struct {
	ProjectID     int64  `json:"project_id"`
	CustomFieldID int64  `json:"custom_field_id"`
	Query         string `json:"query"`
	Page          int    `json:"page"`
	Size          int    `json:"size"`
}

func (r *Registry) listCustomFieldValues(ctx context.Context, args listCustomFieldValuesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.CustomFieldID <= 0 {
		return nil, fmt.Errorf("custom_field_id must be positive")
	}
	size := args.Size
	if size <= 0 {
		size = 10
	}

	result, err := r.allure.ListCustomFieldValues(ctx, args.ProjectID, args.CustomFieldID, args.Query, args.Page, size)
	if err != nil {
		return nil, fmt.Errorf("list custom field values: %w", err)
	}
	return result, nil
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
