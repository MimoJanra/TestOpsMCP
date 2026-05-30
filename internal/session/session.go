package session

import "context"

type idKey struct{}
type userKey struct{}
type remoteAddrKey struct{}

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

// WithUser returns a new context carrying the authenticated user name.
func WithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

// UserFromContext returns the authenticated user name stored in ctx, or "anonymous".
func UserFromContext(ctx context.Context) string {
	user, _ := ctx.Value(userKey{}).(string)
	if user == "" {
		return "anonymous"
	}
	return user
}

// WithRemoteAddr returns a new context carrying the client remote address.
func WithRemoteAddr(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, remoteAddrKey{}, addr)
}

// RemoteAddrFromContext returns the remote address stored in ctx, or "".
func RemoteAddrFromContext(ctx context.Context) string {
	addr, _ := ctx.Value(remoteAddrKey{}).(string)
	return addr
}
