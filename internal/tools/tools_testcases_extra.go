package tools

import (
	"context"
	"fmt"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
)

// registerTestCaseExtraTools registers all remaining test-case-specific tools:
// tags, issues, examples, versions, attachments, search, deleted list,
// step move/copy, scenario delete, and test-case relations.
func (r *Registry) registerTestCaseExtraTools() {
	// ── Tags ──────────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_tags",
		Description: "Get all tags assigned to a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseTags),
	})

	r.register(&Tool{
		Name:        "set_test_case_tags",
		Description: "Replace all tags on a test case (full replace — pass complete desired list)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"tags": map[string]any{
					"type":        "array",
					"description": "Complete list of tags to set. Each tag needs an existing id — the API rejects a name-only entry for a tag that doesn't exist yet with 409 (field 'id' must not be null or empty). Use create_test_tag first to create a new tag and get its id.",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":   map[string]any{"type": "integer"},
							"name": map[string]any{"type": "string"},
						},
					},
				},
			},
			"required": []string{"test_case_id", "tags"},
		},
		Handler: Typed(r.setTestCaseTags),
	})

	r.register(&Tool{
		Name: "create_test_tag",
		Description: "Create a new project-wide tag by name and return its ID. " +
			"Use this first when you need a tag that doesn't exist yet, then attach the returned id " +
			"via set_test_case_tags or bulk_add_test_case_tags.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string", "description": "Tag name"},
			},
			"required": []string{"name"},
		},
		Handler: Typed(r.createTestTag),
	})

	// ── Issues ────────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_issues",
		Description: "Get all issues (bug tracker links) linked to a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseIssues),
	})

	r.register(&Tool{
		Name:        "set_test_case_issues",
		Description: "Replace all issues linked to a test case (full replace). Each issue requires integrationId and displayName.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"issues": map[string]any{
					"type":        "array",
					"description": "Complete list of issues to link",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":             map[string]any{"type": "integer", "description": "Issue ID (if known)"},
							"display_name":   map[string]any{"type": "string", "description": "Issue display name / key (e.g. PROJ-123)"},
							"url":            map[string]any{"type": "string", "description": "Direct URL to the issue"},
							"integration_id": map[string]any{"type": "integer", "description": "Bug tracker integration ID"},
							"closed":         map[string]any{"type": "boolean"},
						},
					},
				},
			},
			"required": []string{"test_case_id", "issues"},
		},
		Handler: Typed(r.setTestCaseIssues),
	})

	// ── Examples (parametrized) ───────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_examples",
		Description: "Get the parametrized data table for a test case — each row is one set of input/output values used for data-driven test runs.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseExamples),
	})

	r.register(&Tool{
		Name: "set_test_case_examples",
		Description: "Replace parametrized test examples for a test case. " +
			"Each example is a row of key-value parameter pairs.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"examples": map[string]any{
					"type":        "array",
					"description": "List of example rows; each row is an array of {name, value} pairs",
					"items": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name":  map[string]any{"type": "string"},
								"value": map[string]any{"type": "string"},
							},
							"required": []string{"name", "value"},
						},
					},
				},
			},
			"required": []string{"test_case_id", "examples"},
		},
		Handler: Typed(r.setTestCaseExamples),
	})

	// ── Versions ──────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "list_test_case_versions",
		Description: "List all saved versions (snapshots) of a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.listTestCaseVersions),
	})

	r.register(&Tool{
		Name:        "create_test_case_version",
		Description: "Save a named snapshot of the current test case state. Create a version before making bulk edits so you can restore with restore_test_case_version if needed.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"title":        map[string]any{"type": "string", "description": "Version title (required)"},
				"description":  map[string]any{"type": "string", "description": "Version description (optional)"},
			},
			"required": []string{"test_case_id", "title"},
		},
		Handler: Typed(r.createTestCaseVersion),
	})

	r.register(&Tool{
		Name:        "restore_test_case_version",
		Description: "Restore a test case to a specific saved version",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"version_id": map[string]any{"type": "integer", "description": "Version ID (from list_test_case_versions)"},
			},
			"required": []string{"version_id"},
		},
		Handler: Typed(r.restoreTestCaseVersion),
	})

	// ── Attachments ───────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_attachments",
		Description: "List attachments of a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"page":         map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":         map[string]any{"type": "integer", "description": "Items per page", "default": 20},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseAttachments),
	})

	r.register(&Tool{
		Name:        "delete_test_case_attachment",
		Description: "Delete an attachment from a test case by attachment ID",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"attachment_id": map[string]any{"type": "integer", "description": "Attachment ID"},
			},
			"required": []string{"attachment_id"},
		},
		Handler: Typed(r.deleteTestCaseAttachment),
	})

	// ── Search & filtered lists ───────────────────────────────────────────────

	r.register(&Tool{
		Name: "search_test_cases",
		Description: "Search test cases in a project using an AQL query. " +
			"String literals MUST be single-quoted — name ~ \"login\" (double quotes) returns 400 Invalid AQL; " +
			"use name ~ 'login' instead. Example queries: name ~ 'login', status = 'active', tag = 'smoke'. " +
			"If a query 400s and the cause isn't obvious, call validate_test_case_query first to check the syntax " +
			"before assuming the field/operator is wrong — suggest_test_cases is also available as a fallback " +
			"for natural-language-style lookups.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"query":      map[string]any{"type": "string", "description": "AQL/RQL filter expression"},
				"page":       map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":       map[string]any{"type": "integer", "description": "Items per page", "default": 20},
			},
			"required": []string{"project_id", "query"},
		},
		Handler: Typed(r.searchTestCases),
	})

	r.register(&Tool{
		Name:        "list_deleted_test_cases",
		Description: "List deleted test cases in a project (use restore_test_case to recover them)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"page":       map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":       map[string]any{"type": "integer", "description": "Items per page", "default": 20},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.listDeletedTestCases),
	})

	// ── Scenario / step position ──────────────────────────────────────────────

	r.register(&Tool{
		Name: "get_test_case_scenario",
		Description: "Read the scenario of a test case: ordered steps with names, keywords, expected results, and nesting " +
			"(same underlying data as get_test_case_steps, including step IDs).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseScenario),
	})

	r.register(&Tool{
		Name: "get_test_case_steps",
		Description: "Read the normalized step list for a test case, including the ID of every step. " +
			"Use this when you need a step ID to call move_test_case_step, copy_test_case_step, or delete_test_case_step. " +
			"For a readable view of the scenario without IDs, use get_test_case_scenario.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseSteps),
	})

	r.register(&Tool{
		Name:        "delete_test_case_scenario",
		Description: "Remove all steps from a test case, leaving the scenario empty. Cannot be undone — save a version with create_test_case_version first if you may need to restore.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.deleteTestCaseScenario),
	})

	r.register(&Tool{
		Name:        "move_test_case_step",
		Description: "Reorder, nest, or relocate a step within a scenario: place it before/after a sibling or under a parent. Get step IDs with get_test_case_steps.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"step_id":   map[string]any{"type": "integer", "description": "Step ID to move"},
				"after_id":  map[string]any{"type": "integer", "description": "Place after this step ID (optional)"},
				"before_id": map[string]any{"type": "integer", "description": "Place before this step ID (optional)"},
				"parent_id": map[string]any{"type": "integer", "description": "New parent step ID for nesting (optional)"},
			},
			"required": []string{"step_id"},
		},
		Handler: Typed(r.moveTestCaseStep),
	})

	r.register(&Tool{
		Name:        "copy_test_case_step",
		Description: "Duplicate a step at a new position in the scenario. Get step IDs with get_test_case_steps.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"step_id":   map[string]any{"type": "integer", "description": "Step ID to copy"},
				"after_id":  map[string]any{"type": "integer", "description": "Place copy after this step ID (optional)"},
				"before_id": map[string]any{"type": "integer", "description": "Place copy before this step ID (optional)"},
				"parent_id": map[string]any{"type": "integer", "description": "Parent step ID for nesting (optional)"},
			},
			"required": []string{"step_id"},
		},
		Handler: Typed(r.copyTestCaseStep),
	})

	// ── Relations ─────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_relations",
		Description: "Get test-case-to-test-case relations (e.g. 'blocks', 'is blocked by')",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseRelations),
	})

	r.register(&Tool{
		Name:        "set_test_case_relations",
		Description: "Set test-case-to-test-case relations (e.g. 'blocks', 'is blocked by'). Full replace — pass the complete list you want to keep.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"relations": map[string]any{
					"type":        "array",
					"description": "List of relations to set",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"target_id":   map[string]any{"type": "integer", "description": "Target test case ID"},
							"target_name": map[string]any{"type": "string", "description": "Target test case name (optional)"},
						},
						"required": []string{"target_id"},
					},
				},
			},
			"required": []string{"test_case_id", "relations"},
		},
		Handler: Typed(r.setTestCaseRelations),
	})

	// ── Muted list ────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "list_muted_test_cases",
		Description: "List test cases that are currently muted in a project. Muted test cases are excluded from pass rate calculations and don't trigger failure alerts.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"page":       map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":       map[string]any{"type": "integer", "description": "Items per page", "default": 20},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.listMutedTestCases),
	})

	// ── Audit log ─────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_audit",
		Description: "Get audit log (change history) for a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"page":         map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":         map[string]any{"type": "integer", "description": "Items per page", "default": 20},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseAudit),
	})

	// ── Query tools ───────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "validate_test_case_query",
		Description: "Check whether an AQL/RQL filter query is syntactically valid without executing it. Use before search_test_cases when building dynamic queries programmatically.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"query":      map[string]any{"type": "string", "description": "AQL/RQL expression to validate"},
			},
			"required": []string{"project_id", "query"},
		},
		Handler: Typed(r.validateTestCaseQuery),
	})

	r.register(&Tool{
		Name:        "suggest_test_cases",
		Description: "Get name-based autocomplete suggestions for a partial test case name. Use to help the user find the right test case before fetching its full details.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"query":      map[string]any{"type": "string", "description": "Search string"},
				"page":       map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":       map[string]any{"type": "integer", "description": "Items per page", "default": 10},
			},
			"required": []string{"project_id", "query"},
		},
		Handler: Typed(r.suggestTestCases),
	})

	// ── Workflow ──────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_workflow",
		Description: "Get the workflow assigned to a test case: its states, transitions, and the current position. Useful for knowing which status changes are allowed next.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseWorkflow),
	})

	// ── Test keys ─────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_keys",
		Description: "Get integration test keys linked to a test case (e.g. Jira, Azure DevOps keys)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseKeys),
	})

	r.register(&Tool{
		Name:        "set_test_case_keys",
		Description: "Set integration test keys for a test case (full replace)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"keys": map[string]any{
					"type":        "array",
					"description": "List of test keys",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":             map[string]any{"type": "integer"},
							"integration_id": map[string]any{"type": "integer"},
							"name":           map[string]any{"type": "string"},
							"url":            map[string]any{"type": "string"},
						},
					},
				},
			},
			"required": []string{"test_case_id", "keys"},
		},
		Handler: Typed(r.setTestCaseKeys),
	})

	// ── Scenario from run ─────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_scenario_from_run",
		Description: "Get the steps actually executed during the last automated test run for this test case. Useful for comparing the designed scenario against what automation actually ran.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.getTestCaseScenarioFromRun),
	})

	// ── Automation ────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "detach_test_case_automation",
		Description: "Unlink a test case from its automation class, converting it back to a manual test. Optionally set a new status after detaching.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"status_id":    map[string]any{"type": "integer", "description": "Status ID to set after detach (optional)"},
				"workflow_id":  map[string]any{"type": "integer", "description": "Workflow ID to apply (optional)"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.detachTestCaseAutomation),
	})

	// ── Version data & delete ─────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_version_data",
		Description: "Get the full test case overview data for a specific version snapshot",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"version_id": map[string]any{"type": "integer", "description": "Version ID (from list_test_case_versions)"},
			},
			"required": []string{"version_id"},
		},
		Handler: Typed(r.getTestCaseVersionData),
	})

	r.register(&Tool{
		Name:        "delete_test_case_version",
		Description: "Delete a specific version snapshot of a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"version_id": map[string]any{"type": "integer", "description": "Version ID to delete"},
			},
			"required": []string{"version_id"},
		},
		Handler: Typed(r.deleteTestCaseVersion),
	})
}

// ── Handlers ──────────────────────────────────────────────────────────────────

type getTestCaseTagsArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseTags(ctx context.Context, args getTestCaseTagsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	tags, err := r.allure.GetTestCaseTags(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case tags: %w", err)
	}
	items := make([]map[string]any, len(tags))
	for i, t := range tags {
		items[i] = map[string]any{"id": t.ID, "name": t.Name}
	}
	return map[string]any{"tags": items}, nil
}

type setTestCaseTagsArgs struct {
	TestCaseID int64               `json:"test_case_id"`
	Tags       []allure.TestTagDto `json:"tags"`
}

func (r *Registry) setTestCaseTags(ctx context.Context, args setTestCaseTagsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if err := r.allure.SetTestCaseTags(ctx, args.TestCaseID, args.Tags); err != nil {
		return nil, fmt.Errorf("set test case tags: %w", err)
	}
	return map[string]any{"status": "updated", "count": len(args.Tags)}, nil
}

type createTestTagArgs struct {
	Name string `json:"name"`
}

func (r *Registry) createTestTag(ctx context.Context, args createTestTagArgs) (any, error) {
	if args.Name == "" {
		return nil, fmt.Errorf("name must be provided")
	}
	tag, err := r.allure.CreateTestTag(ctx, args.Name)
	if err != nil {
		return nil, fmt.Errorf("create test tag: %w", err)
	}
	return map[string]any{"id": tag.ID, "name": tag.Name}, nil
}

type getTestCaseIssuesArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseIssues(ctx context.Context, args getTestCaseIssuesArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	issues, err := r.allure.GetTestCaseIssues(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case issues: %w", err)
	}
	items := make([]map[string]any, len(issues))
	for i, iss := range issues {
		items[i] = map[string]any{
			"id":             iss.ID,
			"display_name":   iss.DisplayName,
			"url":            iss.URL,
			"integration_id": iss.IntegrationID,
			"closed":         iss.Closed,
		}
	}
	return map[string]any{"issues": items}, nil
}

type setTestCaseIssuesArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	Issues     []struct {
		ID            int64  `json:"id"`
		DisplayName   string `json:"display_name"`
		URL           string `json:"url"`
		IntegrationID int64  `json:"integration_id"`
		Closed        bool   `json:"closed"`
	} `json:"issues"`
}

func (r *Registry) setTestCaseIssues(ctx context.Context, args setTestCaseIssuesArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	issues := make([]allure.IssueDto, len(args.Issues))
	for i, iss := range args.Issues {
		issues[i] = allure.IssueDto{
			ID:            iss.ID,
			DisplayName:   iss.DisplayName,
			URL:           iss.URL,
			IntegrationID: iss.IntegrationID,
			Closed:        iss.Closed,
		}
	}
	if err := r.allure.SetTestCaseIssues(ctx, args.TestCaseID, issues); err != nil {
		return nil, fmt.Errorf("set test case issues: %w", err)
	}
	return map[string]any{"status": "updated", "count": len(issues)}, nil
}

type getTestCaseExamplesArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseExamples(ctx context.Context, args getTestCaseExamplesArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	examples, err := r.allure.GetTestCaseExamples(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case examples: %w", err)
	}
	return map[string]any{"examples": examples, "count": len(examples)}, nil
}

type setTestCaseExamplesArgs struct {
	TestCaseID int64                          `json:"test_case_id"`
	Examples   [][]allure.TestCaseExampleParam `json:"examples"`
}

func (r *Registry) setTestCaseExamples(ctx context.Context, args setTestCaseExamplesArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if err := r.allure.SetTestCaseExamples(ctx, args.TestCaseID, args.Examples); err != nil {
		return nil, fmt.Errorf("set test case examples: %w", err)
	}
	return map[string]any{"status": "updated", "rows": len(args.Examples)}, nil
}

type listTestCaseVersionsArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) listTestCaseVersions(ctx context.Context, args listTestCaseVersionsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	versions, err := r.allure.ListTestCaseVersions(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("list test case versions: %w", err)
	}
	items := make([]map[string]any, len(versions))
	for i, v := range versions {
		items[i] = map[string]any{
			"id":           v.ID,
			"title":        v.Title,
			"description":  v.Description,
			"created_date": v.CreatedDate,
			"created_by":   v.CreatedBy,
		}
	}
	return map[string]any{"versions": items}, nil
}

type createTestCaseVersionArgs struct {
	TestCaseID  int64  `json:"test_case_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (r *Registry) createTestCaseVersion(ctx context.Context, args createTestCaseVersionArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	version, err := r.allure.CreateTestCaseVersion(ctx, args.TestCaseID, allure.TestCaseVersionCreateRequest{
		Title:       args.Title,
		Description: args.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("create test case version: %w", err)
	}
	return map[string]any{
		"id":          version.ID,
		"title":       version.Title,
		"description": version.Description,
	}, nil
}

type restoreTestCaseVersionArgs struct {
	VersionID int64 `json:"version_id"`
}

func (r *Registry) restoreTestCaseVersion(ctx context.Context, args restoreTestCaseVersionArgs) (any, error) {
	if args.VersionID <= 0 {
		return nil, fmt.Errorf("version_id must be positive")
	}
	if err := r.allure.RestoreTestCaseVersion(ctx, args.VersionID); err != nil {
		return nil, fmt.Errorf("restore test case version: %w", err)
	}
	return map[string]any{"status": "restored", "version_id": args.VersionID}, nil
}

type getTestCaseAttachmentsArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
}

func (r *Registry) getTestCaseAttachments(ctx context.Context, args getTestCaseAttachmentsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.Size <= 0 {
		args.Size = 20
	}
	result, err := r.allure.GetTestCaseAttachments(ctx, args.TestCaseID, args.Page, args.Size)
	if err != nil {
		return nil, fmt.Errorf("get test case attachments: %w", err)
	}
	items := make([]map[string]any, len(result.Content))
	for i, a := range result.Content {
		items[i] = map[string]any{
			"id":           a.ID,
			"name":         a.Name,
			"content_type": a.ContentType,
			"size":         a.Size,
			"created_date": a.CreatedDate,
		}
	}
	return map[string]any{
		"attachments": items,
		"page":        result.Number,
		"size":        result.Size,
		"total":       result.Total,
		"is_last":     result.Last,
	}, nil
}

type deleteTestCaseAttachmentArgs struct {
	AttachmentID int64 `json:"attachment_id"`
}

func (r *Registry) deleteTestCaseAttachment(ctx context.Context, args deleteTestCaseAttachmentArgs) (any, error) {
	if args.AttachmentID <= 0 {
		return nil, fmt.Errorf("attachment_id must be positive")
	}
	if err := r.allure.DeleteTestCaseAttachment(ctx, args.AttachmentID); err != nil {
		return nil, fmt.Errorf("delete test case attachment: %w", err)
	}
	return map[string]any{"status": "deleted"}, nil
}

type searchTestCasesArgs struct {
	ProjectID int64  `json:"project_id"`
	Query     string `json:"query"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

func (r *Registry) searchTestCases(ctx context.Context, args searchTestCasesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.Query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if args.Size <= 0 {
		args.Size = 20
	}
	if args.Size > 100 {
		args.Size = 100
	}
	cases, err := r.allure.SearchTestCases(ctx, args.ProjectID, args.Query, args.Page, args.Size)
	if err != nil {
		return nil, fmt.Errorf("search test cases: %w", err)
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

type listDeletedTestCasesArgs struct {
	ProjectID int64 `json:"project_id"`
	Page      int   `json:"page"`
	Size      int   `json:"size"`
}

func (r *Registry) listDeletedTestCases(ctx context.Context, args listDeletedTestCasesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.Size <= 0 {
		args.Size = 20
	}
	if args.Size > 100 {
		args.Size = 100
	}
	cases, err := r.allure.ListDeletedTestCases(ctx, args.ProjectID, args.Page, args.Size)
	if err != nil {
		return nil, fmt.Errorf("list deleted test cases: %w", err)
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

type getTestCaseScenarioArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseScenario(ctx context.Context, args getTestCaseScenarioArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	// Uses the same normalized-step endpoint as get_test_case_steps. The old
	// dedicated scenario endpoint (GET /api/testcase/{id}/scenario) is marked
	// deprecated in the API spec and always returns an empty step list.
	result, err := r.allure.GetTestCaseSteps(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case scenario: %w", err)
	}
	return result, nil
}

type getTestCaseStepsArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseSteps(ctx context.Context, args getTestCaseStepsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	result, err := r.allure.GetTestCaseSteps(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case steps: %w", err)
	}
	return result, nil
}

type deleteTestCaseScenarioArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) deleteTestCaseScenario(ctx context.Context, args deleteTestCaseScenarioArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if err := r.allure.DeleteTestCaseScenario(ctx, args.TestCaseID); err != nil {
		return nil, fmt.Errorf("delete test case scenario: %w", err)
	}
	return map[string]any{"status": "deleted"}, nil
}

type moveTestCaseStepArgs struct {
	StepID   int64 `json:"step_id"`
	AfterID  int64 `json:"after_id"`
	BeforeID int64 `json:"before_id"`
	ParentID int64 `json:"parent_id"`
}

func (r *Registry) moveTestCaseStep(ctx context.Context, args moveTestCaseStepArgs) (any, error) {
	if args.StepID <= 0 {
		return nil, fmt.Errorf("step_id must be positive")
	}
	pos := allure.StepPositionDto{AfterID: args.AfterID, BeforeID: args.BeforeID, ParentID: args.ParentID}
	if err := r.allure.MoveTestCaseStep(ctx, args.StepID, pos); err != nil {
		return nil, fmt.Errorf("move test case step: %w", err)
	}
	return map[string]any{"status": "moved"}, nil
}

type copyTestCaseStepArgs struct {
	StepID   int64 `json:"step_id"`
	AfterID  int64 `json:"after_id"`
	BeforeID int64 `json:"before_id"`
	ParentID int64 `json:"parent_id"`
}

func (r *Registry) copyTestCaseStep(ctx context.Context, args copyTestCaseStepArgs) (any, error) {
	if args.StepID <= 0 {
		return nil, fmt.Errorf("step_id must be positive")
	}
	pos := allure.StepPositionDto{AfterID: args.AfterID, BeforeID: args.BeforeID, ParentID: args.ParentID}
	if err := r.allure.CopyTestCaseStep(ctx, args.StepID, pos); err != nil {
		return nil, fmt.Errorf("copy test case step: %w", err)
	}
	return map[string]any{"status": "copied"}, nil
}

type getTestCaseRelationsArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseRelations(ctx context.Context, args getTestCaseRelationsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	relations, err := r.allure.GetTestCaseRelations(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case relations: %w", err)
	}
	items := make([]map[string]any, len(relations))
	for i, rel := range relations {
		items[i] = map[string]any{
			"id":          rel.ID,
			"target_id":   rel.Target.ID,
			"target_name": rel.Target.Name,
		}
	}
	return map[string]any{"relations": items}, nil
}

type setTestCaseRelationsArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	Relations  []struct {
		TargetID   int64  `json:"target_id"`
		TargetName string `json:"target_name"`
	} `json:"relations"`
}

func (r *Registry) setTestCaseRelations(ctx context.Context, args setTestCaseRelationsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	relations := make([]allure.RelationDto, len(args.Relations))
	for i, rel := range args.Relations {
		relations[i] = allure.RelationDto{
			Target: allure.RelationTargetDto{ID: rel.TargetID, Name: rel.TargetName},
		}
	}
	if err := r.allure.SetTestCaseRelations(ctx, args.TestCaseID, relations); err != nil {
		return nil, fmt.Errorf("set test case relations: %w", err)
	}
	return map[string]any{"status": "updated", "count": len(relations)}, nil
}

type listMutedTestCasesArgs struct {
	ProjectID int64 `json:"project_id"`
	Page      int   `json:"page"`
	Size      int   `json:"size"`
}

func (r *Registry) listMutedTestCases(ctx context.Context, args listMutedTestCasesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.Size <= 0 {
		args.Size = 20
	}
	cases, err := r.allure.ListMutedTestCases(ctx, args.ProjectID, args.Page, args.Size)
	if err != nil {
		return nil, fmt.Errorf("list muted test cases: %w", err)
	}
	items := make([]map[string]any, len(cases.Content))
	for i, tc := range cases.Content {
		items[i] = map[string]any{"id": tc.ID, "name": tc.Name, "project_id": tc.ProjectID, "status": tc.Status}
	}
	return map[string]any{"test_cases": items, "page": cases.Number, "size": cases.Size, "total": cases.Total, "is_last": cases.Last}, nil
}

type getTestCaseAuditArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
}

func (r *Registry) getTestCaseAudit(ctx context.Context, args getTestCaseAuditArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.Size <= 0 {
		args.Size = 20
	}
	result, err := r.allure.GetTestCaseAudit(ctx, args.TestCaseID, args.Page, args.Size)
	if err != nil {
		return nil, fmt.Errorf("get test case audit: %w", err)
	}
	return map[string]any{"entries": result.Content, "page": result.Number, "size": result.Size, "total": result.Total, "is_last": result.Last}, nil
}

type validateTestCaseQueryArgs struct {
	ProjectID int64  `json:"project_id"`
	Query     string `json:"query"`
}

func (r *Registry) validateTestCaseQuery(ctx context.Context, args validateTestCaseQueryArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.Query == "" {
		return nil, fmt.Errorf("query is required")
	}
	result, err := r.allure.ValidateTestCaseQuery(ctx, args.ProjectID, args.Query)
	if err != nil {
		return nil, fmt.Errorf("validate query: %w", err)
	}
	return result, nil
}

type suggestTestCasesArgs struct {
	ProjectID int64  `json:"project_id"`
	Query     string `json:"query"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

func (r *Registry) suggestTestCases(ctx context.Context, args suggestTestCasesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.Size <= 0 {
		args.Size = 10
	}
	result, err := r.allure.SuggestTestCases(ctx, args.ProjectID, args.Query, args.Page, args.Size)
	if err != nil {
		return nil, fmt.Errorf("suggest test cases: %w", err)
	}
	return result, nil
}

type getTestCaseWorkflowArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseWorkflow(ctx context.Context, args getTestCaseWorkflowArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	result, err := r.allure.GetTestCaseWorkflow(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case workflow: %w", err)
	}
	return result, nil
}

type getTestCaseKeysArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseKeys(ctx context.Context, args getTestCaseKeysArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	keys, err := r.allure.GetTestCaseKeys(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case keys: %w", err)
	}
	items := make([]map[string]any, len(keys))
	for i, k := range keys {
		items[i] = map[string]any{"id": k.ID, "integration_id": k.IntegrationID, "name": k.Name, "url": k.URL}
	}
	return map[string]any{"keys": items}, nil
}

type setTestCaseKeysArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	Keys       []struct {
		ID            int64  `json:"id"`
		IntegrationID int64  `json:"integration_id"`
		Name          string `json:"name"`
		URL           string `json:"url"`
	} `json:"keys"`
}

func (r *Registry) setTestCaseKeys(ctx context.Context, args setTestCaseKeysArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	keys := make([]allure.TestKeyDto, len(args.Keys))
	for i, k := range args.Keys {
		keys[i] = allure.TestKeyDto{ID: k.ID, IntegrationID: k.IntegrationID, Name: k.Name, URL: k.URL}
	}
	if err := r.allure.SetTestCaseKeys(ctx, args.TestCaseID, keys); err != nil {
		return nil, fmt.Errorf("set test case keys: %w", err)
	}
	return map[string]any{"status": "updated", "count": len(keys)}, nil
}

type getTestCaseScenarioFromRunArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseScenarioFromRun(ctx context.Context, args getTestCaseScenarioFromRunArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	result, err := r.allure.GetTestCaseScenarioFromRun(ctx, args.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get scenario from run: %w", err)
	}
	return result, nil
}

type detachTestCaseAutomationArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	StatusID   int64 `json:"status_id"`
	WorkflowID int64 `json:"workflow_id"`
}

func (r *Registry) detachTestCaseAutomation(ctx context.Context, args detachTestCaseAutomationArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if err := r.allure.DetachTestCaseAutomation(ctx, args.TestCaseID, args.StatusID, args.WorkflowID); err != nil {
		return nil, fmt.Errorf("detach automation: %w", err)
	}
	return map[string]any{"status": "detached"}, nil
}

type getTestCaseVersionDataArgs struct {
	VersionID int64 `json:"version_id"`
}

func (r *Registry) getTestCaseVersionData(ctx context.Context, args getTestCaseVersionDataArgs) (any, error) {
	if args.VersionID <= 0 {
		return nil, fmt.Errorf("version_id must be positive")
	}
	result, err := r.allure.GetTestCaseVersionData(ctx, args.VersionID)
	if err != nil {
		return nil, fmt.Errorf("get version data: %w", err)
	}
	return result, nil
}

type deleteTestCaseVersionArgs struct {
	VersionID int64 `json:"version_id"`
}

func (r *Registry) deleteTestCaseVersion(ctx context.Context, args deleteTestCaseVersionArgs) (any, error) {
	if args.VersionID <= 0 {
		return nil, fmt.Errorf("version_id must be positive")
	}
	if err := r.allure.DeleteTestCaseVersion(ctx, args.VersionID); err != nil {
		return nil, fmt.Errorf("delete test case version: %w", err)
	}
	return map[string]any{"status": "deleted"}, nil
}
