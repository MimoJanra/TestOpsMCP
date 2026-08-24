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

	var parseErr, versionErr, toolsList JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[0]), &parseErr); err != nil || parseErr.Error == nil || parseErr.Error.Code != ErrCodeParse {
		t.Errorf("line 1 = %q, want a parse-error response", lines[0])
	}
	if err := json.Unmarshal([]byte(lines[1]), &versionErr); err != nil || versionErr.Error == nil || versionErr.Error.Code != ErrCodeInvalidRequest {
		t.Errorf("line 2 = %q, want an invalid-request response", lines[1])
	}
	if err := json.Unmarshal([]byte(lines[2]), &toolsList); err != nil || toolsList.Error != nil {
		t.Errorf("line 3 = %q, want a successful tools/list response", lines[2])
	}
}
