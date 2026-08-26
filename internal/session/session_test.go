package session

import (
	"context"
	"errors"
	"testing"
)

func TestWithSampling(t *testing.T) {
	ctx := context.Background()
	if fn, ok := SamplingFromContext(ctx); ok || fn != nil {
		t.Fatalf("expected no SamplingFunc in empty context, got ok=%v fn=%v", ok, fn)
	}

	want := SamplingResult{Role: "assistant", Text: "hi", StopReason: "end_turn"}
	var fn SamplingFunc = func(ctx context.Context, system string, messages []SamplingMessage, maxTokens int) (*SamplingResult, error) {
		return &want, nil
	}
	ctx = WithSampling(ctx, fn)

	got, ok := SamplingFromContext(ctx)
	if !ok {
		t.Fatal("expected SamplingFunc present after WithSampling")
	}
	res, err := got(ctx, "sys", []SamplingMessage{{Role: "user", Text: "hello"}}, 100)
	if err != nil {
		t.Fatalf("unexpected error calling retrieved SamplingFunc: %v", err)
	}
	if *res != want {
		t.Errorf("got %+v, want %+v", *res, want)
	}
}

func TestWithSampling_PropagatesError(t *testing.T) {
	wantErr := errors.New("boom")
	var fn SamplingFunc = func(ctx context.Context, system string, messages []SamplingMessage, maxTokens int) (*SamplingResult, error) {
		return nil, wantErr
	}
	ctx := WithSampling(context.Background(), fn)

	got, _ := SamplingFromContext(ctx)
	res, err := got(ctx, "", nil, 0)
	if !errors.Is(err, wantErr) || res != nil {
		t.Errorf("got res=%v err=%v, want nil, %v", res, err, wantErr)
	}
}

func TestWithElicit(t *testing.T) {
	ctx := context.Background()
	if fn, ok := ElicitFromContext(ctx); ok || fn != nil {
		t.Fatalf("expected no ElicitFunc in empty context, got ok=%v fn=%v", ok, fn)
	}

	want := ElicitResult{Action: "accept", Content: []byte(`{"a":1}`)}
	var fn ElicitFunc = func(ctx context.Context, message string, schema []byte) (*ElicitResult, error) {
		return &want, nil
	}
	ctx = WithElicit(ctx, fn)

	got, ok := ElicitFromContext(ctx)
	if !ok {
		t.Fatal("expected ElicitFunc present after WithElicit")
	}
	res, err := got(ctx, "confirm?", []byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error calling retrieved ElicitFunc: %v", err)
	}
	if res.Action != want.Action || string(res.Content) != string(want.Content) {
		t.Errorf("got %+v, want %+v", *res, want)
	}
}

func TestWithID(t *testing.T) {
	if got := IDFromContext(context.Background()); got != "" {
		t.Errorf("IDFromContext on empty context = %q, want empty", got)
	}

	ctx := WithID(context.Background(), "sess-123")
	if got := IDFromContext(ctx); got != "sess-123" {
		t.Errorf("IDFromContext = %q, want %q", got, "sess-123")
	}
}

func TestWithID_StdioConstant(t *testing.T) {
	ctx := WithID(context.Background(), StdioID)
	if got := IDFromContext(ctx); got != "stdio" {
		t.Errorf("IDFromContext = %q, want %q", got, "stdio")
	}
}

func TestWithUser(t *testing.T) {
	if got := UserFromContext(context.Background()); got != "anonymous" {
		t.Errorf("UserFromContext on empty context = %q, want %q", got, "anonymous")
	}

	ctx := WithUser(context.Background(), "alice")
	if got := UserFromContext(ctx); got != "alice" {
		t.Errorf("UserFromContext = %q, want %q", got, "alice")
	}
}

func TestWithUser_EmptyStringFallsBackToAnonymous(t *testing.T) {
	ctx := WithUser(context.Background(), "")
	if got := UserFromContext(ctx); got != "anonymous" {
		t.Errorf("UserFromContext with empty user = %q, want %q", got, "anonymous")
	}
}

func TestWithRemoteAddr(t *testing.T) {
	if got := RemoteAddrFromContext(context.Background()); got != "" {
		t.Errorf("RemoteAddrFromContext on empty context = %q, want empty", got)
	}

	ctx := WithRemoteAddr(context.Background(), "127.0.0.1:1234")
	if got := RemoteAddrFromContext(ctx); got != "127.0.0.1:1234" {
		t.Errorf("RemoteAddrFromContext = %q, want %q", got, "127.0.0.1:1234")
	}
}

func TestContextValuesAreIndependent(t *testing.T) {
	ctx := context.Background()
	ctx = WithID(ctx, "sess-1")
	ctx = WithUser(ctx, "bob")
	ctx = WithRemoteAddr(ctx, "10.0.0.1:80")

	if got := IDFromContext(ctx); got != "sess-1" {
		t.Errorf("IDFromContext = %q, want %q", got, "sess-1")
	}
	if got := UserFromContext(ctx); got != "bob" {
		t.Errorf("UserFromContext = %q, want %q", got, "bob")
	}
	if got := RemoteAddrFromContext(ctx); got != "10.0.0.1:80" {
		t.Errorf("RemoteAddrFromContext = %q, want %q", got, "10.0.0.1:80")
	}
}
