package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/MimoJanra/TestOpsMCP/internal/core"
)

func newTestRegistry(t *testing.T) *Registry {
	t.Helper()
	logger := core.NewLogger(core.LevelError)
	return NewRegistry(nil, logger)
}

func TestRegistry_HasExpectedTools(t *testing.T) {
	r := newTestRegistry(t)
	for _, name := range []string{"run_allure_launch", "get_launch_status", "get_launch_report"} {
		if r.GetTool(name) == nil {
			t.Errorf("tool %q not registered", name)
		}
	}
	if got := len(r.ListTools()); got != 3 {
		t.Errorf("ListTools() count = %d, want 3", got)
	}
}

func TestRegistry_GetToolUnknown(t *testing.T) {
	r := newTestRegistry(t)
	if r.GetTool("nope") != nil {
		t.Error("unknown tool should return nil")
	}
}

func TestRunAllureLaunch_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []struct {
		name  string
		input string
	}{
		{"missing launch_name", `{"project_id":1}`},
		{"non-positive project_id", `{"project_id":0,"launch_name":"x"}`},
		{"negative project_id", `{"project_id":-5,"launch_name":"x"}`},
		{"bad json", `not-json`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := r.runAllureLaunch(ctx, json.RawMessage(tc.input))
			if err == nil {
				t.Errorf("expected error for input %q", tc.input)
			}
		})
	}
}

func TestGetLaunchStatus_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"launch_id":0}`,
		`{"launch_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.getLaunchStatus(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetLaunchReport_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	if _, err := r.getLaunchReport(ctx, json.RawMessage(`{"launch_id":0}`)); err == nil {
		t.Error("expected error for non-positive launch_id")
	}
}
