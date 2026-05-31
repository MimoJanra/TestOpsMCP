package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/session"
	"github.com/MimoJanra/TestOpsMCP/internal/tasks"
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
		Handler: Typed(r.bulkSetTestCaseStatus),
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
		Handler: Typed(r.bulkAddTestCaseTags),
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
		Handler: Typed(r.bulkRemoveTestCaseTags),
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
		Handler: Typed(r.bulkCloneTestCases),
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
		Handler: Typed(r.bulkAssignTestResults),
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
		Handler: Typed(r.bulkMuteTestResults),
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
		Handler: Typed(r.bulkUnmuteTestResults),
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
		Handler: Typed(r.bulkResolveTestResults),
	})

	// ── Test case bulk: members ───────────────────────────────────────────────

	r.register(&Tool{
		Name:        "bulk_add_test_case_members",
		Description: "Bulk add members to multiple test cases",
		InputSchema: bulkTCSchema("members", "array", "Members to add (each with id and name)", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":   map[string]any{"type": "integer"},
				"name": map[string]any{"type": "string"},
			},
		}),
		Handler: Typed(r.bulkAddTestCaseMembersTool),
	})

	r.register(&Tool{
		Name:        "bulk_remove_test_case_members",
		Description: "Bulk remove members from multiple test cases by member IDs",
		InputSchema: bulkTCSchema("member_ids", "array", "Member IDs to remove", map[string]any{"type": "integer"}),
		Handler:     Typed(r.bulkRemoveTestCaseMembersTool),
	})

	// ── Test case bulk: custom fields ─────────────────────────────────────────

	r.register(&Tool{
		Name:        "bulk_add_test_case_custom_fields",
		Description: "Bulk add custom field values to multiple test cases",
		InputSchema: bulkTCSchema("custom_fields", "array", "Custom field values to add", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"custom_field_id": map[string]any{"type": "integer", "description": "Custom field ID"},
				"value_ids":       map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Value IDs to assign"},
			},
			"required": []string{"custom_field_id", "value_ids"},
		}),
		Handler: Typed(r.bulkAddTestCaseCustomFields),
	})

	r.register(&Tool{
		Name:        "bulk_remove_test_case_custom_fields",
		Description: "Bulk remove custom fields from multiple test cases by custom field IDs",
		InputSchema: bulkTCSchema("custom_field_ids", "array", "Custom field IDs to remove", map[string]any{"type": "integer"}),
		Handler:     Typed(r.bulkRemoveTestCaseCustomFields),
	})

	// ── Test case bulk: external links ────────────────────────────────────────

	r.register(&Tool{
		Name:        "bulk_add_test_case_external_links",
		Description: "Bulk add external links to multiple test cases",
		InputSchema: bulkTCSchema("links", "array", "Links to add (each with name, type, url)", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"type": map[string]any{"type": "string"},
				"url":  map[string]any{"type": "string"},
			},
			"required": []string{"url"},
		}),
		Handler: Typed(r.bulkAddTestCaseExternalLinks),
	})

	// ── Test case bulk: issues ────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "bulk_add_test_case_issues",
		Description: "Bulk add issues (bug tracker links) to multiple test cases",
		InputSchema: bulkTCSchema("issues", "array", "Issues to add", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":             map[string]any{"type": "integer"},
				"display_name":   map[string]any{"type": "string"},
				"url":            map[string]any{"type": "string"},
				"integration_id": map[string]any{"type": "integer"},
			},
		}),
		Handler: Typed(r.bulkAddTestCaseIssues),
	})

	r.register(&Tool{
		Name:        "bulk_remove_test_case_issues",
		Description: "Bulk remove issues from multiple test cases by issue IDs",
		InputSchema: bulkTCSchema("issue_ids", "array", "Issue IDs to remove", map[string]any{"type": "integer"}),
		Handler:     Typed(r.bulkRemoveTestCaseIssues),
	})

	// ── Test case bulk: layer / move / delete ─────────────────────────────────

	r.register(&Tool{
		Name:        "bulk_set_test_case_layer",
		Description: "Bulk set the test layer for multiple test cases",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":    map[string]any{"type": "integer", "description": "Allure project ID"},
				"test_case_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Test case IDs"},
				"layer_id":      map[string]any{"type": "integer", "description": "Test layer ID to set"},
			},
			"required": []string{"project_id", "test_case_ids", "layer_id"},
		},
		Handler: Typed(r.bulkSetTestCaseLayer),
	})

	r.register(&Tool{
		Name:        "bulk_move_test_cases",
		Description: "Bulk move test cases to a different project",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":    map[string]any{"type": "integer", "description": "Source project ID"},
				"test_case_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Test case IDs to move"},
				"to_project_id": map[string]any{"type": "integer", "description": "Destination project ID"},
			},
			"required": []string{"project_id", "test_case_ids", "to_project_id"},
		},
		Handler: Typed(r.bulkMoveTestCases),
	})

	r.register(&Tool{
		Name:        "bulk_delete_test_cases",
		Description: "Bulk permanently delete multiple test cases",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":    map[string]any{"type": "integer", "description": "Allure project ID"},
				"test_case_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Test case IDs to delete"},
			},
			"required": []string{"project_id", "test_case_ids"},
		},
		Handler: Typed(r.bulkDeleteTestCases),
	})

	// ── Test case bulk: run ───────────────────────────────────────────────────

	r.register(&Tool{
		Name:        "bulk_run_test_cases_new_launch",
		Description: "Bulk run multiple test cases in a new launch",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":    map[string]any{"type": "integer", "description": "Allure project ID"},
				"test_case_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Test case IDs to run"},
				"launch_name":   map[string]any{"type": "string", "description": "Name for the new launch (optional)"},
				"assignees":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Usernames to assign (optional)"},
			},
			"required": []string{"project_id", "test_case_ids"},
		},
		Handler: Typed(r.bulkRunTestCasesNewLaunch),
	})

	r.register(&Tool{
		Name:        "bulk_run_test_cases_existing_launch",
		Description: "Bulk run multiple test cases in an existing launch",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":    map[string]any{"type": "integer", "description": "Allure project ID"},
				"test_case_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Test case IDs to run"},
				"launch_id":     map[string]any{"type": "integer", "description": "Existing launch ID"},
				"assignees":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Usernames to assign (optional)"},
			},
			"required": []string{"project_id", "test_case_ids", "launch_id"},
		},
		Handler: Typed(r.bulkRunTestCasesExistingLaunch),
	})

	// ── Test case bulk: test plan / mute ──────────────────────────────────────

	r.register(&Tool{
		Name:        "bulk_create_test_plan",
		Description: "Create a test plan from multiple selected test cases",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":     map[string]any{"type": "integer", "description": "Allure project ID"},
				"test_case_ids":  map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Test case IDs to include"},
				"test_plan_name": map[string]any{"type": "string", "description": "Name for the new test plan"},
			},
			"required": []string{"project_id", "test_case_ids", "test_plan_name"},
		},
		Handler: Typed(r.bulkCreateTestPlan),
	})

	r.register(&Tool{
		Name:        "bulk_mute_test_cases",
		Description: "Bulk mute multiple test cases",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":    map[string]any{"type": "integer", "description": "Allure project ID"},
				"test_case_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Test case IDs to mute"},
			},
			"required": []string{"project_id", "test_case_ids"},
		},
		Handler: Typed(r.bulkMuteTestCases),
	})
}

// bulkTCSchema builds a standard bulk test-case input schema with project_id, test_case_ids
// and one extra field.
func bulkTCSchema(field, fieldType, desc string, items map[string]any) map[string]any {
	fieldDef := map[string]any{"type": fieldType, "description": desc}
	if fieldType == "array" {
		fieldDef["items"] = items
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"project_id":    map[string]any{"type": "integer", "description": "Allure project ID"},
			"test_case_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}, "description": "Test case IDs"},
			field:           fieldDef,
		},
		"required": []string{"project_id", "test_case_ids", field},
	}
}

// ── Bulk handlers ─────────────────────────────────────────────────────────────

type bulkAddTestCaseMembersArgs struct {
	ProjectID   int64              `json:"project_id"`
	TestCaseIDs []int64            `json:"test_case_ids"`
	Members     []allure.MemberDto `json:"members"`
}

func (r *Registry) bulkAddTestCaseMembersTool(ctx context.Context, args bulkAddTestCaseMembersArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if len(args.Members) == 0 {
		return nil, fmt.Errorf("members must not be empty")
	}
	if err := r.allure.BulkAddTestCaseMembers(ctx, args.ProjectID, args.TestCaseIDs, args.Members); err != nil {
		return nil, fmt.Errorf("bulk add members: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkRemoveTestCaseMembersArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	MemberIDs   []int64 `json:"member_ids"`
}

func (r *Registry) bulkRemoveTestCaseMembersTool(ctx context.Context, args bulkRemoveTestCaseMembersArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if len(args.MemberIDs) == 0 {
		return nil, fmt.Errorf("member_ids must not be empty")
	}
	members := make([]allure.MemberDto, len(args.MemberIDs))
	for i, id := range args.MemberIDs {
		members[i] = allure.MemberDto{ID: id}
	}
	if err := r.allure.BulkRemoveTestCaseMembers(ctx, args.ProjectID, args.TestCaseIDs, members); err != nil {
		return nil, fmt.Errorf("bulk remove members: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkAddTestCaseCustomFieldsArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	CustomFields []struct {
		CustomFieldID int64   `json:"custom_field_id"`
		ValueIDs      []int64 `json:"value_ids"`
	} `json:"custom_fields"`
}

func (r *Registry) bulkAddTestCaseCustomFields(ctx context.Context, args bulkAddTestCaseCustomFieldsArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if len(args.CustomFields) == 0 {
		return nil, fmt.Errorf("custom_fields must not be empty")
	}
	cfv := make([]allure.CustomFieldWithValuesDto, len(args.CustomFields))
	for i, cf := range args.CustomFields {
		vals := make([]allure.CustomFieldValueDto, len(cf.ValueIDs))
		for j, vid := range cf.ValueIDs {
			vals[j] = allure.CustomFieldValueDto{ID: vid}
		}
		cfv[i] = allure.CustomFieldWithValuesDto{
			CustomField: allure.CustomFieldDto{ID: cf.CustomFieldID},
			Values:      vals,
		}
	}
	if err := r.allure.BulkAddTestCaseCustomFields(ctx, args.ProjectID, args.TestCaseIDs, cfv); err != nil {
		return nil, fmt.Errorf("bulk add custom fields: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkRemoveTestCaseCustomFieldsArgs struct {
	ProjectID      int64   `json:"project_id"`
	TestCaseIDs    []int64 `json:"test_case_ids"`
	CustomFieldIDs []int64 `json:"custom_field_ids"`
}

func (r *Registry) bulkRemoveTestCaseCustomFields(ctx context.Context, args bulkRemoveTestCaseCustomFieldsArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if err := r.allure.BulkRemoveTestCaseCustomFields(ctx, args.ProjectID, args.TestCaseIDs, args.CustomFieldIDs); err != nil {
		return nil, fmt.Errorf("bulk remove custom fields: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkAddTestCaseExternalLinksArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	Links       []struct {
		Name string `json:"name"`
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"links"`
}

func (r *Registry) bulkAddTestCaseExternalLinks(ctx context.Context, args bulkAddTestCaseExternalLinksArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	links := make([]allure.ExternalLinkDto, len(args.Links))
	for i, l := range args.Links {
		links[i] = allure.ExternalLinkDto{Name: l.Name, Type: l.Type, URL: l.URL}
	}
	if err := r.allure.BulkAddTestCaseExternalLinks(ctx, args.ProjectID, args.TestCaseIDs, links); err != nil {
		return nil, fmt.Errorf("bulk add external links: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkAddTestCaseIssuesArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	Issues      []struct {
		ID            int64  `json:"id"`
		DisplayName   string `json:"display_name"`
		URL           string `json:"url"`
		IntegrationID int64  `json:"integration_id"`
	} `json:"issues"`
}

func (r *Registry) bulkAddTestCaseIssues(ctx context.Context, args bulkAddTestCaseIssuesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	issues := make([]allure.IssueDto, len(args.Issues))
	for i, iss := range args.Issues {
		issues[i] = allure.IssueDto{ID: iss.ID, DisplayName: iss.DisplayName, URL: iss.URL, IntegrationID: iss.IntegrationID}
	}
	if err := r.allure.BulkAddTestCaseIssues(ctx, args.ProjectID, args.TestCaseIDs, issues); err != nil {
		return nil, fmt.Errorf("bulk add issues: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkRemoveTestCaseIssuesArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	IssueIDs    []int64 `json:"issue_ids"`
}

func (r *Registry) bulkRemoveTestCaseIssues(ctx context.Context, args bulkRemoveTestCaseIssuesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if err := r.allure.BulkRemoveTestCaseIssues(ctx, args.ProjectID, args.TestCaseIDs, args.IssueIDs); err != nil {
		return nil, fmt.Errorf("bulk remove issues: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkSetTestCaseLayerArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	LayerID     int64   `json:"layer_id"`
}

func (r *Registry) bulkSetTestCaseLayer(ctx context.Context, args bulkSetTestCaseLayerArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if args.LayerID <= 0 {
		return nil, fmt.Errorf("layer_id must be positive")
	}
	if err := r.allure.BulkSetTestCaseLayer(ctx, args.ProjectID, args.TestCaseIDs, args.LayerID); err != nil {
		return nil, fmt.Errorf("bulk set layer: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkMoveTestCasesArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	ToProjectID int64   `json:"to_project_id"`
}

func (r *Registry) bulkMoveTestCases(ctx context.Context, args bulkMoveTestCasesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if args.ToProjectID <= 0 {
		return nil, fmt.Errorf("to_project_id must be positive")
	}
	if err := r.allure.BulkMoveTestCases(ctx, args.ProjectID, args.TestCaseIDs, args.ToProjectID); err != nil {
		return nil, fmt.Errorf("bulk move: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkDeleteTestCasesArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
}

func (r *Registry) bulkDeleteTestCases(ctx context.Context, args bulkDeleteTestCasesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
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
	result, err := elicit(ctx, fmt.Sprintf("Permanently delete %d test cases from project #%d? This cannot be undone.", len(args.TestCaseIDs), args.ProjectID), schema)
	if err != nil {
		return nil, fmt.Errorf("confirmation failed: %w", err)
	}
	if result.Action != "accept" {
		return map[string]any{"cancelled": true, "message": "Deletion cancelled."}, nil
	}

	if err := r.allure.BulkDeleteTestCases(ctx, args.ProjectID, args.TestCaseIDs); err != nil {
		return nil, fmt.Errorf("bulk delete: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkRunTestCasesNewLaunchArgs struct {
	ProjectID   int64    `json:"project_id"`
	TestCaseIDs []int64  `json:"test_case_ids"`
	LaunchName  string   `json:"launch_name"`
	Assignees   []string `json:"assignees"`
}

func (r *Registry) bulkRunTestCasesNewLaunch(ctx context.Context, args bulkRunTestCasesNewLaunchArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	task, taskCtx := r.taskStore.Create("bulk_run_test_cases_new_launch", ctx)
	r.taskStore.Run(task.ID, taskCtx, func(taskCtx context.Context) {
		if err := r.allure.BulkRunTestCasesNewLaunch(taskCtx, args.ProjectID, args.TestCaseIDs, args.LaunchName, args.Assignees); err != nil {
			r.taskStore.Update(task.ID, tasks.StatusFailed, "", nil, err)
			return
		}
		r.taskStore.Update(task.ID, tasks.StatusSucceeded, "", map[string]any{"status": "started", "count": len(args.TestCaseIDs)}, nil)
	})
	return map[string]any{"task_id": task.ID, "message": "Bulk run started. Use get_task_status to track progress."}, nil
}

type bulkRunTestCasesExistingLaunchArgs struct {
	ProjectID   int64    `json:"project_id"`
	TestCaseIDs []int64  `json:"test_case_ids"`
	LaunchID    int64    `json:"launch_id"`
	Assignees   []string `json:"assignees"`
}

func (r *Registry) bulkRunTestCasesExistingLaunch(ctx context.Context, args bulkRunTestCasesExistingLaunchArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	task, taskCtx := r.taskStore.Create("bulk_run_test_cases_existing_launch", ctx)
	r.taskStore.Run(task.ID, taskCtx, func(taskCtx context.Context) {
		if err := r.allure.BulkRunTestCasesExistingLaunch(taskCtx, args.ProjectID, args.TestCaseIDs, args.LaunchID, args.Assignees); err != nil {
			r.taskStore.Update(task.ID, tasks.StatusFailed, "", nil, err)
			return
		}
		r.taskStore.Update(task.ID, tasks.StatusSucceeded, "", map[string]any{"status": "started", "count": len(args.TestCaseIDs)}, nil)
	})
	return map[string]any{"task_id": task.ID, "message": "Bulk run started. Use get_task_status to track progress."}, nil
}

type bulkCreateTestPlanArgs struct {
	ProjectID    int64   `json:"project_id"`
	TestCaseIDs  []int64 `json:"test_case_ids"`
	TestPlanName string  `json:"test_plan_name"`
}

func (r *Registry) bulkCreateTestPlan(ctx context.Context, args bulkCreateTestPlanArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if args.TestPlanName == "" {
		return nil, fmt.Errorf("test_plan_name is required")
	}
	if err := r.allure.BulkCreateTestPlan(ctx, args.ProjectID, args.TestCaseIDs, args.TestPlanName); err != nil {
		return nil, fmt.Errorf("bulk create test plan: %w", err)
	}
	return map[string]any{"status": "created"}, nil
}

type bulkMuteTestCasesArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
}

func (r *Registry) bulkMuteTestCases(ctx context.Context, args bulkMuteTestCasesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if err := r.allure.BulkMuteTestCases(ctx, args.ProjectID, args.TestCaseIDs); err != nil {
		return nil, fmt.Errorf("bulk mute: %w", err)
	}
	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkSetTestCaseStatusArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
	StatusID    int64   `json:"status_id"`
	WorkflowID  int64   `json:"workflow_id"`
}

func (r *Registry) bulkSetTestCaseStatus(ctx context.Context, args bulkSetTestCaseStatusArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if args.StatusID <= 0 {
		return nil, fmt.Errorf("status_id must be positive")
	}
	if args.WorkflowID <= 0 {
		return nil, fmt.Errorf("workflow_id must be positive")
	}

	r.logger.Info("bulk setting test case status", map[string]any{"project_id": args.ProjectID, "count": len(args.TestCaseIDs)})

	if err := r.allure.BulkSetTestCaseStatus(ctx, args.ProjectID, args.StatusID, args.WorkflowID, args.TestCaseIDs); err != nil {
		r.logger.Error("bulk set test case status", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("bulk set status: %w", err)
	}

	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkAddTestCaseTagsArgs struct {
	ProjectID   int64               `json:"project_id"`
	TestCaseIDs []int64             `json:"test_case_ids"`
	Tags        []allure.TestTagDto `json:"tags"`
}

func (r *Registry) bulkAddTestCaseTags(ctx context.Context, args bulkAddTestCaseTagsArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if len(args.Tags) == 0 {
		return nil, fmt.Errorf("tags must not be empty")
	}

	r.logger.Info("bulk adding test case tags", map[string]any{"project_id": args.ProjectID, "count": len(args.TestCaseIDs)})

	if err := r.allure.BulkAddTestCaseTags(ctx, args.ProjectID, args.TestCaseIDs, args.Tags); err != nil {
		r.logger.Error("bulk add test case tags", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("bulk add tags: %w", err)
	}

	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkRemoveTestCaseTagsArgs struct {
	ProjectID   int64               `json:"project_id"`
	TestCaseIDs []int64             `json:"test_case_ids"`
	Tags        []allure.TestTagDto `json:"tags"`
}

func (r *Registry) bulkRemoveTestCaseTags(ctx context.Context, args bulkRemoveTestCaseTagsArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}
	if len(args.Tags) == 0 {
		return nil, fmt.Errorf("tags must not be empty")
	}

	r.logger.Info("bulk removing test case tags", map[string]any{"project_id": args.ProjectID, "count": len(args.TestCaseIDs)})

	if err := r.allure.BulkRemoveTestCaseTags(ctx, args.ProjectID, args.TestCaseIDs, args.Tags); err != nil {
		r.logger.Error("bulk remove test case tags", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("bulk remove tags: %w", err)
	}

	return map[string]any{"status": "success", "count": len(args.TestCaseIDs)}, nil
}

type bulkCloneTestCasesArgs struct {
	ProjectID   int64   `json:"project_id"`
	TestCaseIDs []int64 `json:"test_case_ids"`
}

func (r *Registry) bulkCloneTestCases(ctx context.Context, args bulkCloneTestCasesArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.TestCaseIDs) == 0 {
		return nil, fmt.Errorf("test_case_ids must not be empty")
	}

	r.logger.Info("bulk cloning test cases async", map[string]any{
		"project_id": args.ProjectID,
		"count":      len(args.TestCaseIDs),
	})

	task, taskCtx := r.taskStore.Create("bulk_clone_test_cases", ctx)
	r.taskStore.Run(task.ID, taskCtx, func(taskCtx context.Context) {
		if err := r.allure.BulkCloneTestCases(taskCtx, args.ProjectID, args.TestCaseIDs); err != nil {
			r.logger.Error("bulk clone test cases", err, map[string]any{"project_id": args.ProjectID})
			r.taskStore.Update(task.ID, tasks.StatusFailed, "", nil, err)
			return
		}
		r.taskStore.Update(task.ID, tasks.StatusSucceeded, "", map[string]any{"status": "cloning_started", "count": len(args.TestCaseIDs)}, nil)
	})

	return map[string]any{"task_id": task.ID, "message": "Bulk clone started. Use get_task_status to track progress."}, nil
}

type bulkAssignTestResultsArgs struct {
	LaunchID      int64    `json:"launch_id"`
	TestResultIDs []int64  `json:"test_result_ids"`
	Assignees     []string `json:"assignees"`
}

func (r *Registry) bulkAssignTestResults(ctx context.Context, args bulkAssignTestResultsArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if len(args.TestResultIDs) == 0 {
		return nil, fmt.Errorf("test_result_ids must not be empty")
	}

	r.logger.Info("bulk assigning test results", map[string]any{"launch_id": args.LaunchID, "count": len(args.TestResultIDs)})

	if err := r.allure.BulkAssignTestResults(ctx, args.LaunchID, args.TestResultIDs, args.Assignees); err != nil {
		r.logger.Error("bulk assign test results", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("bulk assign: %w", err)
	}

	return map[string]any{"status": "success", "count": len(args.TestResultIDs)}, nil
}

type bulkMuteTestResultsArgs struct {
	LaunchID      int64   `json:"launch_id"`
	TestResultIDs []int64 `json:"test_result_ids"`
	Reason        string  `json:"reason"`
}

func (r *Registry) bulkMuteTestResults(ctx context.Context, args bulkMuteTestResultsArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if len(args.TestResultIDs) == 0 {
		return nil, fmt.Errorf("test_result_ids must not be empty")
	}

	r.logger.Info("bulk muting test results", map[string]any{"launch_id": args.LaunchID, "count": len(args.TestResultIDs)})

	if err := r.allure.BulkMuteTestResults(ctx, args.LaunchID, args.TestResultIDs, args.Reason); err != nil {
		r.logger.Error("bulk mute test results", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("bulk mute: %w", err)
	}

	return map[string]any{"status": "success", "count": len(args.TestResultIDs)}, nil
}

type bulkUnmuteTestResultsArgs struct {
	LaunchID      int64   `json:"launch_id"`
	TestResultIDs []int64 `json:"test_result_ids"`
}

func (r *Registry) bulkUnmuteTestResults(ctx context.Context, args bulkUnmuteTestResultsArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if len(args.TestResultIDs) == 0 {
		return nil, fmt.Errorf("test_result_ids must not be empty")
	}

	r.logger.Info("bulk unmuting test results", map[string]any{"launch_id": args.LaunchID, "count": len(args.TestResultIDs)})

	if err := r.allure.BulkUnmuteTestResults(ctx, args.LaunchID, args.TestResultIDs); err != nil {
		r.logger.Error("bulk unmute test results", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("bulk unmute: %w", err)
	}

	return map[string]any{"status": "success", "count": len(args.TestResultIDs)}, nil
}

type bulkResolveTestResultsArgs struct {
	LaunchID      int64   `json:"launch_id"`
	TestResultIDs []int64 `json:"test_result_ids"`
}

func (r *Registry) bulkResolveTestResults(ctx context.Context, args bulkResolveTestResultsArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}
	if len(args.TestResultIDs) == 0 {
		return nil, fmt.Errorf("test_result_ids must not be empty")
	}

	r.logger.Info("bulk resolving test results", map[string]any{"launch_id": args.LaunchID, "count": len(args.TestResultIDs)})

	if err := r.allure.BulkResolveTestResults(ctx, args.LaunchID, args.TestResultIDs); err != nil {
		r.logger.Error("bulk resolve test results", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("bulk resolve: %w", err)
	}

	return map[string]any{"status": "success", "count": len(args.TestResultIDs)}, nil
}
