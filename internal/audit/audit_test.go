package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWrite_ProducesValidJSONLWithExpectedFields(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir, 30)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer l.Close()

	l.Write(Entry{
		User:       "alk",
		SessionID:  "sess-1",
		RemoteAddr: "127.0.0.1",
		Method:     "tools/call",
		Tool:       "list_launches",
		Status:     "ok",
		DurationMS: 12,
	})

	filename := filepath.Join(dir, "audit-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read audit file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d: %q", len(lines), data)
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &raw); err != nil {
		t.Fatalf("line is not valid JSON: %v", err)
	}

	wantKeys := map[string]bool{
		"timestamp": true, "user": true, "session_id": true, "remote_addr": true,
		"method": true, "tool": true, "status": true, "duration_ms": true,
	}
	for k := range raw {
		if !wantKeys[k] {
			t.Errorf("unexpected field %q in audit entry (possible secret leak): %v", k, raw[k])
		}
	}
	for k := range wantKeys {
		if _, ok := raw[k]; !ok {
			t.Errorf("missing expected field %q in audit entry", k)
		}
	}

	if raw["user"] != "alk" || raw["tool"] != "list_launches" || raw["status"] != "ok" {
		t.Errorf("unexpected entry content: %+v", raw)
	}
	if _, err := time.Parse(time.RFC3339, raw["timestamp"].(string)); err != nil {
		t.Errorf("timestamp %q is not RFC3339: %v", raw["timestamp"], err)
	}
}

func TestWrite_OmitsEmptyOptionalFields(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir, 30)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer l.Close()

	l.Write(Entry{User: "alk", Method: "tools/list", Status: "ok"})

	filename := filepath.Join(dir, "audit-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read audit file: %v", err)
	}
	line := strings.TrimSpace(string(data))
	for _, absent := range []string{`"session_id"`, `"remote_addr"`, `"tool"`} {
		if strings.Contains(line, absent) {
			t.Errorf("expected omitempty field %s to be absent, got line: %s", absent, line)
		}
	}
}

func TestWrite_AppendsMultipleEntriesToSameDayFile(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir, 30)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer l.Close()

	for i := 0; i < 3; i++ {
		l.Write(Entry{User: "alk", Method: "tools/call", Status: "ok"})
	}

	filename := filepath.Join(dir, "audit-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("open audit file: %v", err)
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}
	if count != 3 {
		t.Errorf("expected 3 lines, got %d", count)
	}
}

func TestNewLogger_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "audit-logs")
	l, err := NewLogger(dir, 30)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer l.Close()

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("expected directory %q to exist", dir)
	}
}

func TestCleanup_RemovesOnlyExpiredFiles(t *testing.T) {
	dir := t.TempDir()

	old := filepath.Join(dir, "audit-2000-01-01.jsonl")
	recent := filepath.Join(dir, "audit-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	malformed := filepath.Join(dir, "audit-not-a-date.jsonl")
	unrelated := filepath.Join(dir, "notes.txt")

	for _, p := range []string{old, recent, malformed, unrelated} {
		if err := os.WriteFile(p, []byte("{}\n"), 0o644); err != nil {
			t.Fatalf("seed file %s: %v", p, err)
		}
	}

	l := &Logger{dir: dir, retentionDays: 30}
	l.cleanup()

	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("expected expired file %s to be removed, stat err = %v", old, err)
	}
	for _, p := range []string{recent, malformed, unrelated} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to survive cleanup, stat err = %v", p, err)
		}
	}
}

func TestCleanup_ZeroRetentionRemovesEverythingIncludingToday(t *testing.T) {
	// cleanup's cutoff is "now" (time-of-day included), while a file's date is
	// parsed as midnight UTC of that calendar day. With retentionDays=0 the
	// cutoff is the current instant, so even today's file (dated at midnight)
	// is already "before" it and gets removed — retention 0 means "keep
	// nothing dated", not "keep today".
	dir := t.TempDir()
	yesterday := filepath.Join(dir, "audit-"+time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")+".jsonl")
	today := filepath.Join(dir, "audit-"+time.Now().UTC().Format("2006-01-02")+".jsonl")
	for _, p := range []string{yesterday, today} {
		if err := os.WriteFile(p, []byte("{}\n"), 0o644); err != nil {
			t.Fatalf("seed file %s: %v", p, err)
		}
	}

	l := &Logger{dir: dir, retentionDays: 0}
	l.cleanup()

	if _, err := os.Stat(yesterday); !os.IsNotExist(err) {
		t.Errorf("expected yesterday's file to be removed with 0 retention")
	}
	if _, err := os.Stat(today); !os.IsNotExist(err) {
		t.Errorf("expected today's file to also be removed with 0 retention")
	}
}

func TestClose_StopsCleanupLoop(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir, 30)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	done := make(chan struct{})
	go func() {
		l.Close()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not return in time; cleanupLoop goroutine may be stuck")
	}
}
