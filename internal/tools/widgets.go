package tools

import (
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// ext-apps bundle fetching
// ---------------------------------------------------------------------------

const (
	launchDashboardURI       = "ui://widgets/launch-dashboard"
	launchDashboardURIPrefix = launchDashboardURI + "?launch_id="
	launchDashboardName      = "Launch Dashboard"
	widgetMimeType           = "text/html;profile=mcp-app"
)

func launchDashboardURIFor(launchID int64) string {
	return fmt.Sprintf("%s?launch_id=%d", launchDashboardURI, launchID)
}

func parseLaunchDashboardURI(uri string) (int64, bool) {
	if !strings.HasPrefix(uri, launchDashboardURIPrefix) {
		return 0, false
	}
	var id int64
	if _, err := fmt.Sscanf(uri[len(launchDashboardURIPrefix):], "%d", &id); err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// extAppsVersion pins the ext-apps package version vendored below. Floating
// on "latest" let upstream releases silently break the widget handshake twice
// before (see 43f4bc8, 0e258cc) — bump this deliberately (re-download the
// bundle, update assets/ext-apps-<version>.js and the go:embed directive)
// only after verifying the new version against the widget templates.
//
// Stay on the 1.7.x line: 1.6.0 was tried briefly to dodge a suspected
// zod-jitless bug in the App constructor, but its wire protocol turned out
// to be incompatible with the current Claude Desktop host (`i.parts is not
// iterable`). The jitless crash is instead worked around via the
// `allowUnsafeEval` App option passed from the widget templates, which
// skips the `zod.config({jitless:true})` call altogether.
const extAppsVersion = "1.7.4"

// extAppsBundleRaw is the ext-apps browser bundle, vendored at build time
// instead of fetched from a CDN at runtime. This removes a supply-chain
// dependency on unpkg/jsdelivr staying up and unmodified, and removes the
// failure mode where a single transient network error used to poison the
// process-lifetime bundle cache with a stripped-down fallback stub.
//
//go:embed assets/ext-apps-1.7.4.js
var extAppsBundleRaw string

var (
	bundleOnce sync.Once
	bundleJS   string
)

// getExtAppsBundle returns the vendored ext-apps browser bundle with ESM
// exports rewritten to globalThis.ExtApps. The rewrite is pure CPU work, so
// it's cached for the process lifetime purely to avoid repeating it per request.
func getExtAppsBundle(logger interface {
	Info(string, any)
	Warn(string, any)
}) string {
	bundleOnce.Do(func() {
		bundleJS = rewriteESMExports(extAppsBundleRaw)
		if bundleJS == "" {
			// Should be unreachable: TestExtAppsBundleRewrite fails the build
			// before this ships if the vendored bundle's export shape changes.
			if logger != nil {
				logger.Warn("vendored ext-apps bundle failed to rewrite ESM exports", map[string]any{
					"version": extAppsVersion,
				})
			}
			return
		}
		if logger != nil {
			logger.Info("ext-apps bundle loaded", map[string]any{
				"version": extAppsVersion,
				"size":    len(bundleJS),
			})
		}
	})
	return bundleJS
}

// rewriteESMExports converts trailing ESM export statements to globalThis.ExtApps = {...}.
// e.g. `export{Foo, Bar as Baz}` → `globalThis.ExtApps={Foo:Foo, Baz:Bar};`
func rewriteESMExports(js string) string {
	re := regexp.MustCompile(`export\s*\{([^}]+)\}\s*;?\s*$`)
	result := re.ReplaceAllStringFunc(js, func(match string) string {
		inner := re.FindStringSubmatch(match)[1]
		parts := strings.Split(inner, ",")
		props := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			halves := strings.SplitN(p, " as ", 2)
			if len(halves) == 2 {
				local := strings.TrimSpace(halves[0])
				exported := strings.TrimSpace(halves[1])
				props = append(props, exported+":"+local)
			} else {
				props = append(props, p+":"+p)
			}
		}
		return "globalThis.ExtApps={" + strings.Join(props, ",") + "};"
	})
	// Return empty string if no export was rewritten AND the bundle doesn't
	// already set globalThis.ExtApps (e.g. a CJS/UMD build).
	if result == js && !strings.Contains(js, "globalThis.ExtApps") {
		return ""
	}
	return result
}

// ---------------------------------------------------------------------------
// Widget HTML templates
// ---------------------------------------------------------------------------
//
// Vendored as separate .html files (instead of Go string constants) so they
// get real HTML/JS syntax highlighting and linting in editors. Each embeds
// the /*__EXT_APPS_BUNDLE__*/ marker that getExtAppsBundle's caller replaces
// with the vendored ext-apps bundle at serve time.

//go:embed assets/launch-dashboard.html
var launchDashboardTemplate string

//go:embed assets/action-picker.html
var actionPickerTemplate string

//go:embed assets/results-display.html
var resultsDisplayTemplate string

// ---------------------------------------------------------------------------
// Widget registration
// ---------------------------------------------------------------------------

const quickstartContent = `# TestOps MCP Server

Connect to Allure TestOps to manage test launches, results, and test cases directly from your AI assistant.

## Quick Start

1. Configure your Allure token via ` + "`configure_allure_token`" + ` or the ` + "`ALLURE_TOKEN`" + ` environment variable
2. Use ` + "`list_projects`" + ` to see available projects
3. Use ` + "`list_launches`" + ` to browse test runs
4. Use ` + "`get_launch_dashboard`" + ` for a visual launch overview with live stats

## Tool Groups

- **Launches** — list, get, create, close, reopen test launches
- **Results** — list test results, get details, update statuses
- **Test Cases** — search, create, update test cases and steps
- **Analytics** — trends, flaky tests, defect distribution
- **Bulk** — bulk status updates across test results and test cases
- **Search** — ` + "`search_testops_operations`" + ` discovers all 300+ API operations

## Prompts

Use the built-in prompts for common workflows:
- ` + "`analyze-test-failures`" + ` — deep-dive into failures in a specific launch
- ` + "`launch-report-summary`" + ` — generate an executive summary for a launch
`

// registerWidgets registers all MCP app tools and their associated resources.
// Called once from NewRegistry after all other tools are registered.
func (r *Registry) registerWidgets() {
	r.RegisterResource(&Resource{
		URI:      "allure://docs/quickstart",
		Name:     "TestOps MCP Quickstart",
		MimeType: "text/markdown",
		GetHTML:  func() string { return quickstartContent },
	})

	if r.allure == nil {
		return
	}

	// get_launch_dashboard — visual launch dashboard widget
	r.register(&Tool{
		Name: "get_launch_dashboard",
		Description: "Get an interactive visual dashboard for a launch showing real-time status, " +
			"progress bar, and pass/fail statistics. " +
			"Renders an inline widget in Claude Desktop and claude.ai. " +
			"Use instead of (or in addition to) get_launch_report for a richer view.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"launch_id": map[string]any{
					"type":        "integer",
					"description": "Allure launch ID",
				},
			},
			"required": []string{"launch_id"},
		},
		Meta: map[string]any{
			"ui": map[string]any{
				"resourceUri": launchDashboardURI,
			},
		},
		Handler: Typed(r.getLaunchDashboard),
	})

	// Register the widget resource — HTML is rendered lazily on first request
	// so the bundle fetch doesn't block server startup.
	r.RegisterResource(&Resource{
		URI:      launchDashboardURI,
		Name:     launchDashboardName,
		MimeType: widgetMimeType,
		GetHTML: func() string {
			bundle := getExtAppsBundle(r.logger)
			return strings.ReplaceAll(launchDashboardTemplate, "/*__EXT_APPS_BUNDLE__*/", bundle)
		},
	})

	// Register action picker widget resource
	r.RegisterResource(&Resource{
		URI:      "ui://widgets/action-picker",
		Name:     "Action Picker",
		MimeType: widgetMimeType,
		GetHTML: func() string {
			bundle := getExtAppsBundle(r.logger)
			return strings.ReplaceAll(actionPickerTemplate, "/*__EXT_APPS_BUNDLE__*/", bundle)
		},
	})

	// Register results display widget resource
	r.RegisterResource(&Resource{
		URI:      "ui://widgets/results-display",
		Name:     "Results Display",
		MimeType: widgetMimeType,
		GetHTML: func() string {
			bundle := getExtAppsBundle(r.logger)
			return strings.ReplaceAll(resultsDisplayTemplate, "/*__EXT_APPS_BUNDLE__*/", bundle)
		},
	})
}

// ---------------------------------------------------------------------------
// Resource watch / subscriptions
// ---------------------------------------------------------------------------

// watchLaunch polls the launch status every 10 s and calls publishResource when it changes.
// It runs until ctx is cancelled.
func (r *Registry) watchLaunch(ctx context.Context, launchID int64) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	var lastStatus string
	uri := launchDashboardURIFor(launchID)
	for {
		select {
		case <-ticker.C:
			if r.publishResource == nil || r.allure == nil {
				continue
			}
			details, err := r.allure.GetLaunchDetails(ctx, launchID)
			if err != nil {
				continue
			}
			currentStatus := fmt.Sprintf("%v", details.Status)
			if currentStatus != lastStatus {
				lastStatus = currentStatus
				r.publishResource(uri)
			}
		case <-ctx.Done():
			return
		}
	}
}

// StartLaunchWatch begins polling a launch and publishing updates to subscribers.
// Call this after a client subscribes to a launch dashboard resource.
func (r *Registry) StartLaunchWatch(ctx context.Context, launchID int64) {
	go r.watchLaunch(ctx, launchID)
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// normalizeLaunchStatus coerces Allure's raw launch status (a string, a
// {id,name} object, or null for launches with no status yet) into a plain
// string for the widget, which expects a string it can uppercase and badge.
func normalizeLaunchStatus(status any) string {
	switch v := status.(type) {
	case nil:
		return "UNKNOWN"
	case string:
		if v == "" {
			return "UNKNOWN"
		}
		return v
	case map[string]any:
		if name, ok := v["name"].(string); ok && name != "" {
			return name
		}
		return "UNKNOWN"
	default:
		return "UNKNOWN"
	}
}

type getLaunchDashboardArgs struct {
	LaunchID int64 `json:"launch_id"`
}

func (r *Registry) getLaunchDashboard(ctx context.Context, args getLaunchDashboardArgs) (any, error) {
	if args.LaunchID <= 0 {
		return nil, fmt.Errorf("launch_id must be positive")
	}

	r.logger.Info("fetching launch dashboard", map[string]any{"launch_id": args.LaunchID})

	details, err := r.allure.GetLaunchDetails(ctx, args.LaunchID)
	if err != nil {
		r.logger.Error("get launch details", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("get launch details: %w", err)
	}

	stats, err := r.allure.GetLaunchStatistics(ctx, args.LaunchID)
	if err != nil {
		r.logger.Error("get launch statistics", err, map[string]any{"launch_id": args.LaunchID})
		return nil, fmt.Errorf("get launch statistics: %w", err)
	}

	tags := make([]map[string]any, len(details.Tags))
	for i, tag := range details.Tags {
		tags[i] = map[string]any{"id": tag.ID, "name": tag.Name}
	}

	return map[string]any{
		"launch_id":      details.ID,
		"name":           details.Name,
		"status":         normalizeLaunchStatus(details.Status),
		"start_time":     details.StartTime,
		"end_time":       details.EndTime,
		"environment":    details.Environment,
		"tags":           tags,
		"report_web_url": details.ReportWebUrl,
		"stats": map[string]any{
			"total":  stats.Total,
			"passed": stats.Passed,
			"failed": stats.Failed,
			"broken": stats.Broken,
		},
	}, nil
}
