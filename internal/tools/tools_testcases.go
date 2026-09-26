package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
		Name: "create_test_case",
		Description: "Create a new test case, optionally with all its content in one call: description, precondition, expected result, " +
			"status/workflow/layer, tags, links, members, custom fields and the manual scenario (steps with expected results and sub-steps). " +
			"Only name and project_id are required. Built-in ids are negative (e.g. status -1 Draft; custom fields Epic/Feature/Story/Suite -1..-5).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":      map[string]any{"type": "integer", "description": "Allure project ID"},
				"name":            map[string]any{"type": "string", "description": "Test case name"},
				"description":     map[string]any{"type": "string", "description": "Description (optional, markdown)"},
				"precondition":    map[string]any{"type": "string", "description": "Precondition (optional)"},
				"expected_result": map[string]any{"type": "string", "description": "Overall expected result (optional)"},
				"full_name":       map[string]any{"type": "string", "description": "Full name, e.g. the automated test's qualified name (optional)"},
				"automated":       map[string]any{"type": "boolean", "description": "Mark as automated (optional; default manual)"},
				"status_id":       map[string]any{"type": "integer", "description": "Status ID (optional; requires workflow_id; see get_test_case_workflow)"},
				"workflow_id":     map[string]any{"type": "integer", "description": "Workflow ID (optional; requires status_id)"},
				"test_layer_id":   map[string]any{"type": "integer", "description": "Test layer ID (optional)"},
				"tags":            testTagsSchema,
				"links":           externalLinksSchema,
				"members":         membersSchema,
				"custom_fields": map[string]any{
					"type":        "array",
					"description": "Custom field values (optional). Each item sets one value: custom_field_id plus value_id (existing value) or name (found or created by name). Look values up with list_custom_field_values.",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"custom_field_id": map[string]any{"type": "integer"},
							"value_id":        map[string]any{"type": "integer"},
							"name":            map[string]any{"type": "string"},
						},
						"required": []string{"custom_field_id"},
					},
				},
				"steps": scenarioStepsSchema,
			},
			"required": []string{"project_id", "name"},
		},
		Handler: Typed(r.createTestCase),
	})

	r.register(&Tool{
		Name: "update_test_case",
		Description: "Update any fields of an existing test case: name, description, precondition, " +
			"expected_result, status, tags, members, links, test layer, or the whole manual scenario. " +
			"All fields are optional — only the ones you pass are changed. Pass \"\" to clear a text field and [] to clear tags/members/links; " +
			"tags, members and links replace the whole list. " +
			"manual_scenario REPLACES all current steps (including their attachments) with the given steps, built through the step API " +
			"with real text and expected results; to change one step use update_test_case_step instead.",
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
				"tags":    testTagsSchema,
				"members": membersSchema,
				"links":   externalLinksSchema,
				"manual_scenario": map[string]any{
					"type":        "object",
					"description": "Replace the whole manual scenario: {\"steps\": [{\"body\": \"Open the app\", \"expected_result\": \"Login screen shown\", \"steps\": [...sub-steps]}]}. Pass {\"steps\": []} to remove all steps.",
					"properties": map[string]any{
						"steps": scenarioStepsSchema,
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
		Description: "Add a step to a test case scenario, optionally with its expected result. Appended to the end by default; use after_id/before_id to insert next to a specific step, or parent_id to nest it inside another step. Get existing step IDs with get_test_case_steps. " +
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
				"expected_result": map[string]any{
					"type":        "string",
					"description": "Expected result for the new step (optional; same as a follow-up update_test_case_step)",
				},
				"after_id": map[string]any{
					"type":        "integer",
					"description": "Insert after step ID (optional; its parent is resolved automatically)",
				},
				"before_id": map[string]any{
					"type":        "integer",
					"description": "Insert before step ID (optional; its parent is resolved automatically)",
				},
				"parent_id": map[string]any{
					"type":        "integer",
					"description": "Parent step ID to nest under (optional)",
				},
			},
			"required": []string{"test_case_id", "body"},
		},
		Handler: Typed(r.createTestCaseStep),
	})

	r.register(&Tool{
		Name: "update_test_case_step",
		Description: "Edit a step's text (body) or expected result. Requires the step ID — get it from get_test_case_steps. " +
			"test_case_id is always required: the API models expected results as a separate linked step (not a " +
			"plain field), and this tool needs it both to find/create the entry the web UI actually displays and " +
			"to tell whether a body-only edit would wipe an expected result the step already has. " +
			"See github.com/MimoJanra/TestOpsMCP/issues/16 for the underlying API structure.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"step_id": map[string]any{
					"type":        "integer",
					"description": "Step ID",
				},
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Parent test case ID (always required — see tool description)",
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
			"required": []string{"step_id", "test_case_id"},
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
		Description: "Set custom field values for a test case, replacing whatever values each named field currently " +
			"has (implemented as clear-then-set: the API's dedicated single-case update endpoint is unconditionally " +
			"broken — 500 on any payload, any test case). " +
			"Each item must specify the custom field ID and the values to set — each value needs both id and " +
			"name (the API rejects id-only values with a not-null constraint on the value's name). " +
			"Pass an empty values array for a field to deliberately clear it rather than set it — this is not " +
			"an error, so don't pass an empty array by accident. " +
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
								"description": "Values to assign — each needs both id and name. An empty array clears the field instead of setting it.",
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
			"project_id":        args.ProjectID,
			"status":            tc.Status,
			"automation_status": automationStatus(tc.Automated),
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

	// The API's run-existing-launch endpoint requires a full selection
	// (with the test case's project id), not just the launch/test case ids —
	// see Client.RunTestCase.
	overview, err := r.allure.GetTestCaseOverview(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("look up test case project: %w", err)
	}
	projectID := int64(nodeFloat(overview, "projectId"))
	if projectID <= 0 {
		return nil, fmt.Errorf("test case %d has no projectId in its overview", args.TestCaseID)
	}

	if err := r.allure.RunTestCase(ctx, args.TestCaseID, args.LaunchID, projectID); err != nil {
		r.logger.Error("run test case", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("run test case: %w", err)
	}

	return map[string]any{"status": "started"}, nil
}

// customFieldValueArg sets one custom field value on create: value_id picks an
// existing value, name finds or creates one by name.
type customFieldValueArg struct {
	CustomFieldID int64  `json:"custom_field_id"`
	ValueID       int64  `json:"value_id"`
	Name          string `json:"name"`
}

// scenarioStepArg is one manual step. type="expected" is the legacy flat form
// ({type, body} pairs): such an entry becomes the preceding step's expected
// result rather than a step of its own.
type scenarioStepArg struct {
	Type           string            `json:"type"`
	Body           string            `json:"body"`
	ExpectedResult string            `json:"expected_result"`
	Steps          []scenarioStepArg `json:"steps"`
}

type manualScenarioArg struct {
	Steps []scenarioStepArg `json:"steps"`
}

// normalizeScenarioSteps folds legacy type="expected" entries into the
// preceding step's expected_result and drops empty entries.
func normalizeScenarioSteps(in []scenarioStepArg) ([]scenarioStepArg, error) {
	out := make([]scenarioStepArg, 0, len(in))
	for i, st := range in {
		if st.Type == "expected" {
			if len(out) == 0 {
				return nil, fmt.Errorf("step %d: an expected entry must follow the step it belongs to", i)
			}
			prev := &out[len(out)-1]
			if prev.ExpectedResult != "" {
				prev.ExpectedResult += "\n"
			}
			prev.ExpectedResult += st.Body
			continue
		}
		if strings.TrimSpace(st.Body) == "" {
			return nil, fmt.Errorf("step %d: body must not be empty", i)
		}
		children, err := normalizeScenarioSteps(st.Steps)
		if err != nil {
			return nil, fmt.Errorf("step %d: %w", i, err)
		}
		st.Steps = children
		out = append(out, st)
	}
	return out, nil
}

// buildScenario creates steps (with expected results and sub-steps) under
// parentID through the step API — the only path that reliably stores step
// text: the scenario field of PATCH /api/testcase stores every body as
// "<empty>" and turns expected entries into sibling steps (confirmed live).
func (r *Registry) buildScenario(ctx context.Context, testCaseID, parentID int64, steps []scenarioStepArg) (int, error) {
	created := 0
	for _, st := range steps {
		id, err := r.allure.CreateTestCaseStep(ctx, allure.ScenarioStepCreateRequest{
			TestCaseID: testCaseID,
			Body:       st.Body,
			ParentID:   parentID,
		}, 0)
		if err != nil {
			return created, fmt.Errorf("create step %q: %w", st.Body, err)
		}
		created++
		if st.ExpectedResult != "" {
			if _, err := r.setExpectedResult(ctx, updateTestCaseStepArgs{StepID: id, TestCaseID: testCaseID, ExpectedResult: st.ExpectedResult}); err != nil {
				return created, fmt.Errorf("set expected result of step %q: %w", st.Body, err)
			}
		}
		n, err := r.buildScenario(ctx, testCaseID, id, st.Steps)
		created += n
		if err != nil {
			return created, err
		}
	}
	return created, nil
}

type createTestCaseArgs struct {
	ProjectID      int64                    `json:"project_id"`
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	Precondition   string                   `json:"precondition"`
	ExpectedResult string                   `json:"expected_result"`
	FullName       string                   `json:"full_name"`
	Automated      *bool                    `json:"automated"`
	StatusID       *int64                   `json:"status_id"`
	WorkflowID     *int64                   `json:"workflow_id"`
	TestLayerID    *int64                   `json:"test_layer_id"`
	Tags           []allure.TestTagDto      `json:"tags"`
	Links          []allure.ExternalLinkDto `json:"links"`
	Members        []allure.MemberDto       `json:"members"`
	CustomFields   []customFieldValueArg    `json:"custom_fields"`
	Steps          []scenarioStepArg        `json:"steps"`
}

func (r *Registry) createTestCase(ctx context.Context, args createTestCaseArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if (args.StatusID == nil) != (args.WorkflowID == nil) {
		return nil, fmt.Errorf("status_id and workflow_id must be passed together (the API rejects one without the other); see get_test_case_workflow")
	}
	steps, err := normalizeScenarioSteps(args.Steps)
	if err != nil {
		return nil, fmt.Errorf("steps: %w", err)
	}
	cfs := make([]allure.CustomFieldValueWithCfDto, 0, len(args.CustomFields))
	for i, cf := range args.CustomFields {
		// Built-in fields (Epic, Feature, Suite...) have negative ids.
		if cf.CustomFieldID == 0 {
			return nil, fmt.Errorf("custom_fields[%d]: custom_field_id is required", i)
		}
		if cf.ValueID == 0 && strings.TrimSpace(cf.Name) == "" {
			return nil, fmt.Errorf("custom_fields[%d]: pass value_id or name", i)
		}
		cfs = append(cfs, allure.CustomFieldValueWithCfDto{
			ID:          cf.ValueID,
			Name:        strings.TrimSpace(cf.Name),
			CustomField: allure.CustomFieldRef{ID: cf.CustomFieldID},
		})
	}

	r.logger.Info("creating test case", map[string]any{
		"project_id": args.ProjectID,
		"name":       args.Name,
	})

	tc, err := r.allure.CreateTestCase(ctx, allure.CreateTestCaseRequest{
		Name:           args.Name,
		ProjectID:      args.ProjectID,
		Description:    args.Description,
		Precondition:   args.Precondition,
		ExpectedResult: args.ExpectedResult,
		FullName:       args.FullName,
		Automated:      args.Automated,
		StatusID:       args.StatusID,
		WorkflowID:     args.WorkflowID,
		TestLayerID:    args.TestLayerID,
		Tags:           args.Tags,
		Links:          args.Links,
		Members:        args.Members,
		CustomFields:   cfs,
	})
	if err != nil {
		r.logger.Error("create test case", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("create test case: %w", err)
	}

	result := map[string]any{
		"id":                tc.ID,
		"name":              tc.Name,
		"project_id":        args.ProjectID,
		"description":       tc.Description,
		"status":            tc.Status,
		"automation_status": automationStatus(tc.Automated),
		"full_name":         tc.FullName,
	}
	if len(steps) > 0 {
		n, err := r.buildScenario(ctx, tc.ID, 0, steps)
		result["steps_created"] = n
		if err != nil {
			result["error"] = fmt.Sprintf("test case created, but building its steps failed: %v", err)
		}
	}
	return result, nil
}

type updateTestCaseArgs struct {
	TestCaseID     int64                     `json:"test_case_id"`
	Name           string                    `json:"name"`
	Description    *string                   `json:"description"`
	FullName       *string                   `json:"full_name"`
	Precondition   *string                   `json:"precondition"`
	ExpectedResult *string                   `json:"expected_result"`
	Automated      *bool                     `json:"automated"`
	External       *bool                     `json:"external"`
	Deleted        *bool                     `json:"deleted"`
	StatusID       *int64                    `json:"status_id"`
	TestLayerID    *int64                    `json:"test_layer_id"`
	WorkflowID     *int64                    `json:"workflow_id"`
	Tags           *[]allure.TestTagDto      `json:"tags"`
	Members        *[]allure.MemberDto       `json:"members"`
	Links          *[]allure.ExternalLinkDto `json:"links"`
	ManualScenario *manualScenarioArg        `json:"manual_scenario"`
}

func (r *Registry) updateTestCase(ctx context.Context, args updateTestCaseArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
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
	}
	hasFields := req != (allure.UpdateTestCaseRequest{})
	if !hasFields && args.ManualScenario == nil {
		return nil, fmt.Errorf("at least one field must be provided")
	}

	var steps []scenarioStepArg
	if args.ManualScenario != nil {
		var err error
		if steps, err = normalizeScenarioSteps(args.ManualScenario.Steps); err != nil {
			return nil, fmt.Errorf("manual_scenario: %w", err)
		}
	}

	r.logger.Info("updating test case", map[string]any{"test_case_id": args.TestCaseID})

	result := map[string]any{"status": "updated"}
	if hasFields {
		if err := r.allure.UpdateTestCase(ctx, args.TestCaseID, req); err != nil {
			r.logger.Error("update test case", err, map[string]any{"test_case_id": args.TestCaseID})
			return nil, fmt.Errorf("update test case: %w", err)
		}
	}
	if args.ManualScenario != nil {
		if err := r.allure.DeleteTestCaseScenario(ctx, args.TestCaseID); err != nil {
			return nil, fmt.Errorf("replace scenario: clear current steps: %w", err)
		}
		n, err := r.buildScenario(ctx, args.TestCaseID, 0, steps)
		result["steps_created"] = n
		if err != nil {
			return nil, fmt.Errorf("replace scenario: the old steps were removed and %d new ones created before failing: %w", n, err)
		}
	}
	return result, nil
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
	TestCaseID     int64  `json:"test_case_id"`
	Body           string `json:"body"`
	ExpectedResult string `json:"expected_result"`
	AfterID        int64  `json:"after_id"`
	BeforeID       int64  `json:"before_id"`
	ParentID       int64  `json:"parent_id"`
}

func (r *Registry) createTestCaseStep(ctx context.Context, args createTestCaseStepArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.Body == "" {
		return nil, fmt.Errorf("body must be provided")
	}
	pos, err := r.resolveStepPosition(ctx, args.TestCaseID, args.AfterID, args.BeforeID, args.ParentID)
	if err != nil {
		return nil, err
	}

	req := allure.ScenarioStepCreateRequest{
		TestCaseID: args.TestCaseID,
		Body:       args.Body,
		ParentID:   pos.ParentID,
	}

	r.logger.Info("creating test case step", map[string]any{
		"test_case_id": args.TestCaseID,
		"body":         args.Body,
	})

	stepID, err := r.allure.CreateTestCaseStepAt(ctx, req, pos.AfterID, pos.BeforeID)
	if err != nil {
		r.logger.Error("create test case step", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("create test case step: %w", err)
	}

	result := map[string]any{"step_id": stepID}
	if args.ExpectedResult != "" {
		if _, err := r.setExpectedResult(ctx, updateTestCaseStepArgs{StepID: stepID, TestCaseID: args.TestCaseID, ExpectedResult: args.ExpectedResult}); err != nil {
			result["error"] = fmt.Sprintf("step created, but setting its expected result failed: %v", err)
		}
	}
	return result, nil
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
	// Required unconditionally, not just when setting expected_result: a
	// body-only edit without it can't tell whether the step already has an
	// expected result to preserve, and silently wipes it if so. Confirmed
	// live 2026-09-21 — see github.com/MimoJanra/TestOpsMCP/issues/16.
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id is required: without it this tool can't tell whether the step already has an expected result to preserve, and setting expected_result needs it to find or create the entry the web UI actually displays")
	}

	if args.ExpectedResult != "" {
		return r.setExpectedResult(ctx, args)
	}

	// Body-only edit.
	tree, err := r.allure.GetTestCaseSteps(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("look up current step state: %w", err)
	}
	node := stepNodeFromTree(tree, args.StepID)
	if node == nil {
		return nil, fmt.Errorf("step %d not found under test case %d", args.StepID, args.TestCaseID)
	}
	if nodeInt64(node, "attachmentId") > 0 {
		// A body PATCH on a file/table node silently drops its attachmentId,
		// turning the attachment into an empty text step (confirmed live).
		return nil, fmt.Errorf("step %d is a file/table attachment node, not a text step — delete it with delete_test_case_step instead of editing it", args.StepID)
	}
	// withExpectedResult=true is only safe to send when the step already has
	// an expected result to preserve — sending it on a step with none makes
	// the API spawn a new, empty expected-result container that never
	// existed before. See github.com/MimoJanra/TestOpsMCP/issues/16.
	withExpectedResult := nodeInt64(node, "expectedResultId") > 0

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

	if nodeInt64(node, "attachmentId") > 0 {
		return nil, fmt.Errorf("step %d is a file/table attachment node and can't have an expected result", args.StepID)
	}

	containerID := nodeInt64(node, "expectedResultId")
	if containerID <= 0 {
		// First expected result on this step: this same call also creates the
		// container. The API needs a body in this PATCH; when the caller didn't
		// change it, resend the stored rich-text document rather than the plain
		// body, which would flatten its formatting (confirmed live).
		patch := allure.ScenarioStepPatchRequest{Body: args.Body, ExpectedResult: args.ExpectedResult}
		if args.Body == "" {
			if raw, ok := node["bodyJson"]; ok && raw != nil {
				patch.BodyJSON, _ = json.Marshal(raw)
			}
			if patch.BodyJSON == nil {
				patch.Body = nodeString(node, "body")
			}
		}
		if err := r.allure.UpdateTestCaseStep(ctx, args.StepID, patch, true); err != nil {
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
	} else if args.Body != "" {
		// The container already exists, so the branch above (which also writes
		// body) doesn't run — but a body change was requested, and nothing else
		// in this function ever touches the parent step's own body. Without this,
		// body is silently dropped whenever expected_result is set on a step that
		// already has one (confirmed live 2026-09-21).
		if err := r.allure.UpdateTestCaseStep(ctx, args.StepID, allure.ScenarioStepPatchRequest{Body: args.Body}, true); err != nil {
			return nil, fmt.Errorf("update step body: %w", err)
		}
	}

	// Replace the first text entry. File/table entries (attachment nodes) are
	// skipped: patching one with a body silently drops its attachment.
	textChild := int64(0)
	for _, id := range nodeInt64Array(stepNodeFromTree(tree, containerID), "children") {
		if nodeInt64(stepNodeFromTree(tree, id), "attachmentId") == 0 {
			textChild = id
			break
		}
	}
	if textChild == 0 {
		if _, err := r.allure.CreateTestCaseStep(ctx, allure.ScenarioStepCreateRequest{
			TestCaseID: args.TestCaseID,
			Body:       args.ExpectedResult,
			ParentID:   containerID,
		}, 0); err != nil {
			return nil, fmt.Errorf("create expected-result entry: %w", err)
		}
	} else {
		if err := r.allure.UpdateTestCaseStep(ctx, textChild, allure.ScenarioStepPatchRequest{Body: args.ExpectedResult}, false); err != nil {
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

	// The dedicated GET /api/testcase/{id}/cfv endpoint is unreliable on this
	// API: it returns empty values even for a field confirmed (via get_test_case)
	// to have one, even with the required projectId query param supplied.
	// get_test_case's overview response carries the same data correctly, so
	// this is sourced from there instead. Confirmed live: github.com/MimoJanra/TestOpsMCP/issues/18.
	overview, err := r.allure.GetTestCaseOverview(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case custom fields: %w", err)
	}

	fields := customFieldsFromOverview(overview)
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

// customFieldsFromOverview extracts a test case's current custom field values
// from its overview response (map[string]any, as returned by
// GetTestCaseOverview). The overview's "customFields" array is flat — one row
// per value, each carrying its own nested "customField" — so rows are grouped
// here by custom field ID to match the field-with-nested-values shape used
// elsewhere (CustomFieldWithValuesDto).
func customFieldsFromOverview(overview map[string]any) []allure.CustomFieldWithValuesDto {
	raw, _ := overview["customFields"].([]any)
	order := make([]int64, 0, len(raw))
	byID := make(map[int64]*allure.CustomFieldWithValuesDto, len(raw))
	for _, item := range raw {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cf, _ := row["customField"].(map[string]any)
		fieldID := int64(nodeFloat(cf, "id"))
		f, ok := byID[fieldID]
		if !ok {
			f = &allure.CustomFieldWithValuesDto{CustomField: allure.CustomFieldDto{ID: fieldID, Name: nodeString(cf, "name")}}
			byID[fieldID] = f
			order = append(order, fieldID)
		}
		f.Values = append(f.Values, allure.CustomFieldValueDto{ID: nodeInt64(row, "id"), Name: nodeString(row, "name")})
	}
	result := make([]allure.CustomFieldWithValuesDto, 0, len(order))
	for _, id := range order {
		result = append(result, *byID[id])
	}
	return result
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

	fieldIDs := make([]int64, len(args.CustomFields))
	fields := make([]allure.CustomFieldWithValuesDto, len(args.CustomFields))
	for i, cf := range args.CustomFields {
		for _, v := range cf.Values {
			if v.Name == "" {
				return nil, fmt.Errorf("custom_fields[%d].values: name must be set for value id %d (the API rejects id-only values) — get it via list_custom_field_values", i, v.ID)
			}
		}
		fieldIDs[i] = cf.CustomFieldID
		fields[i] = allure.CustomFieldWithValuesDto{
			CustomField: allure.CustomFieldDto{ID: cf.CustomFieldID},
			Values:      cf.Values,
		}
	}

	overview, err := r.allure.GetTestCaseOverview(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("look up test case project: %w", err)
	}
	projectID := int64(nodeFloat(overview, "projectId"))
	if projectID <= 0 {
		return nil, fmt.Errorf("test case %d has no projectId in its overview", args.TestCaseID)
	}

	r.logger.Info("updating test case custom fields", map[string]any{
		"test_case_id": args.TestCaseID,
		"fields_count": len(fields),
	})

	// Snapshot the fields' current values before touching them, so a failed
	// clear or set below (transient error, bad value on one field) can be
	// rolled back on a best-effort basis instead of leaving them wiped. This
	// is purely an optimistic safety net — if the lookup itself fails, fall
	// back to the old best-effort (no rollback) behavior rather than
	// aborting an update that would otherwise have succeeded.
	// Sourced from the overview response, not GetTestCaseCustomFields (the
	// dedicated GET /cfv endpoint) — that endpoint returns empty values even
	// for a field confirmed to have one, which would make this snapshot
	// silently useless (every field would look already-empty, so "restore"
	// would never restore anything). Confirmed live: github.com/MimoJanra/TestOpsMCP/issues/18.
	originalByID := make(map[int64]allure.CustomFieldWithValuesDto)
	for _, f := range customFieldsFromOverview(overview) {
		originalByID[f.CustomField.ID] = f
	}

	// valueIDsForFields collects the cfv VALUE ids (not custom field ids)
	// currently set for the given field IDs, per a snapshot map keyed by
	// field ID. The v2 bulk remove endpoint's "ids" parameter means cfv value
	// ids — passing a custom field id there is a silent no-op that leaves the
	// value in place while still reporting success (confirmed live, #18).
	valueIDsForFields := func(byID map[int64]allure.CustomFieldWithValuesDto, ids []int64) []int64 {
		var out []int64
		for _, id := range ids {
			if f, ok := byID[id]; ok {
				for _, v := range f.Values {
					out = append(out, v.ID)
				}
			}
		}
		return out
	}

	// restoreOriginalValues is called on any failure below. The bulk
	// endpoints don't guarantee atomicity across rows, so a failed call may
	// have partially applied its changes — re-fetch to find whatever value
	// ids are actually present now (old ones, partially-applied new ones, or
	// a mix) and clear those before restoring whichever fields had a value
	// prior to this update, rather than assuming the failed call had no effect.
	restoreOriginalValues := func(cause error) error {
		currentByID := originalByID
		if freshOverview, err := r.allure.GetTestCaseOverview(ctx, args.TestCaseID); err == nil {
			currentByID = make(map[int64]allure.CustomFieldWithValuesDto)
			for _, f := range customFieldsFromOverview(freshOverview) {
				currentByID[f.CustomField.ID] = f
			}
		}
		if toClear := valueIDsForFields(currentByID, fieldIDs); len(toClear) > 0 {
			if clearErr := r.allure.BulkRemoveTestCaseCustomFields(ctx, projectID, []int64{args.TestCaseID}, toClear); clearErr != nil {
				return fmt.Errorf("%w (rollback also failed, custom fields may be left in a partial state: %v)", cause, clearErr)
			}
		}
		var original []allure.CustomFieldWithValuesDto
		for _, id := range fieldIDs {
			if f, ok := originalByID[id]; ok && len(f.Values) > 0 {
				original = append(original, f)
			}
		}
		if len(original) == 0 {
			return fmt.Errorf("%w (fields were cleared back to their original empty state)", cause)
		}
		if restoreErr := r.allure.BulkAddTestCaseCustomFields(ctx, projectID, []int64{args.TestCaseID}, original); restoreErr != nil {
			return fmt.Errorf("%w (rollback also failed, custom fields may be left empty: %v)", cause, restoreErr)
		}
		return fmt.Errorf("%w (original values were restored)", cause)
	}

	// PATCH /api/testcase/{id}/cfv (the single-case "update" endpoint) is
	// unconditionally broken on this API — even an empty-array body 500s,
	// regardless of test case (github.com/MimoJanra/TestOpsMCP/issues/18).
	// Routed through the bulk v2 endpoints instead, with a single test case ID:
	// clear each field's existing values first, then set the desired ones.
	if toClear := valueIDsForFields(originalByID, fieldIDs); len(toClear) > 0 {
		if err := r.allure.BulkRemoveTestCaseCustomFields(ctx, projectID, []int64{args.TestCaseID}, toClear); err != nil {
			return nil, restoreOriginalValues(fmt.Errorf("clear existing custom field values: %w", err))
		}
	}
	// A field with no values in this request is a deliberate "clear" (see
	// the tool description) — skip the add call entirely when nothing is
	// left to set, since an all-empty cfv row set is otherwise indistinguishable
	// from a caller mistake and best rejected upstream (bulk_add_test_case_custom_fields
	// does exactly that), not silently swallowed by the client.
	hasValuesToSet := false
	for _, f := range fields {
		if len(f.Values) > 0 {
			hasValuesToSet = true
			break
		}
	}
	if hasValuesToSet {
		if err := r.allure.BulkAddTestCaseCustomFields(ctx, projectID, []int64{args.TestCaseID}, fields); err != nil {
			return nil, restoreOriginalValues(fmt.Errorf("set custom field values: %w", err))
		}
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
	if args.CustomFieldID == 0 {
		return nil, fmt.Errorf("custom_field_id is required (built-in fields like Epic/Feature/Story/Suite have negative ids)")
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

// automationStatus renders the API's automated flag. The API has no
// automationStatus field (it always decoded as null), only automated.
func automationStatus(automated bool) string {
	if automated {
		return "automated"
	}
	return "manual"
}

var testTagsSchema = map[string]any{
	"type":        "array",
	"description": "Tags (optional). Each tag by id or name; an unknown name creates the tag.",
	"items": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":   map[string]any{"type": "integer"},
			"name": map[string]any{"type": "string"},
		},
	},
}

var membersSchema = map[string]any{
	"type":        "array",
	"description": "Members (optional). id is a project member id from GET /api/member/suggest?projectId=<id> (via execute_testops_operation).",
	"items": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":   map[string]any{"type": "integer"},
			"name": map[string]any{"type": "string"},
		},
		"required": []string{"id"},
	},
}

var externalLinksSchema = map[string]any{
	"type":        "array",
	"description": "External links (optional)",
	"items": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"type": map[string]any{"type": "string"},
			"url":  map[string]any{"type": "string"},
		},
		"required": []string{"url"},
	},
}

// scenarioStepsSchema describes manual steps. Nesting is recursive: each
// step's "steps" holds sub-steps of the same shape.
var scenarioStepsSchema = map[string]any{
	"type":        "array",
	"description": "Manual steps in order. Each step: body (required), expected_result (optional), steps (optional sub-steps of the same shape).",
	"items": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"body":            map[string]any{"type": "string", "description": "Step action text"},
			"expected_result": map[string]any{"type": "string", "description": "Expected result of this step"},
			"steps":           map[string]any{"type": "array", "description": "Sub-steps (same shape)", "items": map[string]any{"type": "object"}},
		},
		"required": []string{"body"},
	},
}
