package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Entry is one audited MCP request.
type Entry struct {
	Timestamp  string `json:"timestamp"`
	User       string `json:"user"`
	SessionID  string `json:"session_id,omitempty"`
	RemoteAddr string `json:"remote_addr,omitempty"`
	Method     string `json:"method"`
	Tool       string `json:"tool,omitempty"`
	Status     string `json:"status"` // "ok" | "error"
	DurationMS int64  `json:"duration_ms"`
}

// Logger writes audit entries to daily JSONL files under dir and
// purges files older than retentionDays.
type Logger struct {
	dir           string
	retentionDays int
	mu            sync.Mutex
	stop          chan struct{}
}

// NewLogger creates the log directory and starts the cleanup goroutine.
func NewLogger(dir string, retentionDays int) (*Logger, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	l := &Logger{
		dir:           dir,
		retentionDays: retentionDays,
		stop:          make(chan struct{}),
	}
	l.cleanup()
	go l.cleanupLoop()
	return l, nil
}

// Write appends an entry to today's log file.
func (l *Logger) Write(e Entry) {
	e.Timestamp = time.Now().UTC().Format(time.RFC3339)
	data, err := json.Marshal(e)
	if err != nil {
		return
	}

	filename := filepath.Join(l.dir, "audit-"+time.Now().UTC().Format("2006-01-02")+".jsonl")

	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(data, '\n'))
}

// Close stops the background cleanup goroutine.
func (l *Logger) Close() {
	close(l.stop)
}

func (l *Logger) cleanupLoop() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.cleanup()
		case <-l.stop:
			return
		}
	}
}

func (l *Logger) cleanup() {
	cutoff := time.Now().UTC().AddDate(0, 0, -l.retentionDays)
	entries, _ := os.ReadDir(l.dir)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "audit-") || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		dateStr := strings.TrimSuffix(strings.TrimPrefix(name, "audit-"), ".jsonl")
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			_ = os.Remove(filepath.Join(l.dir, name))
		}
	}
}
