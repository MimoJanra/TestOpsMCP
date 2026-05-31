package tasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
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

func (s *Store) Create(toolName string) (*Task, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
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
	return t, ctx
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

func (s *Store) Get(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *Store) List() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		list = append(list, t)
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

func newTaskID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}
