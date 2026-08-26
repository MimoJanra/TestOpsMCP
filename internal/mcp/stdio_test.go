package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/core"
	"github.com/MimoJanra/TestOpsMCP/internal/tools"
)

// TestStdioHandler_Run redirects the process-wide os.Stdin/os.Stdout to pipes
// so the stdio transport (which reads/writes them directly) can be driven
// from a test. This mutates global state, so it must not run in parallel
// with other tests in this package — none currently call t.Parallel().
func TestStdioHandler_Run(t *testing.T) {
	origStdin, origStdout := os.Stdin, os.Stdout
	defer func() { os.Stdin = origStdin; os.Stdout = origStdout }()

	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdin = inR
	os.Stdout = outW

	logger := core.NewLogger(core.LevelError)
	registry := tools.NewRegistry(nil, logger)
	server := NewServer(registry, logger, Options{})
	sh := NewStdioHandler(server, logger)

	runDone := make(chan error, 1)
	go func() { runDone <- sh.Run() }()

	go func() {
		lines := []string{
			`not json`,
			`{"jsonrpc":"1.0","id":1,"method":"initialize"}`,
			`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
			`{"jsonrpc":"2.0","method":"notifications/initialized"}`, // notification: produces no output line
		}
		for _, l := range lines {
			fmt.Fprintln(inW, l)
		}
		inW.Close()
	}()

	var lines []string
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		scanner := bufio.NewScanner(outR)
		scanner.Buffer(make([]byte, 4096), 1<<20)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
	}()

	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("Run() returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return after stdin EOF")
	}
	outW.Close()

	select {
	case <-readDone:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out reading stdout")
	}

	if len(lines) != 3 {
		t.Fatalf("got %d output lines, want 3 (parse-error, invalid-version-error, tools/list result): %v", len(lines), lines)
	}

	// Each request is now dispatched on its own goroutine (see Run), so
	// responses are not guaranteed to arrive in request order — match each
	// expected response by its distinguishing property instead of position.
	var haveParseErr, haveVersionErr, haveToolsList bool
	for _, line := range lines {
		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			t.Fatalf("unmarshal line %q: %v", line, err)
		}
		switch {
		case resp.Error != nil && resp.Error.Code == ErrCodeParse:
			haveParseErr = true
		case resp.Error != nil && resp.Error.Code == ErrCodeInvalidRequest:
			haveVersionErr = true
		case resp.Error == nil && string(resp.ID) == "2":
			haveToolsList = true
		}
	}
	if !haveParseErr {
		t.Errorf("no parse-error response found in %v", lines)
	}
	if !haveVersionErr {
		t.Errorf("no invalid-request response found in %v", lines)
	}
	if !haveToolsList {
		t.Errorf("no successful tools/list response (id=2) found in %v", lines)
	}
}

// TestStdioHandler_Elicitation drives a full round trip through the fix for
// "no interactive session is available" over stdio: a tool call that needs
// user confirmation must (1) emit a server-initiated elicitation/create
// request on stdout without blocking the read loop, and (2) resume once the
// matching JSON-RPC response arrives on a later stdin line.
func TestStdioHandler_Elicitation(t *testing.T) {
	origStdin, origStdout := os.Stdin, os.Stdout
	defer func() { os.Stdin = origStdin; os.Stdout = origStdout }()

	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdin = inR
	os.Stdout = outW

	logger := core.NewLogger(core.LevelError)
	registry := tools.NewRegistry(nil, logger)
	server := NewServer(registry, logger, Options{})
	sh := NewStdioHandler(server, logger)

	runDone := make(chan error, 1)
	go func() { runDone <- sh.Run() }()

	scanner := bufio.NewScanner(outR)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	readLine := func() JSONRPCRequest {
		if !scanner.Scan() {
			t.Fatalf("stdout closed early: %v", scanner.Err())
		}
		var msg JSONRPCRequest
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			t.Fatalf("unmarshal %q: %v", scanner.Text(), err)
		}
		return msg
	}

	fmt.Fprintln(inW, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"delete_test_case","arguments":{"test_case_id":42}}}`)

	// First line out must be the server-initiated confirmation request, not
	// the tools/call response — proving the read loop wasn't blocked waiting
	// on it and elicitation actually reached the client.
	elicitReq := readLine()
	if elicitReq.Method != "elicitation/create" {
		t.Fatalf("first stdout message method = %q, want elicitation/create", elicitReq.Method)
	}
	var elicitID string
	if err := json.Unmarshal(elicitReq.ID, &elicitID); err != nil {
		t.Fatalf("elicitation request id: %v", err)
	}

	// Client rejects the confirmation.
	fmt.Fprintf(inW, "{\"jsonrpc\":\"2.0\",\"id\":%q,\"result\":{\"action\":\"reject\"}}\n", elicitID)

	toolResp := readLine()
	if string(toolResp.ID) != "1" {
		t.Fatalf("expected the tools/call response (id=1) next, got %+v", toolResp)
	}

	inW.Close()
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("Run() returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return after stdin EOF")
	}
	outW.Close()
}
