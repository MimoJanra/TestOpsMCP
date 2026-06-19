package tools

import (
	"context"
	"fmt"
	"strings"
)

func (r *Registry) registerProjectTools() {
	r.register(&Tool{
		Name:        "list_projects",
		Description: "List all accessible projects",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
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
		},
		Handler: Typed(r.listProjects),
	})

	r.register(&Tool{
		Name: "find_project",
		Description: "Find a project by name or code (case-insensitive substring match). " +
			"Use this to resolve a human-readable project name or code (e.g. \"TSi\") to its numeric " +
			"Allure project ID — instead of paging through list_projects or guessing IDs.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Name or code to search for (case-insensitive substring match)",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Max matches to return (1-100, default 20)",
					"default":     20,
				},
			},
			"required": []string{"query"},
		},
		Handler: Typed(r.findProject),
	})

	r.register(&Tool{
		Name:        "get_project",
		Description: "Get project details and settings",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.getProject),
	})

	r.register(&Tool{
		Name:        "get_project_stats",
		Description: "Get project statistics (test count, runs)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.getProjectStats),
	})
}

type listProjectsArgs struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

func (r *Registry) listProjects(ctx context.Context, args listProjectsArgs) (any, error) {
	if args.Size <= 0 {
		args.Size = 10
	}
	if args.Size > 100 {
		args.Size = 100
	}

	r.logger.Info("listing projects", map[string]any{
		"page": args.Page,
		"size": args.Size,
	})

	projects, err := r.allure.ListProjects(ctx, args.Page, args.Size)
	if err != nil {
		r.logger.Error("list projects", err, map[string]any{})
		return nil, fmt.Errorf("list projects: %w", err)
	}

	items := make([]map[string]any, len(projects.Content))
	for i, p := range projects.Content {
		items[i] = map[string]any{
			"id":   p.ID,
			"name": p.Name,
			"code": p.Code,
		}
	}

	return map[string]any{
		"projects": items,
		"page":     projects.Number,
		"size":     projects.Size,
		"total":    projects.Total,
		"is_last":  projects.Last,
	}, nil
}

// find_project scans pages of /api/project client-side because the Allure
// TestOps API exposes no reliable server-side name/code filter for projects.
const (
	findProjectPageSize = 100
	findProjectMaxPages = 50 // scan at most 5000 projects before giving up
)

type findProjectArgs struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func (r *Registry) findProject(ctx context.Context, args findProjectArgs) (any, error) {
	query := strings.TrimSpace(args.Query)
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	limit := args.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	r.logger.Info("finding project", map[string]any{"query": query, "limit": limit})

	needle := strings.ToLower(query)
	matches := make([]map[string]any, 0)
	scanned := 0
	truncated := false

	for page := 0; page < findProjectMaxPages; page++ {
		resp, err := r.allure.ListProjects(ctx, page, findProjectPageSize)
		if err != nil {
			r.logger.Error("find project", err, map[string]any{"query": query})
			return nil, fmt.Errorf("find project: %w", err)
		}

		for _, p := range resp.Content {
			scanned++
			if strings.Contains(strings.ToLower(p.Name), needle) ||
				strings.Contains(strings.ToLower(p.Code), needle) {
				matches = append(matches, map[string]any{
					"id":   p.ID,
					"name": p.Name,
					"code": p.Code,
				})
			}
		}

		if len(matches) >= limit {
			// More matches may exist beyond the requested limit.
			truncated = len(matches) > limit
			matches = matches[:limit]
			break
		}
		if resp.Last || len(resp.Content) == 0 {
			break
		}
		if page == findProjectMaxPages-1 {
			// Hit the page cap before exhausting the project list.
			truncated = true
		}
	}

	return map[string]any{
		"query":     query,
		"matches":   matches,
		"count":     len(matches),
		"scanned":   scanned,
		"truncated": truncated,
	}, nil
}

type getProjectArgs struct {
	ProjectID int64 `json:"project_id"`
}

func (r *Registry) getProject(ctx context.Context, args getProjectArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching project", map[string]any{"project_id": args.ProjectID})

	project, err := r.allure.GetProject(ctx, args.ProjectID)
	if err != nil {
		r.logger.Error("get project", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("get project: %w", err)
	}

	return map[string]any{
		"id":          project.ID,
		"name":        project.Name,
		"code":        project.Code,
		"description": project.Description,
	}, nil
}

type getProjectStatsArgs struct {
	ProjectID int64 `json:"project_id"`
}

func (r *Registry) getProjectStats(ctx context.Context, args getProjectStatsArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching project stats", map[string]any{"project_id": args.ProjectID})

	stats, err := r.allure.GetProjectStats(ctx, args.ProjectID)
	if err != nil {
		r.logger.Error("get project stats", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("get project stats: %w", err)
	}

	return map[string]any{
		"project_id":           stats.ID,
		"automated_test_cases": stats.AutomatedTestCases,
		"manual_test_cases":    stats.ManualTestCases,
		"automation_percent":   stats.AutomationPercent,
		"launches":             stats.Launches,
	}, nil
}
