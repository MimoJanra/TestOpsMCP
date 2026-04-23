package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/example/allure-mcp-server/internal/adapters/allure"
	"github.com/example/allure-mcp-server/internal/core"
)

type HandlerFunc func(ctx context.Context, input json.RawMessage) (interface{}, error)

type Tool struct {
	Name        string
	Description string
	InputSchema interface{}
	Handler     HandlerFunc
}

type Registry struct {
	tools      map[string]*Tool
	allure     *allure.Client
	logger     *core.Logger
	mu         sync.RWMutex
}

func NewRegistry(allureClient *allure.Client, logger *core.Logger) *Registry {
	r := &Registry{
		tools:  make(map[string]*Tool),
		allure: allureClient,
		logger: logger,
	}
	r.registerTools()
	return r
}

func (r *Registry) registerTools() {
	r.register(&Tool{
		Name:        "run_allure_launch",
		Description: "Start a test launch in Allure TestOps",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"launch_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the launch",
				},
			},
			"required": []string{"project_id", "launch_name"},
		},
		Handler: r.runAllureLaunch,
	})

	r.register(&Tool{
		Name:        "get_launch_status",
		Description: "Get the status of a test launch",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"launch_id": map[string]interface{}{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: r.getLaunchStatus,
	})

	r.register(&Tool{
		Name:        "get_launch_report",
		Description: "Get the test report for a launch",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"launch_id": map[string]interface{}{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Handler: r.getLaunchReport,
	})
}

func (r *Registry) register(tool *Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
}

func (r *Registry) GetTool(name string) *Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tools[name]
}

func (r *Registry) ListTools() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	tools := make([]*Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (r *Registry) runAllureLaunch(ctx context.Context, input json.RawMessage) (interface{}, error) {
	var params struct {
		ProjectID  int64  `json:"project_id"`
		LaunchName string `json:"launch_name"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}

	if params.LaunchName == "" {
		return nil, fmt.Errorf("launch_name is required")
	}

	r.logger.Info("Starting Allure launch", map[string]interface{}{
		"project_id":  params.ProjectID,
		"launch_name": params.LaunchName,
	})

	launch, err := r.allure.CreateLaunch(ctx, params.ProjectID, params.LaunchName)
	if err != nil {
		r.logger.Error("Failed to create launch", err, map[string]interface{}{
			"project_id": params.ProjectID,
		})
		return nil, fmt.Errorf("create launch: %w", err)
	}

	r.logger.Info("Launch created successfully", map[string]interface{}{
		"launch_id": launch.ID,
	})

	return map[string]interface{}{
		"launch_id": launch.ID,
		"status":    "started",
	}, nil
}

func (r *Registry) getLaunchStatus(ctx context.Context, input json.RawMessage) (interface{}, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("Fetching launch status", map[string]interface{}{
		"launch_id": params.LaunchID,
	})

	status, err := r.allure.GetLaunchStatus(ctx, params.LaunchID)
	if err != nil {
		r.logger.Error("Failed to get launch status", err, map[string]interface{}{
			"launch_id": params.LaunchID,
		})
		return nil, fmt.Errorf("get launch status: %w", err)
	}

	return map[string]interface{}{
		"status": status,
	}, nil
}

func (r *Registry) getLaunchReport(ctx context.Context, input json.RawMessage) (interface{}, error) {
	var params struct {
		LaunchID int64 `json:"launch_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if params.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("Fetching launch report", map[string]interface{}{
		"launch_id": params.LaunchID,
	})

	stats, err := r.allure.GetLaunchStatistics(ctx, params.LaunchID)
	if err != nil {
		r.logger.Error("Failed to get launch statistics", err, map[string]interface{}{
			"launch_id": params.LaunchID,
		})
		return nil, fmt.Errorf("get launch statistics: %w", err)
	}

	return map[string]interface{}{
		"total":   stats.Total,
		"passed":  stats.Passed,
		"failed":  stats.Failed,
		"broken":  stats.Broken,
	}, nil
}
