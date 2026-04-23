package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func ParseLevel(s string) Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		return LevelDebug
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

type Logger struct {
	mu    sync.Mutex
	level Level
}

func NewLogger(level Level) *Logger {
	return &Logger{level: level}
}

type logEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
}

func (l *Logger) Debug(message string, data any) {
	l.log(LevelDebug, message, data)
}

func (l *Logger) Info(message string, data any) {
	l.log(LevelInfo, message, data)
}

func (l *Logger) Warn(message string, data any) {
	l.log(LevelWarn, message, data)
}

func (l *Logger) Error(message string, err error, data any) {
	logData := map[string]any{}
	if m, ok := data.(map[string]any); ok {
		for k, v := range m {
			logData[k] = v
		}
	} else if data != nil {
		logData["data"] = data
	}
	if err != nil {
		logData["error"] = err.Error()
	}
	l.log(LevelError, message, logData)
}

func (l *Logger) log(level Level, message string, data any) {
	if level < l.level {
		return
	}
	entry := logEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Message:   message,
		Data:      data,
	}

	bytes, err := json.Marshal(entry)
	l.mu.Lock()
	defer l.mu.Unlock()
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"timestamp":%q,"level":%q,"message":%q,"error":"log marshal failed: %s"}`+"\n",
			entry.Timestamp, entry.Level, entry.Message, err.Error())
		return
	}
	fmt.Fprintln(os.Stderr, string(bytes))
}
