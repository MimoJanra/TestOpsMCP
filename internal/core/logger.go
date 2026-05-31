package core

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
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

type Logger struct {
	slog *slog.Logger
}

func NewLogger(level Level) *Logger {
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return &Logger{slog: slog.New(handler)}
}

func (l *Logger) Debug(message string, data any) {
	l.slog.LogAttrs(context.Background(), LevelDebug, message, toAttr(data)...)
}

func (l *Logger) Info(message string, data any) {
	l.slog.LogAttrs(context.Background(), LevelInfo, message, toAttr(data)...)
}

func (l *Logger) Warn(message string, data any) {
	l.slog.LogAttrs(context.Background(), LevelWarn, message, toAttr(data)...)
}

func (l *Logger) Error(message string, err error, data any) {
	attrs := toAttr(data)
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	l.slog.LogAttrs(context.Background(), LevelError, message, attrs...)
}

func toAttr(data any) []slog.Attr {
	m, ok := data.(map[string]any)
	if !ok || m == nil {
		if data != nil {
			return []slog.Attr{slog.Any("data", data)}
		}
		return nil
	}
	attrs := make([]slog.Attr, 0, len(m))
	for k, v := range m {
		attrs = append(attrs, slog.Any(k, v))
	}
	return attrs
}
