package tools

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
)

// repoRoot returns the repository root, computed from this file's own path
// rather than the process's working directory (which `go test` sets to the
// package directory, not the repo root).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/tools/docs_sync_test.go -> repo root is two levels up.
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

// fullyConfiguredToolCount builds a Registry the way the real server does
// (non-nil Allure client, real OpenAPI spec) and returns len(ListTools()).
// FindSpecFile only checks a handful of relative paths, none of which reach
// <repo>/spec/testops.json from the internal/tools package directory that
// `go test` uses as its working directory — so this temporarily chdirs to
// the repo root, matching how the compiled binary is actually run.
func fullyConfiguredToolCount(t *testing.T) int {
	t.Helper()
	root := repoRoot(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to repo root %q: %v", root, err)
	}
	defer func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	client := allure.NewClient("https://example-allure.invalid", "", 5*time.Second)
	r := NewRegistry(client, core.NewLogger(core.LevelError))

	if r.opIndex == nil {
		t.Fatal("opIndex is nil — spec/testops.json was not found from the repo root; " +
			"this test can't establish the real fully-configured tool count")
	}
	return len(r.ListTools())
}

// TestDocsToolCountMatchesRegistry guards against the tool count drifting out
// of sync between the code and the docs that quote it (README.md, docs/API.md,
// llms.txt) — this drift previously went unnoticed for multiple releases
// (104 in some docs, 114 in others, both wrong: the real number is whatever
// this test measures).
func TestDocsToolCountMatchesRegistry(t *testing.T) {
	want := fullyConfiguredToolCount(t)
	root := repoRoot(t)

	re := regexp.MustCompile(`(\d+)\s*(?:curated\s+)?[Tt]ools\b`)

	cases := []struct {
		file      string
		allowMiss bool // file may not mention a count at all
	}{
		{"README.md", false},
		{filepath.Join("docs", "API.md"), false},
		{"llms.txt", false},
	}

	for _, tc := range cases {
		path := filepath.Join(root, tc.file)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", tc.file, err)
		}
		matches := re.FindAllStringSubmatch(string(data), -1)
		if len(matches) == 0 {
			if tc.allowMiss {
				continue
			}
			t.Errorf("%s: no tool-count mention found matching %q", tc.file, re.String())
			continue
		}
		for _, m := range matches {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				continue
			}
			if n != want {
				t.Errorf("%s: mentions %d tools, want %d (from a fully-configured registry)", tc.file, n, want)
			}
		}
	}
}
