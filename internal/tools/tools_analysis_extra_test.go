package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
	"github.com/MimoJanra/TestOpsMCP/internal/session"
)

func TestTruncateRunes(t *testing.T) {
	cases := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"shorter than limit", "hello", 10, "hello"},
		{"exact limit", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello…"},
		{"multibyte truncated", "héllo wörld", 3, "hél…"},
		{"empty", "", 5, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := truncateRunes(tc.in, tc.n); got != tc.want {
				t.Errorf("truncateRunes(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
			}
		})
	}
}

func newAnalysisTestRegistry(t *testing.T, handler http.HandlerFunc) *Registry {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/uaa/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "fake-jwt", "expires_in": 3600})
			return
		}
		handler(w, req)
	}))
	t.Cleanup(server.Close)
	client := allure.NewClient(server.URL, "test-token", 5*time.Second)
	return NewRegistry(client, core.NewLogger(core.LevelError))
}

func TestAnalyzeLaunchFailures_ValidatesInput(t *testing.T) {
	r := newAnalysisTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	if _, err := r.analyzeLaunchFailures(context.Background(), analyzeLaunchFailuresArgs{LaunchID: 0}); err == nil {
		t.Error("expected error for non-positive launch_id")
	}
}

func TestAnalyzeLaunchFailures_RequiresSampling(t *testing.T) {
	r := newAnalysisTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_, err := r.analyzeLaunchFailures(context.Background(), analyzeLaunchFailuresArgs{LaunchID: 1})
	if err == nil || !strings.Contains(err.Error(), "sampling") {
		t.Errorf("expected sampling-not-available error, got %v", err)
	}
}

func TestAnalyzeLaunchFailures_NoFailures(t *testing.T) {
	r := newAnalysisTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []any{}, "empty": true, "last": true, "totalElements": 0,
		})
	})
	ctx := session.WithSampling(context.Background(), func(ctx context.Context, system string, messages []session.SamplingMessage, maxTokens int) (*session.SamplingResult, error) {
		t.Fatal("sampling should not be called when there are no failures")
		return nil, nil
	})
	result, err := r.analyzeLaunchFailures(ctx, analyzeLaunchFailuresArgs{LaunchID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["failures"] != 0 {
		t.Errorf("failures = %v, want 0", m["failures"])
	}
}

func TestAnalyzeLaunchFailures_WithSamplingSuccess(t *testing.T) {
	r := newAnalysisTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{
				{"id": 1, "name": "test A", "message": "boom", "trace": "stack trace here"},
			},
			"empty": false, "last": true, "totalElements": 1,
		})
	})

	var gotPrompt string
	ctx := session.WithSampling(context.Background(), func(ctx context.Context, system string, messages []session.SamplingMessage, maxTokens int) (*session.SamplingResult, error) {
		if len(messages) > 0 {
			gotPrompt = messages[0].Text
		}
		return &session.SamplingResult{Text: "root cause: flaky network"}, nil
	})

	result, err := r.analyzeLaunchFailures(ctx, analyzeLaunchFailuresArgs{LaunchID: 5, MaxFailures: 500})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := result.(map[string]any)
	if m["analysis"] != "root cause: flaky network" {
		t.Errorf("analysis = %v", m["analysis"])
	}
	if m["failures"] != 1 {
		t.Errorf("failures = %v, want 1", m["failures"])
	}
	if !strings.Contains(gotPrompt, "test A") || !strings.Contains(gotPrompt, "boom") {
		t.Errorf("prompt did not include failure details: %q", gotPrompt)
	}
}

func TestAnalyzeLaunchFailures_SamplingErrorFallsBackToSummary(t *testing.T) {
	r := newAnalysisTestRegistry(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]any{{"id": 1, "name": "test A", "message": "boom"}},
			"empty":   false, "last": true, "totalElements": 1,
		})
	})
	ctx := session.WithSampling(context.Background(), func(ctx context.Context, system string, messages []session.SamplingMessage, maxTokens int) (*session.SamplingResult, error) {
		return nil, fmt.Errorf("model unavailable")
	})

	result, err := r.analyzeLaunchFailures(ctx, analyzeLaunchFailuresArgs{LaunchID: 5})
	if err != nil {
		t.Fatalf("expected sampling failure to be reported in result, not returned as error: %v", err)
	}
	m := result.(map[string]any)
	if !strings.Contains(fmt.Sprintf("%v", m["error"]), "model unavailable") {
		t.Errorf("expected error field to mention sampling failure, got %v", m["error"])
	}
}

func TestAnalyzeLaunchFailures_ListResultsError(t *testing.T) {
	r := newAnalysisTestRegistry(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	})
	ctx := session.WithSampling(context.Background(), func(ctx context.Context, system string, messages []session.SamplingMessage, maxTokens int) (*session.SamplingResult, error) {
		t.Fatal("sampling should not be called when listing results fails")
		return nil, nil
	})
	if _, err := r.analyzeLaunchFailures(ctx, analyzeLaunchFailuresArgs{LaunchID: 5}); err == nil {
		t.Error("expected error when ListTestResults fails")
	}
}
