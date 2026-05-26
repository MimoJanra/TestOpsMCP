package tools

import (
	"context"
	"encoding/json"
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
		Handler: r.getTestCaseTags,
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
					"description": "Complete list of tags to set. Each tag needs either id (existing) or name (new).",
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
		Handler: r.setTestCaseTags,
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
		Handler: r.getTestCaseIssues,
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
							"id":            map[string]any{"type": "integer", "description": "Issue ID (if known)"},
							"display_name":  map[string]any{"type": "string", "description": "Issue display name / key (e.g. PROJ-123)"},
							"url":           map[string]any{"type": "string", "description": "Direct URL to the issue"},
							"integration_id": map[string]any{"type": "integer", "description": "Bug tracker integration ID"},
							"closed":        map[string]any{"type": "boolean"},
						},
					},
				},
			},
			"required": []string{"test_case_id", "issues"},
		},
		Handler: r.setTestCaseIssues,
	})

	// ── Examples (parametrized) ───────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_examples",
		Description: "Get parametrized test examples (data table rows) for a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.getTestCaseExamples,
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
		Handler: r.setTestCaseExamples,
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
		Handler: r.listTestCaseVersions,
	})

	r.register(&Tool{
		Name:        "create_test_case_version",
		Description: "Create a named version snapshot of a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"title":        map[string]any{"type": "string", "description": "Version title (required)"},
				"description":  map[string]any{"type": "string", "description": "Version description (optional)"},
			},
			"required": []string{"test_case_id", "title"},
		},
		Handler: r.createTestCaseVersion,
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
		Handler: r.restoreTestCaseVersion,
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
		Handler: r.getTestCaseAttachments,
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
		Handler: r.deleteTestCaseAttachment,
	})

	// ── Search & filtered lists ───────────────────────────────────────────────

	r.register(&Tool{
		Name: "search_test_cases",
		Description: "Search test cases in a project using AQL/RQL query. " +
			"Example queries: \"name ~ 'login'\", \"status = 'active'\", \"tag = 'smoke'\".",
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
		Handler: r.searchTestCases,
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
		Handler: r.listDeletedTestCases,
	})

	// ── Scenario / step position ──────────────────────────────────────────────

	r.register(&Tool{
		Name:        "delete_test_case_scenario",
		Description: "Delete the entire scenario (all steps) from a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.deleteTestCaseScenario,
	})

	r.register(&Tool{
		Name:        "move_test_case_step",
		Description: "Move a scenario step to a different position within the scenario",
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
		Handler: r.moveTestCaseStep,
	})

	r.register(&Tool{
		Name:        "copy_test_case_step",
		Description: "Copy a scenario step to a different position within the scenario",
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
		Handler: r.copyTestCaseStep,
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
		Handler: r.getTestCaseRelations,
	})

	r.register(&Tool{
		Name:        "set_test_case_relations",
		Description: "Replace all test-case-to-test-case relations (full replace)",
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
		Handler: r.setTestCaseRelations,
	})

	// ── Muted list ────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "list_muted_test_cases",
		Description: "List all muted test cases in a project",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"page":       map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":       map[string]any{"type": "integer", "description": "Items per page", "default": 20},
			},
			"required": []string{"project_id"},
		},
		Handler: r.listMutedTestCases,
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
		Handler: r.getTestCaseAudit,
	})

	// ── Query tools ───────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "validate_test_case_query",
		Description: "Validate an AQL/RQL query for test cases without running it",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{"type": "integer", "description": "Allure project ID"},
				"query":      map[string]any{"type": "string", "description": "AQL/RQL expression to validate"},
			},
			"required": []string{"project_id", "query"},
		},
		Handler: r.validateTestCaseQuery,
	})

	r.register(&Tool{
		Name:        "suggest_test_cases",
		Description: "Get test case name suggestions for a search string (useful for autocomplete)",
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
		Handler: r.suggestTestCases,
	})

	// ── Workflow ──────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_workflow",
		Description: "Get the workflow definition assigned to a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.getTestCaseWorkflow,
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
		Handler: r.getTestCaseKeys,
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
		Handler: r.setTestCaseKeys,
	})

	// ── Scenario from run ─────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "get_test_case_scenario_from_run",
		Description: "Get the scenario (steps) captured from the last test run for a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.getTestCaseScenarioFromRun,
	})

	// ── Automation ────────────────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "detach_test_case_automation",
		Description: "Detach automation from a test case, converting it back to manual",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"status_id":    map[string]any{"type": "integer", "description": "Status ID to set after detach (optional)"},
				"workflow_id":  map[string]any{"type": "integer", "description": "Workflow ID to apply (optional)"},
			},
			"required": []string{"test_case_id"},
		},
		Handler: r.detachTestCaseAutomation,
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
		Handler: r.getTestCaseVersionData,
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
		Handler: r.deleteTestCaseVersion,
	})
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func (r *Registry) getTestCaseTags(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	tags, err := r.allure.GetTestCaseTags(ctx, p.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case tags: %w", err)
	}
	items := make([]map[string]any, len(tags))
	for i, t := range tags {
		items[i] = map[string]any{"id": t.ID, "name": t.Name}
	}
	return map[string]any{"tags": items}, nil
}

func (r *Registry) setTestCaseTags(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64               `json:"test_case_id"`
		Tags       []allure.TestTagDto `json:"tags"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if err := r.allure.SetTestCaseTags(ctx, p.TestCaseID, p.Tags); err != nil {
		return nil, fmt.Errorf("set test case tags: %w", err)
	}
	return map[string]any{"status": "updated", "count": len(p.Tags)}, nil
}

func (r *Registry) getTestCaseIssues(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	issues, err := r.allure.GetTestCaseIssues(ctx, p.TestCaseID)
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

func (r *Registry) setTestCaseIssues(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
		Issues     []struct {
			ID            int64  `json:"id"`
			DisplayName   string `json:"display_name"`
			URL           string `json:"url"`
			IntegrationID int64  `json:"integration_id"`
			Closed        bool   `json:"closed"`
		} `json:"issues"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	issues := make([]allure.IssueDto, len(p.Issues))
	for i, iss := range p.Issues {
		issues[i] = allure.IssueDto{
			ID:            iss.ID,
			DisplayName:   iss.DisplayName,
			URL:           iss.URL,
			IntegrationID: iss.IntegrationID,
			Closed:        iss.Closed,
		}
	}
	if err := r.allure.SetTestCaseIssues(ctx, p.TestCaseID, issues); err != nil {
		return nil, fmt.Errorf("set test case issues: %w", err)
	}
	return map[string]any{"status": "updated", "count": len(issues)}, nil
}

func (r *Registry) getTestCaseExamples(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	examples, err := r.allure.GetTestCaseExamples(ctx, p.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case examples: %w", err)
	}
	return map[string]any{"examples": examples, "count": len(examples)}, nil
}

func (r *Registry) setTestCaseExamples(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64                         `json:"test_case_id"`
		Examples   [][]allure.TestCaseExampleParam `json:"examples"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if err := r.allure.SetTestCaseExamples(ctx, p.TestCaseID, p.Examples); err != nil {
		return nil, fmt.Errorf("set test case examples: %w", err)
	}
	return map[string]any{"status": "updated", "rows": len(p.Examples)}, nil
}

func (r *Registry) listTestCaseVersions(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	versions, err := r.allure.ListTestCaseVersions(ctx, p.TestCaseID)
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

func (r *Registry) createTestCaseVersion(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID  int64  `json:"test_case_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if p.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	version, err := r.allure.CreateTestCaseVersion(ctx, p.TestCaseID, allure.TestCaseVersionCreateRequest{
		Title:       p.Title,
		Description: p.Description,
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

func (r *Registry) restoreTestCaseVersion(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		VersionID int64 `json:"version_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.VersionID <= 0 {
		return nil, fmt.Errorf("version_id must be positive")
	}
	if err := r.allure.RestoreTestCaseVersion(ctx, p.VersionID); err != nil {
		return nil, fmt.Errorf("restore test case version: %w", err)
	}
	return map[string]any{"status": "restored", "version_id": p.VersionID}, nil
}

func (r *Registry) getTestCaseAttachments(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
		Page       int   `json:"page"`
		Size       int   `json:"size"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if p.Size == 0 {
		p.Size = 20
	}
	result, err := r.allure.GetTestCaseAttachments(ctx, p.TestCaseID, p.Page, p.Size)
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

func (r *Registry) deleteTestCaseAttachment(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		AttachmentID int64 `json:"attachment_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.AttachmentID <= 0 {
		return nil, fmt.Errorf("attachment_id must be positive")
	}
	if err := r.allure.DeleteTestCaseAttachment(ctx, p.AttachmentID); err != nil {
		return nil, fmt.Errorf("delete test case attachment: %w", err)
	}
	return map[string]any{"status": "deleted"}, nil
}

func (r *Registry) searchTestCases(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		ProjectID int64  `json:"project_id"`
		Query     string `json:"query"`
		Page      int    `json:"page"`
		Size      int    `json:"size"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if p.Query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if p.Size == 0 {
		p.Size = 20
	}
	if p.Size > 100 {
		p.Size = 100
	}
	cases, err := r.allure.SearchTestCases(ctx, p.ProjectID, p.Query, p.Page, p.Size)
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

func (r *Registry) listDeletedTestCases(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		ProjectID int64 `json:"project_id"`
		Page      int   `json:"page"`
		Size      int   `json:"size"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if p.Size == 0 {
		p.Size = 20
	}
	if p.Size > 100 {
		p.Size = 100
	}
	cases, err := r.allure.ListDeletedTestCases(ctx, p.ProjectID, p.Page, p.Size)
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

func (r *Registry) deleteTestCaseScenario(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if err := r.allure.DeleteTestCaseScenario(ctx, p.TestCaseID); err != nil {
		return nil, fmt.Errorf("delete test case scenario: %w", err)
	}
	return map[string]any{"status": "deleted"}, nil
}

func (r *Registry) moveTestCaseStep(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		StepID   int64 `json:"step_id"`
		AfterID  int64 `json:"after_id"`
		BeforeID int64 `json:"before_id"`
		ParentID int64 `json:"parent_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.StepID <= 0 {
		return nil, fmt.Errorf("step_id must be positive")
	}
	pos := allure.StepPositionDto{AfterID: p.AfterID, BeforeID: p.BeforeID, ParentID: p.ParentID}
	if err := r.allure.MoveTestCaseStep(ctx, p.StepID, pos); err != nil {
		return nil, fmt.Errorf("move test case step: %w", err)
	}
	return map[string]any{"status": "moved"}, nil
}

func (r *Registry) copyTestCaseStep(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		StepID   int64 `json:"step_id"`
		AfterID  int64 `json:"after_id"`
		BeforeID int64 `json:"before_id"`
		ParentID int64 `json:"parent_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.StepID <= 0 {
		return nil, fmt.Errorf("step_id must be positive")
	}
	pos := allure.StepPositionDto{AfterID: p.AfterID, BeforeID: p.BeforeID, ParentID: p.ParentID}
	if err := r.allure.CopyTestCaseStep(ctx, p.StepID, pos); err != nil {
		return nil, fmt.Errorf("copy test case step: %w", err)
	}
	return map[string]any{"status": "copied"}, nil
}

func (r *Registry) getTestCaseRelations(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	relations, err := r.allure.GetTestCaseRelations(ctx, p.TestCaseID)
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

func (r *Registry) setTestCaseRelations(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
		Relations  []struct {
			TargetID   int64  `json:"target_id"`
			TargetName string `json:"target_name"`
		} `json:"relations"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	relations := make([]allure.RelationDto, len(p.Relations))
	for i, rel := range p.Relations {
		relations[i] = allure.RelationDto{
			Target: allure.RelationTargetDto{ID: rel.TargetID, Name: rel.TargetName},
		}
	}
	if err := r.allure.SetTestCaseRelations(ctx, p.TestCaseID, relations); err != nil {
		return nil, fmt.Errorf("set test case relations: %w", err)
	}
	return map[string]any{"status": "updated", "count": len(relations)}, nil
}

func (r *Registry) listMutedTestCases(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		ProjectID int64 `json:"project_id"`
		Page      int   `json:"page"`
		Size      int   `json:"size"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if p.Size == 0 {
		p.Size = 20
	}
	cases, err := r.allure.ListMutedTestCases(ctx, p.ProjectID, p.Page, p.Size)
	if err != nil {
		return nil, fmt.Errorf("list muted test cases: %w", err)
	}
	items := make([]map[string]any, len(cases.Content))
	for i, tc := range cases.Content {
		items[i] = map[string]any{"id": tc.ID, "name": tc.Name, "project_id": tc.ProjectID, "status": tc.Status}
	}
	return map[string]any{"test_cases": items, "page": cases.Number, "size": cases.Size, "total": cases.Total, "is_last": cases.Last}, nil
}

func (r *Registry) getTestCaseAudit(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
		Page       int   `json:"page"`
		Size       int   `json:"size"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if p.Size == 0 {
		p.Size = 20
	}
	result, err := r.allure.GetTestCaseAudit(ctx, p.TestCaseID, p.Page, p.Size)
	if err != nil {
		return nil, fmt.Errorf("get test case audit: %w", err)
	}
	return map[string]any{"entries": result.Content, "page": result.Number, "size": result.Size, "total": result.Total, "is_last": result.Last}, nil
}

func (r *Registry) validateTestCaseQuery(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		ProjectID int64  `json:"project_id"`
		Query     string `json:"query"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if p.Query == "" {
		return nil, fmt.Errorf("query is required")
	}
	result, err := r.allure.ValidateTestCaseQuery(ctx, p.ProjectID, p.Query)
	if err != nil {
		return nil, fmt.Errorf("validate query: %w", err)
	}
	return result, nil
}

func (r *Registry) suggestTestCases(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		ProjectID int64  `json:"project_id"`
		Query     string `json:"query"`
		Page      int    `json:"page"`
		Size      int    `json:"size"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if p.Size == 0 {
		p.Size = 10
	}
	result, err := r.allure.SuggestTestCases(ctx, p.ProjectID, p.Query, p.Page, p.Size)
	if err != nil {
		return nil, fmt.Errorf("suggest test cases: %w", err)
	}
	return result, nil
}

func (r *Registry) getTestCaseWorkflow(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	result, err := r.allure.GetTestCaseWorkflow(ctx, p.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case workflow: %w", err)
	}
	return result, nil
}

func (r *Registry) getTestCaseKeys(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	keys, err := r.allure.GetTestCaseKeys(ctx, p.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get test case keys: %w", err)
	}
	items := make([]map[string]any, len(keys))
	for i, k := range keys {
		items[i] = map[string]any{"id": k.ID, "integration_id": k.IntegrationID, "name": k.Name, "url": k.URL}
	}
	return map[string]any{"keys": items}, nil
}

func (r *Registry) setTestCaseKeys(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
		Keys       []struct {
			ID            int64  `json:"id"`
			IntegrationID int64  `json:"integration_id"`
			Name          string `json:"name"`
			URL           string `json:"url"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	keys := make([]allure.TestKeyDto, len(p.Keys))
	for i, k := range p.Keys {
		keys[i] = allure.TestKeyDto{ID: k.ID, IntegrationID: k.IntegrationID, Name: k.Name, URL: k.URL}
	}
	if err := r.allure.SetTestCaseKeys(ctx, p.TestCaseID, keys); err != nil {
		return nil, fmt.Errorf("set test case keys: %w", err)
	}
	return map[string]any{"status": "updated", "count": len(keys)}, nil
}

func (r *Registry) getTestCaseScenarioFromRun(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	result, err := r.allure.GetTestCaseScenarioFromRun(ctx, p.TestCaseID)
	if err != nil {
		return nil, fmt.Errorf("get scenario from run: %w", err)
	}
	return result, nil
}

func (r *Registry) detachTestCaseAutomation(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		TestCaseID int64 `json:"test_case_id"`
		StatusID   int64 `json:"status_id"`
		WorkflowID int64 `json:"workflow_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if err := r.allure.DetachTestCaseAutomation(ctx, p.TestCaseID, p.StatusID, p.WorkflowID); err != nil {
		return nil, fmt.Errorf("detach automation: %w", err)
	}
	return map[string]any{"status": "detached"}, nil
}

func (r *Registry) getTestCaseVersionData(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		VersionID int64 `json:"version_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.VersionID <= 0 {
		return nil, fmt.Errorf("version_id must be positive")
	}
	result, err := r.allure.GetTestCaseVersionData(ctx, p.VersionID)
	if err != nil {
		return nil, fmt.Errorf("get version data: %w", err)
	}
	return result, nil
}

func (r *Registry) deleteTestCaseVersion(ctx context.Context, input json.RawMessage) (any, error) {
	var p struct {
		VersionID int64 `json:"version_id"`
	}
	if err := json.Unmarshal(input, &p); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}
	if p.VersionID <= 0 {
		return nil, fmt.Errorf("version_id must be positive")
	}
	if err := r.allure.DeleteTestCaseVersion(ctx, p.VersionID); err != nil {
		return nil, fmt.Errorf("delete test case version: %w", err)
	}
	return map[string]any{"status": "deleted"}, nil
}
