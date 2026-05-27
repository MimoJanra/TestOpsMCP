package tools

import (
	"context"
	"encoding/json"
	"fmt"
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
		Handler: r.listProjects,
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
		Handler: r.getProject,
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
		Handler: r.getProjectStats,
	})
}

func (r *Registry) listProjects(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		Page int `json:"page"`
		Size int `json:"size"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.Size <= 0 {
		params.Size = 10
	}
	if params.Size > 100 {
		params.Size = 100
	}

	r.logger.Info("listing projects", map[string]any{
		"page": params.Page,
		"size": params.Size,
	})

	projects, err := r.allure.ListProjects(ctx, params.Page, params.Size)
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

func (r *Registry) getProject(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID int64 `json:"project_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching project", map[string]any{"project_id": params.ProjectID})

	project, err := r.allure.GetProject(ctx, params.ProjectID)
	if err != nil {
		r.logger.Error("get project", err, map[string]any{"project_id": params.ProjectID})
		return nil, fmt.Errorf("get project: %w", err)
	}

	return map[string]any{
		"id":          project.ID,
		"name":        project.Name,
		"code":        project.Code,
		"description": project.Description,
	}, nil
}

func (r *Registry) getProjectStats(ctx context.Context, input json.RawMessage) (any, error) {
	var params struct {
		ProjectID int64 `json:"project_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	r.logger.Info("fetching project stats", map[string]any{"project_id": params.ProjectID})

	stats, err := r.allure.GetProjectStats(ctx, params.ProjectID)
	if err != nil {
		r.logger.Error("get project stats", err, map[string]any{"project_id": params.ProjectID})
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
