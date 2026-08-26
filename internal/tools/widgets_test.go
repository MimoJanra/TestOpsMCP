package tools

import (
	"strings"
	"testing"
)

// TestExtAppsBundleRewrite guards the vendored ext-apps bundle: if a future
// version bump (updating extAppsBundleRaw's embedded asset) ships a bundle
// whose ESM export shape rewriteESMExports no longer recognizes, this fails
// the build instead of silently shipping a broken widget to users.
func TestExtAppsBundleRewrite(t *testing.T) {
	rewritten := rewriteESMExports(extAppsBundleRaw)
	if rewritten == "" {
		t.Fatal("rewriteESMExports returned empty for the vendored ext-apps bundle; " +
			"the bundle's export statement shape likely changed and the regex needs updating")
	}
	if !strings.Contains(rewritten, "globalThis.ExtApps") {
		t.Fatalf("rewritten bundle does not define globalThis.ExtApps: %q", truncate(rewritten, 200))
	}
	if !strings.Contains(rewritten, "App:") {
		t.Errorf("rewritten bundle does not export App under the expected name; widgets construct `new App(...)`")
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
