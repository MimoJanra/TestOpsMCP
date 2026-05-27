package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
)

func (r *Registry) registerRelationTools() {
	r.register(&Tool{
		Name:        "get_test_case_defects",
		Description: "Get defects linked to a test case",
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
		Handler: r.getTestCaseDefects,
	})

	r.register(&Tool{
		Name:        "add_test_case_defect",
		Description: "Link a defect to a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"defect_id": map[string]any{
					"type":        "integer",
					"description": "Defect ID to link",
				},
			},
			"required": []string{"test_case_id", "defect_id"},
		},
		Handler: r.addTestCaseDefect,
	})

	r.register(&Tool{
		Name:        "remove_test_case_defect",
		Description: "Unlink a defect from a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"defect_id": map[string]any{
					"type":        "integer",
					"description": "Defect ID to unlink",
				},
			},
			"required": []string{"test_case_id", "defect_id"},
		},
		Handler: r.removeTestCaseDefect,
	})

	r.register(&Tool{
		Name:        "get_test_case_members",
		Description: "Get team members assigned to a test case",
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
		Handler: r.getTestCaseMembers,
	})

	r.register(&Tool{
		Name:        "add_test_case_members",
		Description: "Add team members to a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"members": map[string]any{
					"type":        "array",
					"description": "Members to add (with id and name)",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":   map[string]any{"type": "integer"},
							"name": map[string]any{"type": "string"},
						},
					},
				},
			},
			"required": []string{"test_case_id", "members"},
		},
		Handler: r.addTestCaseMembers,
	})

	r.register(&Tool{
		Name:        "remove_test_case_members",
		Description: "Remove team members from a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"member_ids": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "integer",
					},
					"description": "Member IDs to remove",
				},
			},
			"required": []string{"test_case_id", "member_ids"},
		},
		Handler: r.removeTestCaseMembers,
	})

	r.register(&Tool{
		Name:        "get_test_case_external_links",
		Description: "Get external links and relations for a test case",
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
		Handler: r.getTestCaseExternalLinks,
	})

	r.register(&Tool{
		Name:        "add_test_case_external_link",
		Description: "Add an external link/relation to a test case",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Link name/title",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Link type (e.g., GITHUB, JIRA, ISSUE)",
				},
				"url": map[string]any{
					"type":        "string",
					"description": "Link URL",
				},
			},
			"required": []string{"test_case_id", "url"},
		},
		Handler: r.addTestCaseExternalLink,
	})

	r.register(&Tool{
		Name:        "delete_test_case_external_link",
		Description: "Delete an external URL link from a test case by its URL. The API has no per-item delete endpoint — the link is removed by fetching the current list and patching without the matching entry.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"url": map[string]any{
					"type":        "string",
					"description": "URL of the external link to delete",
				},
			},
			"required": []string{"test_case_id", "url"},
		},
		Handler: r.deleteTestCaseExternalLink,
	})
}

func (r *Registry) getTestCaseDefects(ctx context.Context, input json.RawMessage) (any, error) {
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

	r.logger.Info("fetching test case defects", map[string]any{
		"test_case_id": params.TestCaseID,
		"page":         params.Page,
		"size":         params.Size,
	})

	defects, err := r.allure.GetTestCaseDefects(ctx, params.TestCaseID, params.Page, params.Size)
	if err != nil {
		r.logger.Error("get test case defects", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("get test case defects: %w", err)
	}

	return defects, nil
}

func (r *Registry) addTestCaseDefect(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
		DefectID   int64 `json:"defect_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if params.DefectID <= 0 {
		return nil, fmt.Errorf("defect_id must be positive")
	}

	r.logger.Info("adding defect to test case", map[string]any{
		"test_case_id": params.TestCaseID,
		"defect_id":    params.DefectID,
	})

	if err := r.allure.AddTestCaseDefect(ctx, params.TestCaseID, params.DefectID); err != nil {
		r.logger.Error("add test case defect", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("add test case defect: %w", err)
	}

	return map[string]any{"status": "defect_added"}, nil
}

func (r *Registry) removeTestCaseDefect(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
		DefectID   int64 `json:"defect_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if params.DefectID <= 0 {
		return nil, fmt.Errorf("defect_id must be positive")
	}

	r.logger.Info("removing defect from test case", map[string]any{
		"test_case_id": params.TestCaseID,
		"defect_id":    params.DefectID,
	})

	if err := r.allure.RemoveTestCaseDefect(ctx, params.TestCaseID, params.DefectID); err != nil {
		r.logger.Error("remove test case defect", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("remove test case defect: %w", err)
	}

	return map[string]any{"status": "defect_removed"}, nil
}

func (r *Registry) getTestCaseMembers(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("fetching test case members", map[string]any{"test_case_id": params.TestCaseID})

	members, err := r.allure.GetTestCaseMembers(ctx, params.TestCaseID)
	if err != nil {
		r.logger.Error("get test case members", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("get test case members: %w", err)
	}

	items := make([]map[string]any, len(members))
	for i, m := range members {
		items[i] = map[string]any{
			"id":   m.ID,
			"name": m.Name,
		}
	}

	return map[string]any{"members": items}, nil
}

func (r *Registry) addTestCaseMembers(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64              `json:"test_case_id"`
		Members    []allure.MemberDto `json:"members"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if len(params.Members) == 0 {
		return nil, fmt.Errorf("members must not be empty")
	}

	r.logger.Info("adding members to test case", map[string]any{
		"test_case_id": params.TestCaseID,
		"count":        len(params.Members),
	})

	if err := r.allure.AddTestCaseMembers(ctx, params.TestCaseID, params.Members); err != nil {
		r.logger.Error("add test case members", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("add test case members: %w", err)
	}

	return map[string]any{"status": "members_added", "count": len(params.Members)}, nil
}

func (r *Registry) removeTestCaseMembers(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64   `json:"test_case_id"`
		MemberIDs  []int64 `json:"member_ids"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if len(params.MemberIDs) == 0 {
		return nil, fmt.Errorf("member_ids must not be empty")
	}

	r.logger.Info("removing members from test case", map[string]any{
		"test_case_id": params.TestCaseID,
		"count":        len(params.MemberIDs),
	})

	if err := r.allure.RemoveTestCaseMembers(ctx, params.TestCaseID, params.MemberIDs); err != nil {
		r.logger.Error("remove test case members", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("remove test case members: %w", err)
	}

	return map[string]any{"status": "members_removed", "count": len(params.MemberIDs)}, nil
}

func (r *Registry) getTestCaseExternalLinks(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64 `json:"test_case_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("fetching test case external links", map[string]any{"test_case_id": params.TestCaseID})

	links, err := r.allure.GetTestCaseExternalLinks(ctx, params.TestCaseID)
	if err != nil {
		r.logger.Error("get test case external links", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("get test case external links: %w", err)
	}

	items := make([]map[string]any, len(links))
	for i, link := range links {
		items[i] = map[string]any{
			"name": link.Name,
			"type": link.Type,
			"url":  link.URL,
		}
	}

	return map[string]any{"links": items}, nil
}

func (r *Registry) addTestCaseExternalLink(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64  `json:"test_case_id"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		URL        string `json:"url"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if params.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	link := allure.ExternalLinkDto{
		Name: params.Name,
		Type: params.Type,
		URL:  params.URL,
	}

	r.logger.Info("adding external link to test case", map[string]any{
		"test_case_id": params.TestCaseID,
		"url":          params.URL,
	})

	if err := r.allure.AddTestCaseExternalLink(ctx, params.TestCaseID, link); err != nil {
		r.logger.Error("add test case external link", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("add test case external link: %w", err)
	}

	return map[string]any{"status": "link_added"}, nil
}

func (r *Registry) deleteTestCaseExternalLink(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		TestCaseID int64  `json:"test_case_id"`
		URL        string `json:"url"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if params.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	r.logger.Info("deleting external link from test case", map[string]any{
		"test_case_id": params.TestCaseID,
		"url":          params.URL,
	})

	if err := r.allure.DeleteTestCaseExternalLink(ctx, params.TestCaseID, params.URL); err != nil {
		r.logger.Error("delete test case external link", err, map[string]any{"test_case_id": params.TestCaseID})
		return nil, fmt.Errorf("delete test case external link: %w", err)
	}

	return map[string]any{"status": "link_deleted"}, nil
}
