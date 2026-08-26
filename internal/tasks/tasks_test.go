package tasks

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil || s.tasks == nil {
		t.Fatal("NewStore returned a store with a nil task map")
	}
}

func TestCreate(t *testing.T) {
	s := NewStore()

	type ctxKey struct{}
	parent := context.WithValue(context.Background(), ctxKey{}, "session-123")
	parent, cancelParent := context.WithCancel(parent)

	task, taskCtx := s.Create("my_tool", parent)

	if task.ID == "" {
		t.Error("expected non-empty task ID")
	}
	if len(task.ID) != 32 { // 16 bytes hex-encoded
		t.Errorf("task ID length = %d, want 32", len(task.ID))
	}
	if task.Tool != "my_tool" {
		t.Errorf("Tool = %q, want %q", task.Tool, "my_tool")
	}
	if task.Status != StatusWorking {
		t.Errorf("Status = %q, want %q", task.Status, StatusWorking)
	}
	if task.CreatedAt.IsZero() || task.UpdatedAt.IsZero() {
		t.Error("expected CreatedAt/UpdatedAt to be set")
	}

	if got, ok := taskCtx.Value(ctxKey{}).(string); !ok || got != "session-123" {
		t.Errorf("taskCtx did not propagate parent value, got %v", taskCtx.Value(ctxKey{}))
	}

	// Cancelling the parent must NOT cancel the task context (context.WithoutCancel).
	cancelParent()
	select {
	case <-taskCtx.Done():
		t.Error("taskCtx was cancelled when parent was cancelled; Create should decouple cancellation")
	default:
	}

	stored, ok := s.Get(task.ID)
	if !ok {
		t.Fatal("task not found in store after Create")
	}
	if stored.ID != task.ID {
		t.Errorf("stored task ID = %q, want %q", stored.ID, task.ID)
	}
}

func TestCreate_UniqueIDs(t *testing.T) {
	s := NewStore()
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		task, _ := s.Create("t", context.Background())
		if seen[task.ID] {
			t.Fatalf("duplicate task ID generated: %s", task.ID)
		}
		seen[task.ID] = true
	}
}

func TestUpdate(t *testing.T) {
	s := NewStore()
	task, _ := s.Create("t", context.Background())

	before := task.UpdatedAt
	time.Sleep(time.Millisecond)
	s.Update(task.ID, StatusSucceeded, "done", map[string]any{"n": 1}, nil)

	got, ok := s.Get(task.ID)
	if !ok {
		t.Fatal("task not found")
	}
	if got.Status != StatusSucceeded {
		t.Errorf("Status = %q, want %q", got.Status, StatusSucceeded)
	}
	if got.Message != "done" {
		t.Errorf("Message = %q, want %q", got.Message, "done")
	}
	if got.Result == nil {
		t.Error("expected Result to be set")
	}
	if got.Error != "" {
		t.Errorf("Error = %q, want empty", got.Error)
	}
	if !got.UpdatedAt.After(before) {
		t.Error("expected UpdatedAt to advance")
	}
}

func TestUpdate_WithError(t *testing.T) {
	s := NewStore()
	task, _ := s.Create("t", context.Background())

	s.Update(task.ID, StatusFailed, "", nil, errors.New("boom"))

	got, _ := s.Get(task.ID)
	if got.Status != StatusFailed {
		t.Errorf("Status = %q, want %q", got.Status, StatusFailed)
	}
	if got.Error != "boom" {
		t.Errorf("Error = %q, want %q", got.Error, "boom")
	}
}

func TestUpdate_UnknownID(t *testing.T) {
	s := NewStore()
	// Must not panic on an unknown ID.
	s.Update("does-not-exist", StatusFailed, "", nil, nil)
}

func TestGet_NotFound(t *testing.T) {
	s := NewStore()
	if _, ok := s.Get("nope"); ok {
		t.Error("expected ok=false for unknown task ID")
	}
}

func TestGet_ReturnsCopyNotAlias(t *testing.T) {
	s := NewStore()
	task, _ := s.Create("t", context.Background())

	got, _ := s.Get(task.ID)
	got.Message = "mutated locally"

	again, _ := s.Get(task.ID)
	if again.Message == "mutated locally" {
		t.Error("Get returned an alias to internal state instead of a copy")
	}
}

func TestList(t *testing.T) {
	s := NewStore()
	if len(s.List()) != 0 {
		t.Error("expected empty list for a fresh store")
	}

	ids := map[string]bool{}
	for i := 0; i < 3; i++ {
		task, _ := s.Create("t", context.Background())
		ids[task.ID] = true
	}

	list := s.List()
	if len(list) != 3 {
		t.Fatalf("List() len = %d, want 3", len(list))
	}
	for _, task := range list {
		if !ids[task.ID] {
			t.Errorf("unexpected task ID in list: %s", task.ID)
		}
		task.Message = "mutated"
	}
	for _, task := range s.List() {
		if task.Message == "mutated" {
			t.Error("List returned aliases to internal state instead of copies")
		}
	}
}

func TestCancel(t *testing.T) {
	s := NewStore()
	task, taskCtx := s.Create("t", context.Background())

	if !s.Cancel(task.ID) {
		t.Fatal("expected Cancel to succeed on a working task")
	}

	got, _ := s.Get(task.ID)
	if got.Status != StatusCancelled {
		t.Errorf("Status = %q, want %q", got.Status, StatusCancelled)
	}

	select {
	case <-taskCtx.Done():
	default:
		t.Error("expected taskCtx to be cancelled after Cancel")
	}

	// Cancelling an already-cancelled (no longer working) task is a no-op.
	if s.Cancel(task.ID) {
		t.Error("expected second Cancel on a non-working task to return false")
	}
}

func TestCancel_UnknownID(t *testing.T) {
	s := NewStore()
	if s.Cancel("nope") {
		t.Error("expected Cancel to return false for unknown task ID")
	}
}

func TestRun_Success(t *testing.T) {
	s := NewStore()
	task, ctx := s.Create("t", context.Background())

	done := make(chan struct{})
	s.Run(task.ID, ctx, func(ctx context.Context) {
		s.Update(task.ID, StatusSucceeded, "ok", nil, nil)
		close(done)
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not execute fn in time")
	}

	got, _ := s.Get(task.ID)
	if got.Status != StatusSucceeded {
		t.Errorf("Status = %q, want %q", got.Status, StatusSucceeded)
	}
}

func TestRun_PanicIsRecoveredAsFailed(t *testing.T) {
	s := NewStore()
	task, ctx := s.Create("t", context.Background())

	s.Run(task.ID, ctx, func(ctx context.Context) {
		panic("kaboom")
	})

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, _ := s.Get(task.ID)
		if got.Status == StatusFailed {
			if got.Error == "" {
				t.Error("expected Error to be populated after a recovered panic")
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("task never reached StatusFailed after fn panicked")
}

func TestPurgeOldTasks(t *testing.T) {
	s := NewStore()

	oldFinished, _ := s.Create("t", context.Background())
	s.Update(oldFinished.ID, StatusSucceeded, "", nil, nil)

	recentFinished, _ := s.Create("t", context.Background())
	s.Update(recentFinished.ID, StatusSucceeded, "", nil, nil)

	oldWorking, _ := s.Create("t", context.Background())

	// Reach into internal state to simulate age without waiting a real hour;
	// this file lives in package tasks so it can touch unexported fields directly.
	s.mu.Lock()
	s.tasks[oldFinished.ID].UpdatedAt = time.Now().Add(-2 * taskRetention)
	s.tasks[oldWorking.ID].UpdatedAt = time.Now().Add(-2 * taskRetention)
	s.mu.Unlock()

	s.purgeOldTasks()

	if _, ok := s.Get(oldFinished.ID); ok {
		t.Error("expected old finished task to be purged")
	}
	if _, ok := s.Get(recentFinished.ID); !ok {
		t.Error("expected recent finished task to survive purge")
	}
	if _, ok := s.Get(oldWorking.ID); !ok {
		t.Error("expected old but still-working task to survive purge (only finished tasks expire)")
	}
}

// TestStartJanitor only exercises setup and the ctx.Done() shutdown path — the
// ticker.C fire path (calling purgeOldTasks from the janitor loop) runs on a
// real 5-minute period (janitorPeriod) that isn't parameterized for tests, so
// it isn't exercised here; purgeOldTasks itself is fully covered directly by
// TestPurgeOldTasks above.
func TestStartJanitor_StopsOnContextCancel(t *testing.T) {
	s := NewStore()
	ctx, cancel := context.WithCancel(context.Background())

	s.StartJanitor(ctx)
	cancel()

	// Give the goroutine a moment to observe cancellation; run with -race to
	// catch any concurrent-access issues in the janitor loop.
	time.Sleep(50 * time.Millisecond)
}
