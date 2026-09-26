package tools

import (
	"context"
	"fmt"
	"strings"
)

// registerDefectTools registers project defect management. Linking tools
// (add_test_case_defect, get_launch_defects...) need defect ids, which only
// these tools can find or create.
func (r *Registry) registerDefectTools() {
	r.register(&Tool{
		Name:        "list_defects",
		Description: "List a project's defects (newest first), optionally filtered by name or status. Returns defect ids for add_test_case_defect.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":  map[string]any{"type": "integer", "description": "Allure project ID"},
				"name_filter": map[string]any{"type": "string", "description": "Name substring (optional)"},
				"status":      map[string]any{"type": "string", "enum": []string{"open", "closed"}, "description": "Only open or only closed defects (optional)"},
				"page":        map[string]any{"type": "integer", "description": "Page number (0-based)", "default": 0},
				"size":        map[string]any{"type": "integer", "description": "Items per page (default 20, max 100)", "default": 20},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.listDefects),
	})

	r.register(&Tool{
		Name:        "get_defect",
		Description: "Get a defect: name, description, open/closed state and linked issue.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"defect_id": map[string]any{"type": "integer", "description": "Defect ID"},
			},
			"required": []string{"defect_id"},
		},
		Handler: Typed(r.getDefect),
	})

	r.register(&Tool{
		Name:        "create_defect",
		Description: "Create a defect in a project. Link it to test cases with add_test_case_defect.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id":  map[string]any{"type": "integer", "description": "Allure project ID"},
				"name":        map[string]any{"type": "string", "description": "Defect name"},
				"description": map[string]any{"type": "string", "description": "Description (optional)"},
			},
			"required": []string{"project_id", "name"},
		},
		Handler: Typed(r.createDefect),
	})

	r.register(&Tool{
		Name:        "update_defect",
		Description: "Rename a defect, change its description, or close/reopen it (closed=true/false). Only the fields you pass change.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"defect_id":   map[string]any{"type": "integer", "description": "Defect ID"},
				"name":        map[string]any{"type": "string", "description": "New name (optional)"},
				"description": map[string]any{"type": "string", "description": "New description (optional)"},
				"closed":      map[string]any{"type": "boolean", "description": "true closes the defect, false reopens it (optional)"},
			},
			"required": []string{"defect_id"},
		},
		Handler: Typed(r.updateDefect),
	})

	r.register(&Tool{
		Name:        "delete_defect",
		Description: "Permanently delete a defect; test cases and results only lose the link. Prefer update_defect with closed=true. Asks the user to confirm.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"defect_id": map[string]any{"type": "integer", "description": "Defect ID"},
			},
			"required": []string{"defect_id"},
		},
		Handler: Typed(r.deleteDefect),
	})
}

type listDefectsArgs struct {
	ProjectID  int64  `json:"project_id"`
	NameFilter string `json:"name_filter"`
	Status     string `json:"status"`
	Page       int    `json:"page"`
	Size       int    `json:"size"`
}

func (r *Registry) listDefects(ctx context.Context, args listDefectsArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	status := strings.ToLower(args.Status)
	if status != "" && status != "open" && status != "closed" {
		return nil, fmt.Errorf("status must be \"open\" or \"closed\"")
	}
	if args.Size <= 0 {
		args.Size = 20
	}
	if args.Size > 100 {
		args.Size = 100
	}
	page, err := r.allure.ListDefects(ctx, args.ProjectID, args.NameFilter, status, args.Page, args.Size)
	if err != nil {
		return nil, fmt.Errorf("list defects: %w", err)
	}
	items := make([]map[string]any, len(page.Content))
	for i, d := range page.Content {
		items[i] = map[string]any{"id": d.ID, "name": d.Name, "closed": d.Closed, "created_date": d.CreatedDate}
	}
	return map[string]any{"defects": items, "page": page.Number, "total": page.Total, "is_last": page.Last}, nil
}

type defectIDArgs struct {
	DefectID int64 `json:"defect_id"`
}

func (r *Registry) getDefect(ctx context.Context, args defectIDArgs) (any, error) {
	if args.DefectID <= 0 {
		return nil, fmt.Errorf("defect_id must be positive")
	}
	d, err := r.allure.GetDefect(ctx, args.DefectID)
	if err != nil {
		return nil, fmt.Errorf("get defect: %w", err)
	}
	out := map[string]any{
		"id":                 d.ID,
		"name":               d.Name,
		"description":        d.Description,
		"project_id":         d.ProjectID,
		"closed":             d.Closed,
		"created_date":       d.CreatedDate,
		"last_modified_date": d.LastModifiedDate,
	}
	if d.Issue != nil {
		out["issue"] = map[string]any{"id": d.Issue.ID, "key": issueKey(*d.Issue), "url": d.Issue.URL}
	}
	return out, nil
}

type createDefectArgs struct {
	ProjectID   int64  `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *Registry) createDefect(ctx context.Context, args createDefectArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if strings.TrimSpace(args.Name) == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	d, err := r.allure.CreateDefect(ctx, args.ProjectID, strings.TrimSpace(args.Name), args.Description)
	if err != nil {
		return nil, fmt.Errorf("create defect: %w", err)
	}
	return map[string]any{"id": d.ID, "name": d.Name, "project_id": d.ProjectID}, nil
}

type updateDefectArgs struct {
	DefectID    int64   `json:"defect_id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Closed      *bool   `json:"closed"`
}

func (r *Registry) updateDefect(ctx context.Context, args updateDefectArgs) (any, error) {
	if args.DefectID <= 0 {
		return nil, fmt.Errorf("defect_id must be positive")
	}
	if args.Name == nil && args.Description == nil && args.Closed == nil {
		return nil, fmt.Errorf("pass at least one of name, description, closed")
	}
	if args.Name != nil && strings.TrimSpace(*args.Name) == "" {
		return nil, fmt.Errorf("name must not be empty")
	}
	if err := r.allure.UpdateDefect(ctx, args.DefectID, args.Name, args.Description, args.Closed); err != nil {
		return nil, fmt.Errorf("update defect: %w", err)
	}
	return map[string]any{"status": "updated"}, nil
}

func (r *Registry) deleteDefect(ctx context.Context, args defectIDArgs) (any, error) {
	if args.DefectID <= 0 {
		return nil, fmt.Errorf("defect_id must be positive")
	}
	d, err := r.allure.GetDefect(ctx, args.DefectID)
	if err != nil {
		return nil, fmt.Errorf("look up defect: %w", err)
	}
	if ok, resp, err := confirmDestructive(ctx, fmt.Sprintf("Permanently delete defect #%d %q?", d.ID, d.Name)); !ok {
		return resp, err
	}
	if err := r.allure.DeleteDefect(ctx, args.DefectID); err != nil {
		return nil, fmt.Errorf("delete defect: %w", err)
	}
	return map[string]any{"status": "deleted"}, nil
}
