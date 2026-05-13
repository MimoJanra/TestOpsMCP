package session

import "context"

type idKey struct{}

// StdioID is the fixed session identifier used in stdio mode (single user).
const StdioID = "stdio"

// WithID returns a new context carrying the given MCP session ID.
func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, idKey{}, id)
}

// IDFromContext returns the MCP session ID stored in ctx, or "" if absent.
func IDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(idKey{}).(string)
	return id
}
