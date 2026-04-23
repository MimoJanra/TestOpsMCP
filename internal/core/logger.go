package core

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Logger struct {
}

func NewLogger() *Logger {
	return &Logger{}
}

type logEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

func (l *Logger) Info(message string, data interface{}) {
	l.log("INFO", message, data)
}

func (l *Logger) Error(message string, err error, data interface{}) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	logData := map[string]interface{}{
		"error": errStr,
	}
	if data != nil {
		if m, ok := data.(map[string]interface{}); ok {
			for k, v := range m {
				logData[k] = v
			}
		}
	}
	l.log("ERROR", message, logData)
}

func (l *Logger) Warn(message string, data interface{}) {
	l.log("WARN", message, data)
}

func (l *Logger) log(level, message string, data interface{}) {
	entry := logEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Data:      data,
	}
	
	bytes, _ := json.Marshal(entry)
	fmt.Fprintln(os.Stderr, string(bytes))
}
