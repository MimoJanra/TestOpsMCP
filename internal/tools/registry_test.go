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

func callTool(r *Registry, name string, input string) (any, error) {
	tool := r.GetTool(name)
	if tool == nil {
		return nil, nil
	}
	return tool.Handler(context.Background(), json.RawMessage(input))
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
		"find_project",
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
		"bulk_clone_test_cases",
		"add_test_cases_to_launch",
		"add_test_plan_to_launch",
		"remove_test_cases_from_launch",
		"create_test_case_step",
		"update_test_case_step",
		"delete_test_case_step",
		"clone_test_case",
		"copy_launch",
		"resolve_test_result",
		"unmute_test_result",
		"get_launch_environment",
		"get_test_case_history",
		"get_launch_defects",
		"get_test_case_defects",
		"merge_launches",
		"add_test_case_defect",
		"remove_test_case_defect",
		"get_test_case_members",
		"add_test_case_members",
		"remove_test_case_members",
		"get_test_case_external_links",
		"add_test_case_external_link",
		"delete_test_case_external_link",
		"restore_test_case",
		"get_test_case_custom_fields",
		"update_test_case_custom_fields",
		// Extra test-case tools
		"get_test_case_tags",
		"set_test_case_tags",
		"get_test_case_issues",
		"set_test_case_issues",
		"get_test_case_examples",
		"set_test_case_examples",
		"list_test_case_versions",
		"create_test_case_version",
		"restore_test_case_version",
		"get_test_case_attachments",
		"delete_test_case_attachment",
		"search_test_cases",
		"list_deleted_test_cases",
		"delete_test_case_scenario",
		"move_test_case_step",
		"copy_test_case_step",
		"get_test_case_relations",
		"set_test_case_relations",
		// Bulk test-case operations
		"bulk_add_test_case_members",
		"bulk_remove_test_case_members",
		"bulk_add_test_case_custom_fields",
		"bulk_remove_test_case_custom_fields",
		"bulk_add_test_case_external_links",
		"bulk_add_test_case_issues",
		"bulk_remove_test_case_issues",
		"bulk_set_test_case_layer",
		"bulk_move_test_cases",
		"bulk_delete_test_cases",
		"bulk_run_test_cases_new_launch",
		"bulk_run_test_cases_existing_launch",
		"bulk_create_test_plan",
		"bulk_mute_test_cases",
		// Extra single test-case operations
		"list_muted_test_cases",
		"get_test_case_audit",
		"validate_test_case_query",
		"suggest_test_cases",
		"get_test_case_workflow",
		"get_test_case_keys",
		"set_test_case_keys",
		"get_test_case_scenario_from_run",
		"detach_test_case_automation",
		"get_test_case_version_data",
		"delete_test_case_version",
		// Task management tools
		"get_task_status",
		"list_running_tasks",
		"cancel_task",
		// Analysis tools
		"analyze_launch_failures",
	} {
		if r.GetTool(name) == nil {
			t.Errorf("tool %q not registered", name)
		}
	}

	// Check for OpenAPI-based tools (only registered if spec is found)
	has_search := r.GetTool("search_testops_operations") != nil
	has_execute := r.GetTool("execute_testops_operation") != nil

	// newTestRegistry uses a nil Allure client, so the 2 tools gated on r.allure
	// != nil (configure_allure_token, get_launch_dashboard) never register here
	// regardless of this count — see TestDocsToolCountMatchesRegistry for the
	// fully-configured count (117) that end users actually see.
	// Expected count: 113 base tools + 2 OpenAPI tools (if spec found)
	expected_count := 113
	if has_search && has_execute {
		expected_count = 115
	}

	if got := len(r.ListTools()); got != expected_count {
		t.Errorf("ListTools() count = %d, want %d", got, expected_count)
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
			if _, err := callTool(r, "run_allure_launch", tc.input); err == nil {
				t.Errorf("expected error for input %q", tc.input)
			}
		})
	}
}

func TestGetLaunchStatus_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"launch_id":0}`, `{"launch_id":-1}`, `not-json`} {
		if _, err := callTool(r, "get_launch_status", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetLaunchReport_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	if _, err := callTool(r, "get_launch_report", `{"launch_id":0}`); err == nil {
		t.Error("expected error for non-positive launch_id")
	}
}

func TestCloseLaunch_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"launch_id":0}`, `{"launch_id":-1}`, `not-json`} {
		if _, err := callTool(r, "close_launch", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestReopenLaunch_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"launch_id":0}`, `{"launch_id":-1}`, `not-json`} {
		if _, err := callTool(r, "reopen_launch", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestListLaunches_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"project_id":0}`, `{"project_id":-1}`, `not-json`} {
		if _, err := callTool(r, "list_launches", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetLaunchDetails_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"launch_id":0}`, `{"launch_id":-1}`, `not-json`} {
		if _, err := callTool(r, "get_launch_details", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestListTestResults_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"launch_id":0}`, `{"launch_id":-1}`, `not-json`} {
		if _, err := callTool(r, "list_test_results", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetTestResult_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"test_result_id":0}`, `{"test_result_id":-1}`, `not-json`} {
		if _, err := callTool(r, "get_test_result", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestAssignTestResult_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	cases := []string{
		`{}`,
		`{"test_result_id":1}`,
		`{"username":"john"}`,
		`{"test_result_id":0,"username":"john"}`,
		`{"test_result_id":1,"username":""}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := callTool(r, "assign_test_result", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestMuteTestResult_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"test_result_id":0}`, `{"test_result_id":-1}`, `not-json`} {
		if _, err := callTool(r, "mute_test_result", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestListTestCases_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"project_id":0}`, `{"project_id":-1}`, `not-json`} {
		if _, err := callTool(r, "list_test_cases", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"test_case_id":0}`, `{"test_case_id":-1}`, `not-json`} {
		if _, err := callTool(r, "get_test_case", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestRunTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	cases := []string{
		`{}`,
		`{"test_case_id":1}`,
		`{"launch_id":1}`,
		`{"test_case_id":0,"launch_id":1}`,
		`{"test_case_id":1,"launch_id":0}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := callTool(r, "run_test_case", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestListProjects_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	if _, err := callTool(r, "list_projects", `not-json`); err == nil {
		t.Error("expected error for bad json")
	}
}

func TestRemoveTestCasesFromLaunch_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	cases := []string{
		`{}`,
		`{"launch_id":0,"test_case_ids":[1]}`,
		`{"launch_id":-1,"test_case_ids":[1]}`,
		`{"launch_id":1}`,
		`{"launch_id":1,"test_case_ids":[]}`,
		`{"launch_id":1,"test_case_ids":[0]}`,
		`{"launch_id":1,"test_case_ids":[1],"mode":"bogus"}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := callTool(r, "remove_test_cases_from_launch", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestFindProject_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"query":""}`, `{"query":"   "}`, `not-json`} {
		if _, err := callTool(r, "find_project", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetProject_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"project_id":0}`, `{"project_id":-1}`, `not-json`} {
		if _, err := callTool(r, "get_project", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetProjectStats_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"project_id":0}`, `{"project_id":-1}`, `not-json`} {
		if _, err := callTool(r, "get_project_stats", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetLaunchTrendAnalytics_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"project_id":0}`, `{"project_id":-1}`, `not-json`} {
		if _, err := callTool(r, "get_launch_trend_analytics", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetLaunchDurationAnalytics_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"project_id":0}`, `{"project_id":-1}`, `not-json`} {
		if _, err := callTool(r, "get_launch_duration_analytics", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestGetTestSuccessRate_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"project_id":0}`, `{"project_id":-1}`, `not-json`} {
		if _, err := callTool(r, "get_test_success_rate", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestCreateTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	cases := []string{
		`{}`,
		`{"project_id":1}`,
		`{"name":"test"}`,
		`{"project_id":0,"name":"test"}`,
		`{"project_id":-1,"name":"test"}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := callTool(r, "create_test_case", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestUpdateTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	cases := []string{
		`{}`,
		`{"test_case_id":0}`,
		`{"test_case_id":-1}`,
		`{"test_case_id":1}`,
		`not-json`,
	}
	for _, in := range cases {
		if _, err := callTool(r, "update_test_case", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

func TestDeleteTestCase_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)

	for _, in := range []string{`{}`, `{"test_case_id":0}`, `{"test_case_id":-1}`, `not-json`} {
		if _, err := callTool(r, "delete_test_case", in); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}
