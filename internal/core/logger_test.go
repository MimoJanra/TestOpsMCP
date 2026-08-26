package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]Level{
		"DEBUG":   LevelDebug,
		"debug":   LevelDebug,
		" Debug ": LevelDebug,
		"WARN":    LevelWarn,
		"warn":    LevelWarn,
		"WARNING": LevelWarn,
		"warning": LevelWarn,
		"ERROR":   LevelError,
		"error":   LevelError,
		"INFO":    LevelInfo,
		"info":    LevelInfo,
		"":        LevelInfo,
		"bogus":   LevelInfo,
	}
	for input, want := range cases {
		if got := ParseLevel(input); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", input, got, want)
		}
	}
}

// bufferLogger builds a Logger backed by a buffer instead of os.Stderr so
// tests can assert on the JSON output. It relies on being in package core to
// reach the unexported slog field.
func bufferLogger(buf *bytes.Buffer, level Level) *Logger {
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})
	return &Logger{slog: slog.New(handler)}
}

func decodeLastLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	last := lines[len(lines)-1]
	var m map[string]any
	if err := json.Unmarshal([]byte(last), &m); err != nil {
		t.Fatalf("failed to decode log line %q: %v", last, err)
	}
	return m
}

func TestLogger_Info_WithMapData(t *testing.T) {
	var buf bytes.Buffer
	l := bufferLogger(&buf, LevelInfo)

	l.Info("hello", map[string]any{"foo": "bar", "n": 3})

	entry := decodeLastLine(t, &buf)
	if entry["msg"] != "hello" {
		t.Errorf("msg = %v, want hello", entry["msg"])
	}
	if entry["foo"] != "bar" {
		t.Errorf("foo = %v, want bar", entry["foo"])
	}
	if entry["n"] != float64(3) {
		t.Errorf("n = %v, want 3", entry["n"])
	}
}

func TestLogger_Debug_WithNonMapData(t *testing.T) {
	var buf bytes.Buffer
	l := bufferLogger(&buf, LevelDebug)

	l.Debug("plain", 42)

	entry := decodeLastLine(t, &buf)
	if entry["msg"] != "plain" {
		t.Errorf("msg = %v, want plain", entry["msg"])
	}
	if entry["data"] != float64(42) {
		t.Errorf("data = %v, want 42", entry["data"])
	}
}

func TestLogger_Warn_WithNilData(t *testing.T) {
	var buf bytes.Buffer
	l := bufferLogger(&buf, LevelDebug)

	l.Warn("no data", nil)

	entry := decodeLastLine(t, &buf)
	if entry["msg"] != "no data" {
		t.Errorf("msg = %v, want 'no data'", entry["msg"])
	}
	if _, ok := entry["data"]; ok {
		t.Errorf("did not expect a 'data' key for nil data, got %v", entry)
	}
}

func TestLogger_Error_WithErrAndMapData(t *testing.T) {
	var buf bytes.Buffer
	l := bufferLogger(&buf, LevelDebug)

	l.Error("boom", errors.New("kaboom"), map[string]any{"ctx": "x"})

	entry := decodeLastLine(t, &buf)
	if entry["msg"] != "boom" {
		t.Errorf("msg = %v, want boom", entry["msg"])
	}
	if entry["error"] != "kaboom" {
		t.Errorf("error = %v, want kaboom", entry["error"])
	}
	if entry["ctx"] != "x" {
		t.Errorf("ctx = %v, want x", entry["ctx"])
	}
}

func TestLogger_Error_WithNilErr(t *testing.T) {
	var buf bytes.Buffer
	l := bufferLogger(&buf, LevelDebug)

	l.Error("no err", nil, nil)

	entry := decodeLastLine(t, &buf)
	if _, ok := entry["error"]; ok {
		t.Errorf("did not expect an 'error' key when err is nil, got %v", entry)
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := bufferLogger(&buf, LevelError)

	l.Debug("debug msg", nil)
	l.Info("info msg", nil)
	l.Warn("warn msg", nil)

	if buf.Len() != 0 {
		t.Fatalf("expected no output below LevelError, got: %s", buf.String())
	}

	l.Error("error msg", nil, nil)
	if buf.Len() == 0 {
		t.Fatal("expected output for Error at LevelError")
	}
}

func TestNewLogger_WritesJSONToStderr(t *testing.T) {
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = origStderr }()

	l := NewLogger(LevelInfo)
	l.Info("via new logger", map[string]any{"k": "v"})

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read pipe: %v", err)
	}

	entry := decodeLastLine(t, &buf)
	if entry["msg"] != "via new logger" {
		t.Errorf("msg = %v, want 'via new logger'", entry["msg"])
	}
	if entry["k"] != "v" {
		t.Errorf("k = %v, want v", entry["k"])
	}
}
