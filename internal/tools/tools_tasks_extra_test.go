package tools

import (
	"context"
	"testing"

	"github.com/MimoJanra/TestOpsMCP/internal/tasks"
)

func TestGetTaskStatus(t *testing.T) {
	r := newTestRegistry(t)
	task, _ := r.taskStore.Create("some_tool", context.Background())

	result, err := r.getTaskStatus(context.Background(), getTaskStatusArgs{TaskID: task.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := result.(map[string]any)
	if out["id"] != task.ID || out["status"] != string(tasks.StatusWorking) {
		t.Errorf("unexpected task map: %+v", out)
	}

	if _, err := r.getTaskStatus(context.Background(), getTaskStatusArgs{}); err == nil {
		t.Error("expected error for empty task_id")
	}
	if _, err := r.getTaskStatus(context.Background(), getTaskStatusArgs{TaskID: "nope"}); err == nil {
		t.Error("expected error for unknown task_id")
	}
}

func TestListRunningTasks(t *testing.T) {
	r := newTestRegistry(t)
	working, _ := r.taskStore.Create("tool_a", context.Background())
	done, _ := r.taskStore.Create("tool_b", context.Background())
	r.taskStore.Update(done.ID, tasks.StatusSucceeded, "", map[string]any{"ok": true}, nil)

	result, err := r.listRunningTasks(context.Background(), listRunningTasksArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := result.(map[string]any)
	if out["count"] != 1 {
		t.Errorf("count = %v, want 1 (only %q should still be working)", out["count"], working.ID)
	}
	list := out["tasks"].([]map[string]any)
	if len(list) != 1 || list[0]["id"] != working.ID {
		t.Errorf("tasks = %+v, want just %q", list, working.ID)
	}
}

func TestCancelTask(t *testing.T) {
	r := newTestRegistry(t)
	task, _ := r.taskStore.Create("some_tool", context.Background())

	result, err := r.cancelTask(context.Background(), cancelTaskArgs{TaskID: task.ID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := result.(map[string]any)
	if out["status"] != "cancelled" {
		t.Errorf("status = %v, want cancelled", out["status"])
	}

	got, _ := r.taskStore.Get(task.ID)
	if got.Status != tasks.StatusCancelled {
		t.Errorf("task store status = %v, want cancelled", got.Status)
	}

	if _, err := r.cancelTask(context.Background(), cancelTaskArgs{}); err == nil {
		t.Error("expected error for empty task_id")
	}
	if _, err := r.cancelTask(context.Background(), cancelTaskArgs{TaskID: "nope"}); err == nil {
		t.Error("expected error for unknown task_id")
	}
	// Already cancelled: Cancel() should now report false (not "working" anymore).
	if _, err := r.cancelTask(context.Background(), cancelTaskArgs{TaskID: task.ID}); err == nil {
		t.Error("expected error when cancelling an already-cancelled task")
	}
}

func TestTaskToMap(t *testing.T) {
	task := &tasks.Task{
		ID:      "abc",
		Tool:    "some_tool",
		Status:  tasks.StatusFailed,
		Message: "partial progress",
		Error:   "boom",
		Result:  map[string]any{"n": 1},
	}
	m := taskToMap(task)
	for _, key := range []string{"id", "tool", "status", "message", "error", "result", "created_at", "updated_at"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected key %q in task map, got %+v", key, m)
		}
	}
}
