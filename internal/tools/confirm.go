package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MimoJanra/TestOpsMCP/internal/session"
)

// confirmDestructive asks the user to confirm an irreversible action over MCP
// elicitation. ok is false when the action must not proceed; resp/err then hold
// what the tool should return.
//
// The distinction between the two "not ok" outcomes matters: resp means the
// user saw the question and said no, err means the server could not ask at all
// (no interactive session, a client without elicitation support, or no open
// server-to-client stream). Reporting the second as the first would tell the
// user they declined something they were never shown.
func confirmDestructive(ctx context.Context, message string) (ok bool, resp any, err error) {
	elicit, has := session.ElicitFromContext(ctx)
	if !has {
		return false, nil, fmt.Errorf("this action requires user confirmation but no interactive session is available")
	}
	schema, _ := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"confirmed": map[string]any{"type": "boolean", "description": "Confirm"},
		},
	})
	result, err := elicit(ctx, message, schema)
	if err != nil {
		return false, nil, fmt.Errorf("confirmation failed: %w", err)
	}
	if result.Action != "accept" {
		return false, map[string]any{"cancelled": true, "message": "Cancelled."}, nil
	}
	return true, nil, nil
}
