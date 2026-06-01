package tools

import "fmt"

func (r *Registry) registerPrompts() {
	r.RegisterPrompt(&RegistryPrompt{
		Name:        "test-case-management",
		Description: "Full workflow for managing test cases in a project: browsing, creating, editing, organizing into folders, and cleanup",
		Arguments: []PromptArg{
			{Name: "project_id", Description: "Allure project ID to work in", Required: true},
			{Name: "goal", Description: "What you want to do (e.g. 'clean up step formatting', 'organize by feature', 'create test cases for login')", Required: false},
		},
		GetMessages: func(args map[string]string) []RegistryPromptMessage {
			projectID := args["project_id"]
			goal := args["goal"]
			if goal == "" {
				goal = "review and organize the test cases"
			}
			text := fmt.Sprintf(
				"Work with test cases in project %s. Goal: %s\n\n"+
					"Follow this workflow:\n\n"+
					"1. **Explore** — use browse_test_case_tree (empty path = root) to see the folder structure, "+
					"then get_test_case_tree_folders to go deeper. Use list_test_cases or search_test_cases to find specific test cases.\n\n"+
					"2. **Read before writing** — always call get_test_case to see current content before updating. "+
					"For steps, call get_test_case_scenario to read the full step tree, "+
					"or get_test_case_steps when you need step IDs for move/delete operations.\n\n"+
					"3. **Edit** — use update_test_case to change name, description, precondition, or expected_result. "+
					"Use create_test_case_step / update_test_case_step / delete_test_case_step for step-level edits. "+
					"Use move_test_case_step to reorder steps within a scenario.\n\n"+
					"4. **Organize** — use move_test_cases_to_folder to move test cases between folders. "+
					"Use create_test_case_folder to create new folders. "+
					"Use bulk_set_test_case_status / bulk_add_test_case_tags for mass updates.\n\n"+
					"5. **Safety** — before bulk edits, create a version snapshot with create_test_case_version. "+
					"Deleted test cases go to trash (recoverable with restore_test_case).\n\n"+
					"Start by exploring the project structure, then proceed with the goal.",
				projectID, goal,
			)
			return []RegistryPromptMessage{{Role: "user", Text: text}}
		},
	})

	r.RegisterPrompt(&RegistryPrompt{
		Name:        "analyze-test-failures",
		Description: "Analyze failed, broken, and skipped tests in an Allure launch to identify patterns and root causes",
		Arguments: []PromptArg{
			{Name: "launch_id", Description: "Allure launch ID to analyze", Required: true},
			{Name: "project_id", Description: "Allure project ID for additional context", Required: false},
		},
		GetMessages: func(args map[string]string) []RegistryPromptMessage {
			launchID := args["launch_id"]
			if launchID == "" {
				launchID = "<launch_id>"
			}
			text := fmt.Sprintf(
				"Please analyze the test failures in launch %s.\n\n"+
					"Use the available TestOps tools to:\n"+
					"1. Get the launch details and overall statistics\n"+
					"2. List failed and broken test results\n"+
					"3. Group failures by error message or test suite to identify common patterns\n"+
					"4. Highlight the most critical issues and suggest next steps for the team",
				launchID,
			)
			return []RegistryPromptMessage{{Role: "user", Text: text}}
		},
	})

	r.RegisterPrompt(&RegistryPrompt{
		Name:        "launch-report-summary",
		Description: "Generate a concise executive summary for an Allure TestOps launch",
		Arguments: []PromptArg{
			{Name: "launch_id", Description: "Allure launch ID", Required: true},
			{Name: "project_id", Description: "Allure project ID", Required: false},
		},
		GetMessages: func(args map[string]string) []RegistryPromptMessage {
			launchID := args["launch_id"]
			if launchID == "" {
				launchID = "<launch_id>"
			}
			text := fmt.Sprintf(
				"Generate a concise executive summary for launch %s.\n\n"+
					"Use the available TestOps tools to retrieve launch details and statistics, then provide:\n"+
					"- Overall pass rate and test counts (passed / failed / broken / skipped)\n"+
					"- Launch duration and environment\n"+
					"- Brief status assessment (stable / needs attention / critical)\n"+
					"- Top 3 failures if any exist",
				launchID,
			)
			return []RegistryPromptMessage{{Role: "user", Text: text}}
		},
	})
}
