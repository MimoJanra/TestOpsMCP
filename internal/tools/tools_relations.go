package tools

import (
	"context"
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
		Handler: Typed(r.getTestCaseDefects),
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
		Handler: Typed(r.addTestCaseDefect),
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
		Handler: Typed(r.removeTestCaseDefect),
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
		Handler: Typed(r.getTestCaseMembers),
	})

	r.register(&Tool{
		Name: "add_test_case_members",
		Description: "Add members to a test case, keeping the existing ones. The member id is a project member id — find one via " +
			"execute_testops_operation on GET /api/member/suggest with projectId (an org-wide user id that isn't a project member fails with " +
			"\"Some role users not found\"). role is optional: the project's role scheme decides the stored role (confirmed live).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{
					"type":        "integer",
					"description": "Allure test case ID",
				},
				"members": map[string]any{
					"type":        "array",
					"description": "Members to add",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"id":   map[string]any{"type": "integer", "description": "Project member ID"},
							"name": map[string]any{"type": "string"},
							"role": map[string]any{
								"type":        "object",
								"description": "Optional role, e.g. {\"id\": -1, \"name\": \"Owner\"}",
								"properties": map[string]any{
									"id":   map[string]any{"type": "integer"},
									"name": map[string]any{"type": "string"},
								},
							},
						},
						"required": []string{"id"},
					},
				},
			},
			"required": []string{"test_case_id", "members"},
		},
		Handler: Typed(r.addTestCaseMembers),
	})

	r.register(&Tool{
		Name:        "remove_test_case_members",
		Description: "Remove members from a test case (ids as returned by get_test_case_members).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID (optional — looked up from the test case; if passed it must match)",
				},
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
		Handler: Typed(r.removeTestCaseMembers),
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
		Handler: Typed(r.getTestCaseExternalLinks),
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
		Handler: Typed(r.addTestCaseExternalLink),
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
		Handler: Typed(r.deleteTestCaseExternalLink),
	})
}

type getTestCaseDefectsArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
}

func (r *Registry) getTestCaseDefects(ctx context.Context, args getTestCaseDefectsArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	if args.Size <= 0 {
		args.Size = 10
	}
	if args.Size > 100 {
		args.Size = 100
	}

	r.logger.Info("fetching test case defects", map[string]any{
		"test_case_id": args.TestCaseID,
		"page":         args.Page,
		"size":         args.Size,
	})

	defects, err := r.allure.GetTestCaseDefects(ctx, args.TestCaseID, args.Page, args.Size)
	if err != nil {
		r.logger.Error("get test case defects", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("get test case defects: %w", err)
	}

	return defects, nil
}

type addTestCaseDefectArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	DefectID   int64 `json:"defect_id"`
}

func (r *Registry) addTestCaseDefect(ctx context.Context, args addTestCaseDefectArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.DefectID <= 0 {
		return nil, fmt.Errorf("defect_id must be positive")
	}

	r.logger.Info("adding defect to test case", map[string]any{
		"test_case_id": args.TestCaseID,
		"defect_id":    args.DefectID,
	})

	if err := r.allure.AddTestCaseDefect(ctx, args.TestCaseID, args.DefectID); err != nil {
		r.logger.Error("add test case defect", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("add test case defect: %w", err)
	}

	return map[string]any{"status": "defect_added"}, nil
}

type removeTestCaseDefectArgs struct {
	TestCaseID int64 `json:"test_case_id"`
	DefectID   int64 `json:"defect_id"`
}

func (r *Registry) removeTestCaseDefect(ctx context.Context, args removeTestCaseDefectArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.DefectID <= 0 {
		return nil, fmt.Errorf("defect_id must be positive")
	}

	r.logger.Info("removing defect from test case", map[string]any{
		"test_case_id": args.TestCaseID,
		"defect_id":    args.DefectID,
	})

	if err := r.allure.RemoveTestCaseDefect(ctx, args.TestCaseID, args.DefectID); err != nil {
		r.logger.Error("remove test case defect", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("remove test case defect: %w", err)
	}

	return map[string]any{"status": "defect_removed"}, nil
}

type getTestCaseMembersArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseMembers(ctx context.Context, args getTestCaseMembersArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("fetching test case members", map[string]any{"test_case_id": args.TestCaseID})

	members, err := r.allure.GetTestCaseMembers(ctx, args.TestCaseID)
	if err != nil {
		r.logger.Error("get test case members", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("get test case members: %w", err)
	}

	items := make([]map[string]any, len(members))
	for i, m := range members {
		item := map[string]any{
			"id":   m.ID,
			"name": m.Name,
		}
		if m.Role != nil {
			item["role"] = map[string]any{"id": m.Role.ID, "name": m.Role.Name}
		}
		items[i] = item
	}

	return map[string]any{"members": items}, nil
}

type addTestCaseMembersArgs struct {
	TestCaseID int64              `json:"test_case_id"`
	Members    []allure.MemberDto `json:"members"`
}

func (r *Registry) addTestCaseMembers(ctx context.Context, args addTestCaseMembersArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if len(args.Members) == 0 {
		return nil, fmt.Errorf("members must not be empty")
	}
	for i, m := range args.Members {
		if m.ID <= 0 {
			return nil, fmt.Errorf("member %d: id must be positive", i)
		}
	}
	projectID, err := r.testCaseProjectID(ctx, args.TestCaseID)
	if err != nil {
		return nil, err
	}

	r.logger.Info("adding members to test case", map[string]any{
		"test_case_id": args.TestCaseID,
		"count":        len(args.Members),
	})

	// v2 bulk add appends; the v1 POST /api/testcase/{id}/members this used
	// to call replaced every existing member (confirmed live).
	if err := r.allure.BulkAddTestCaseMembers(ctx, projectID, []int64{args.TestCaseID}, args.Members); err != nil {
		r.logger.Error("add test case members", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("add test case members: %w", err)
	}

	return map[string]any{"status": "members_added", "count": len(args.Members)}, nil
}

type removeTestCaseMembersArgs struct {
	ProjectID  int64   `json:"project_id"`
	TestCaseID int64   `json:"test_case_id"`
	MemberIDs  []int64 `json:"member_ids"`
}

func (r *Registry) removeTestCaseMembers(ctx context.Context, args removeTestCaseMembersArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if len(args.MemberIDs) == 0 {
		return nil, fmt.Errorf("member_ids must not be empty")
	}
	// A selection with the wrong project matches nothing, and the API still
	// answers 204 — so resolve the real project instead of trusting the caller.
	projectID, err := r.testCaseProjectID(ctx, args.TestCaseID)
	if err != nil {
		return nil, err
	}
	if args.ProjectID != 0 && args.ProjectID != projectID {
		return nil, fmt.Errorf("test case %d belongs to project %d, not %d", args.TestCaseID, projectID, args.ProjectID)
	}

	r.logger.Info("removing members from test case", map[string]any{
		"test_case_id": args.TestCaseID,
		"count":        len(args.MemberIDs),
	})

	if err := r.allure.RemoveTestCaseMembers(ctx, projectID, args.TestCaseID, args.MemberIDs); err != nil {
		r.logger.Error("remove test case members", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("remove test case members: %w", err)
	}

	return map[string]any{"status": "members_removed", "count": len(args.MemberIDs)}, nil
}

type getTestCaseExternalLinksArgs struct {
	TestCaseID int64 `json:"test_case_id"`
}

func (r *Registry) getTestCaseExternalLinks(ctx context.Context, args getTestCaseExternalLinksArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}

	r.logger.Info("fetching test case external links", map[string]any{"test_case_id": args.TestCaseID})

	links, err := r.allure.GetTestCaseExternalLinks(ctx, args.TestCaseID)
	if err != nil {
		r.logger.Error("get test case external links", err, map[string]any{"test_case_id": args.TestCaseID})
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

type addTestCaseExternalLinkArgs struct {
	TestCaseID int64  `json:"test_case_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	URL        string `json:"url"`
}

func (r *Registry) addTestCaseExternalLink(ctx context.Context, args addTestCaseExternalLinkArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	link := allure.ExternalLinkDto{
		Name: args.Name,
		Type: args.Type,
		URL:  args.URL,
	}

	r.logger.Info("adding external link to test case", map[string]any{
		"test_case_id": args.TestCaseID,
		"url":          args.URL,
	})

	if err := r.allure.AddTestCaseExternalLink(ctx, args.TestCaseID, link); err != nil {
		r.logger.Error("add test case external link", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("add test case external link: %w", err)
	}

	return map[string]any{"status": "link_added"}, nil
}

type deleteTestCaseExternalLinkArgs struct {
	TestCaseID int64  `json:"test_case_id"`
	URL        string `json:"url"`
}

func (r *Registry) deleteTestCaseExternalLink(ctx context.Context, args deleteTestCaseExternalLinkArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	r.logger.Info("deleting external link from test case", map[string]any{
		"test_case_id": args.TestCaseID,
		"url":          args.URL,
	})

	if err := r.allure.DeleteTestCaseExternalLink(ctx, args.TestCaseID, args.URL); err != nil {
		r.logger.Error("delete test case external link", err, map[string]any{"test_case_id": args.TestCaseID})
		return nil, fmt.Errorf("delete test case external link: %w", err)
	}

	return map[string]any{"status": "link_deleted"}, nil
}

// testCaseProjectID looks up the project a test case belongs to. v2 bulk
// endpoints need it in the selection and silently match nothing without the
// right one.
func (r *Registry) testCaseProjectID(ctx context.Context, testCaseID int64) (int64, error) {
	overview, err := r.allure.GetTestCaseOverview(ctx, testCaseID)
	if err != nil {
		return 0, fmt.Errorf("look up test case %d: %w", testCaseID, err)
	}
	projectID := int64(nodeFloat(overview, "projectId"))
	if projectID <= 0 {
		return 0, fmt.Errorf("test case %d has no projectId in its overview", testCaseID)
	}
	return projectID, nil
}
