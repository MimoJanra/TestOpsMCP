package tools

import (
	"context"
	"fmt"

	"github.com/MimoJanra/TestOpsMCP/internal/tasks"
)

func (r *Registry) registerTaskTools() {
	r.register(&Tool{
		Name:        "get_task_status",
		Description: "Get the status of an async background task by its ID",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task_id": map[string]any{
					"type":        "string",
					"description": "Task ID returned by the long-running tool",
				},
			},
			"required": []string{"task_id"},
		},
		Annotations: map[string]any{
			"readOnlyHint":    true,
			"destructiveHint": false,
		},
		Handler: Typed(r.getTaskStatus),
	})

	r.register(&Tool{
		Name:        "list_running_tasks",
		Description: "List all active (working) background tasks",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Annotations: map[string]any{
			"readOnlyHint":    true,
			"destructiveHint": false,
		},
		Handler: Typed(r.listRunningTasks),
	})

	r.register(&Tool{
		Name:        "cancel_task",
		Description: "Cancel a running background task",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task_id": map[string]any{
					"type":        "string",
					"description": "Task ID to cancel",
				},
			},
			"required": []string{"task_id"},
		},
		Annotations: map[string]any{
			"readOnlyHint":    false,
			"destructiveHint": false,
		},
		Handler: Typed(r.cancelTask),
	})
}

type getTaskStatusArgs struct {
	TaskID string `json:"task_id"`
}

func (r *Registry) getTaskStatus(ctx context.Context, args getTaskStatusArgs) (any, error) {
	if args.TaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	t, ok := r.taskStore.Get(args.TaskID)
	if !ok {
		return nil, fmt.Errorf("task not found: %s", args.TaskID)
	}
	return taskToMap(t), nil
}

type listRunningTasksArgs struct{}

func (r *Registry) listRunningTasks(_ context.Context, _ listRunningTasksArgs) (any, error) {
	all := r.taskStore.List()
	running := make([]map[string]any, 0)
	for _, t := range all {
		if t.Status == tasks.StatusWorking {
			running = append(running, taskToMap(t))
		}
	}
	return map[string]any{"tasks": running, "count": len(running)}, nil
}

type cancelTaskArgs struct {
	TaskID string `json:"task_id"`
}

func (r *Registry) cancelTask(_ context.Context, args cancelTaskArgs) (any, error) {
	if args.TaskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}
	if !r.taskStore.Cancel(args.TaskID) {
		return nil, fmt.Errorf("task not found or not cancellable: %s", args.TaskID)
	}
	return map[string]any{"status": "cancelled", "task_id": args.TaskID}, nil
}

func taskToMap(t *tasks.Task) map[string]any {
	m := map[string]any{
		"id":         t.ID,
		"tool":       t.Tool,
		"status":     string(t.Status),
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	}
	if t.Message != "" {
		m["message"] = t.Message
	}
	if t.Error != "" {
		m["error"] = t.Error
	}
	if t.Result != nil {
		m["result"] = t.Result
	}
	return m
}
