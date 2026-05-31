package session

import "context"

// ElicitResult is the response from an elicitation request.
type ElicitResult struct {
	Action  string // "accept" | "reject" | "cancel"
	Content []byte // JSON-encoded content (filled form)
}

// ElicitFunc asks the user to confirm or fill a form.
// message is the prompt text; schema is a JSON Schema for the form (may be nil).
// Returns the user's choice via ElicitResult.
type ElicitFunc func(ctx context.Context, message string, schema []byte) (*ElicitResult, error)

// SamplingResult is the response from a sampling (LLM) request.
type SamplingResult struct {
	Role       string
	Text       string
	StopReason string
}

// SamplingFunc requests LLM inference through the client.
type SamplingFunc func(ctx context.Context, system string, messages []SamplingMessage, maxTokens int) (*SamplingResult, error)

// SamplingMessage is a single message in a sampling request.
type SamplingMessage struct {
	Role string
	Text string
}

type samplingFuncKey struct{}

// WithSampling stores a SamplingFunc in the context.
func WithSampling(ctx context.Context, fn SamplingFunc) context.Context {
	return context.WithValue(ctx, samplingFuncKey{}, fn)
}

// SamplingFromContext retrieves the SamplingFunc from the context, if any.
func SamplingFromContext(ctx context.Context) (SamplingFunc, bool) {
	fn, ok := ctx.Value(samplingFuncKey{}).(SamplingFunc)
	return fn, ok
}

type elicitFuncKey struct{}

// WithElicit stores an ElicitFunc in the context.
func WithElicit(ctx context.Context, fn ElicitFunc) context.Context {
	return context.WithValue(ctx, elicitFuncKey{}, fn)
}

// ElicitFromContext retrieves the ElicitFunc from the context, if any.
func ElicitFromContext(ctx context.Context) (ElicitFunc, bool) {
	fn, ok := ctx.Value(elicitFuncKey{}).(ElicitFunc)
	return fn, ok
}

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
