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
	for _, name := range []string{
		"run_allure_launch",
		"get_launch_status",
		"get_launch_report",
		"close_launch",
		"reopen_launch",
		"list_launches",
		"get_launch_details",
		"list_test_results",
		"get_test_result",
		"assign_test_result",
		"mute_test_result",
		"list_test_cases",
		"get_test_case",
		"run_test_case",
		"create_test_case",
		"update_test_case",
		"delete_test_case",
		"list_projects",
		"get_project",
		"get_project_stats",
		"get_launch_trend_analytics",
		"get_launch_duration_analytics",
		"get_test_success_rate",
	} {
		if r.GetTool(name) == nil {
			t.Errorf("tool %q not registered", name)
		}
	}
	for _, name := range []string{
		"bulk_set_test_case_status",
		"bulk_add_test_case_tags",
		"bulk_remove_test_case_tags",
		"bulk_assign_test_results",
		"bulk_mute_test_results",
		"bulk_unmute_test_results",
		"bulk_resolve_test_results",
		"add_test_cases_to_launch",
		"add_test_plan_to_launch",
	} {
		if r.GetTool(name) == nil {
			t.Errorf("tool %q not registered", name)
		}
	}
	if got := len(r.ListTools()); got != 32 {
		t.Errorf("ListTools() count = %d, want 32", got)
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

func TestCloseLaunch_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"launch_id":0}`,
		`{"launch_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.closeLaunch(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestReopenLaunch_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"launch_id":0}`,
		`{"launch_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.reopenLaunch(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestListLaunches_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"project_id":0}`,
		`{"project_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.listLaunches(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetLaunchDetails_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"launch_id":0}`,
		`{"launch_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.getLaunchDetails(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestListTestResults_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"launch_id":0}`,
		`{"launch_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.listTestResults(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetTestResult_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"test_result_id":0}`,
		`{"test_result_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.getTestResult(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestAssignTestResult_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"test_result_id":1}`,
		`{"username":"john"}`,
		`{"test_result_id":0,"username":"john"}`,
		`{"test_result_id":1,"username":""}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.assignTestResult(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestMuteTestResult_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"test_result_id":0}`,
		`{"test_result_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.muteTestResult(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestListTestCases_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"project_id":0}`,
		`{"project_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.listTestCases(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"test_case_id":0}`,
		`{"test_case_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.getTestCase(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestRunTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"test_case_id":1}`,
		`{"launch_id":1}`,
		`{"test_case_id":0,"launch_id":1}`,
		`{"test_case_id":1,"launch_id":0}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.runTestCase(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestListProjects_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	if _, err := r.listProjects(ctx, json.RawMessage(`not-json`)); err == nil {
		t.Error("expected error for bad json")
	}
}

func TestGetProject_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"project_id":0}`,
		`{"project_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.getProject(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetProjectStats_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"project_id":0}`,
		`{"project_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.getProjectStats(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetLaunchTrendAnalytics_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"project_id":0}`,
		`{"project_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.getLaunchTrendAnalytics(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetLaunchDurationAnalytics_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"project_id":0}`,
		`{"project_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.getLaunchDurationAnalytics(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetTestSuccessRate_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"project_id":0}`,
		`{"project_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.getTestSuccessRate(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestCreateTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"project_id":1}`,
		`{"name":"test"}`,
		`{"project_id":0,"name":"test"}`,
		`{"project_id":-1,"name":"test"}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.createTestCase(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestUpdateTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"test_case_id":0}`,
		`{"test_case_id":-1}`,
		`{"test_case_id":1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.updateTestCase(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestDeleteTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	ctx := context.Background()

	cases := []string{
		`{}`,
		`{"test_case_id":0}`,
		`{"test_case_id":-1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := r.deleteTestCase(ctx, json.RawMessage(in)); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}
