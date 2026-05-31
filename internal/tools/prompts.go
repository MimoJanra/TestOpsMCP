package tools

import "fmt"

func (r *Registry) registerPrompts() {
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
