package tasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

const (
	taskTimeout   = 30 * time.Minute
	taskRetention = 1 * time.Hour  // how long finished tasks are kept before purge
	janitorPeriod = 5 * time.Minute
)

type Status string

const (
	StatusWorking   Status = "working"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Task struct {
	ID        string    `json:"id"`
	Tool      string    `json:"tool"`
	Status    Status    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Result    any       `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	cancel context.CancelFunc
}

type Store struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

func NewStore() *Store {
	return &Store{tasks: make(map[string]*Task)}
}

// Create registers a new task and returns a context that carries values from
// parentCtx (e.g. session ID for token resolution) but is not cancelled when
// parentCtx is cancelled. A 30-minute hard timeout is applied.
func (s *Store) Create(toolName string, parentCtx context.Context) (*Task, context.Context) {
	// context.WithoutCancel propagates all key-value pairs (session ID, etc.)
	// without inheriting the parent's cancellation signal.
	base := context.WithoutCancel(parentCtx)
	taskCtx, cancel := context.WithTimeout(base, taskTimeout)
	t := &Task{
		ID:        newTaskID(),
		Tool:      toolName,
		Status:    StatusWorking,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		cancel:    cancel,
	}
	s.mu.Lock()
	s.tasks[t.ID] = t
	s.mu.Unlock()
	return t, taskCtx
}

// Run starts fn in a background goroutine. A panic inside fn marks the task
// as Failed (instead of crashing the whole server).
func (s *Store) Run(id string, ctx context.Context, fn func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.Update(id, StatusFailed, "", nil,
					fmt.Errorf("panic: %v\n%s", r, debug.Stack()))
			}
		}()
		fn(ctx)
	}()
}

func (s *Store) Update(id string, status Status, msg string, result any, taskErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return
	}
	t.Status = status
	t.Message = msg
	t.Result = result
	if taskErr != nil {
		t.Error = taskErr.Error()
	}
	t.UpdatedAt = time.Now()
}

// Get returns a copy of the task so callers can read fields without racing
// against concurrent Update calls.
func (s *Store) Get(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, false
	}
	cp := *t
	return &cp, true
}

// List returns copies of all tasks for the same reason as Get.
func (s *Store) List() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		cp := *t
		list = append(list, &cp)
	}
	return list
}

func (s *Store) Cancel(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok || t.Status != StatusWorking {
		return false
	}
	t.Status = StatusCancelled
	t.UpdatedAt = time.Now()
	if t.cancel != nil {
		t.cancel()
	}
	return true
}

// StartJanitor starts a background goroutine that purges finished tasks older
// than taskRetention. It stops when ctx is cancelled.
func (s *Store) StartJanitor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(janitorPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.purgeOldTasks()
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Store) purgeOldTasks() {
	cutoff := time.Now().Add(-taskRetention)
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, t := range s.tasks {
		if t.Status != StatusWorking && t.UpdatedAt.Before(cutoff) {
			delete(s.tasks, id)
		}
	}
}

func newTaskID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}
