# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.1] - 2026-06-01 - Security & Health Check Fixes

### Added
- `GET /health` — unauthenticated liveness probe endpoint; returns `200 ok`. Used by Docker `HEALTHCHECK`, Kubernetes liveness/readiness probes, and uptime monitors. Does not expose any server state.

### Fixed
- **Audit log bypass** — notification-style JSON-RPC requests (no `id` field) were silently skipped by the audit middleware. A client could send `tools/call` without an `id` and the tool would execute without being logged. All requests are now audited regardless of notification status; notification entries get `status: "notification"`.
- **Docker `HEALTHCHECK`** — was calling unauthenticated `GET /sse`, which returns `401` when `MCP_AUTH_TOKENS` is configured. Now calls `GET /health`.
- **CORS wildcard default** — `CORS_ALLOWED_ORIGIN` now defaults to `""` (disabled) instead of `"*"`. Operators must explicitly set the origin. A startup INFO log explains the setting and suggests `https://claude.ai` for browser-based access. Claude Desktop and CLI tools (mcp-remote) are unaffected — they do not enforce browser CORS policy.
- **`/health` method restriction** — previously accepted any HTTP method and returned `200`. Now only `GET` and `HEAD` are accepted; other methods return `405 Method Not Allowed`.

### Changed
- All documentation updated to prefer `MCP_AUTH_TOKENS=name:token,...` over the legacy `MCP_AUTH_TOKEN` single-token variable in shared server examples.
- Kubernetes liveness/readiness probes in docs updated to use `/health` instead of `/sse` or `/messages`.

## [2.0.0] - 2026-05-31 - MCP Protocol Improvements

### Added

#### Generic Typed Handlers
- All 104+ tool handlers now use a `Typed[T]` generic wrapper that auto-deserializes JSON input — eliminates ~200 manual `json.Unmarshal` calls.
- Handler signatures changed from `(ctx, json.RawMessage)` to `(ctx, TypedArgs)` — cleaner and type-safe.

#### slog Integration
- `internal/core.Logger` now wraps `log/slog` with `slog.NewJSONHandler`. Same public API, standard Go logging internals.

#### Middleware Chain + Panic Recovery
- New `internal/mcp/middleware.go` with composable `middlewareFunc` chain.
- **Panic recovery middleware** (first in chain): catches handler panics, logs stack trace, returns `{code: -32603}` instead of crashing the server.
- Audit logging extracted to its own middleware; `dispatch()` simplified to a single call.

#### MCP Protocol 2025-11-25
- `ProtocolVersion` bumped to `"2025-11-25"`.
- Version negotiation in `initialize`: server accepts client's requested version if in supported list (2024-11-05, 2025-03-26, 2025-06-18, 2025-11-25).
- New `logging` and `elicitation` capability fields in `ServerCapabilities`.

#### Cursor-Based Pagination
- `tools/list`, `resources/list`, `prompts/list` now support cursor-based pagination (page size 50).
- Response includes `nextCursor` when more pages exist. Tools list sorted alphabetically for stable cursors.

#### Async Task System
- New `internal/tasks` package: `Task` struct with status lifecycle (working/succeeded/failed/cancelled), in-memory `Store` with context-based cancellation.
- Three new tools: **`get_task_status`**, **`list_running_tasks`**, **`cancel_task`**.
- Long-running operations now return `{task_id, message}` immediately and complete in the background: `run_allure_launch`, `copy_launch`, `merge_launches`, `bulk_run_test_cases_new_launch`, `bulk_run_test_cases_existing_launch`, `bulk_clone_test_cases`.

#### Completion (Argument Autocompletion)
- New `completion/complete` JSON-RPC method.
- Registry has a `Complete(promptName, argName, partial)` method — returns up to 10 suggestions.
- `project_id` arguments complete against live Allure project list when API is configured.

#### Elicitation (Confirmation Dialogs)
- Server can ask the user to confirm destructive operations via `elicitation/create` notification.
- `session.ElicitFunc` context key lets handlers call into the elicitation round-trip.
- Applied to: **`delete_test_case`** and **`bulk_delete_test_cases`** — user must accept before deletion proceeds.
- Route: `notifications/elicitation/complete` delivers user's answer back to the server.

#### Resource Subscriptions
- `resources/subscribe` and `resources/unsubscribe` JSON-RPC methods are now handled.
- `Server.PublishResource(uri)` sends `notifications/resources/updated` to all subscribed sessions.
- Launch dashboard widget auto-starts a 10-second polling watcher when subscribed via `ui://widgets/launch-dashboard?launch_id=N`.
- Server advertises `resources.subscribe = true` in capabilities.

#### Sampling (Server → LLM via Client)
- Server can request LLM inference through the client via `sampling/createMessage`.
- `session.SamplingFunc` context key exposes this to handlers.
- New **`analyze_launch_failures`** tool: fetches failed test results and asks Claude to identify root causes and suggest fixes. Gracefully degrades if sampling is unavailable.

### Fixed

#### Async Task System
- **Panic recovery in goroutines** — `tasks.Store.Run()` wraps every background goroutine in `recover()`. A panic marks the task `Failed` and logs the stack trace; the server and other sessions remain unaffected.
- **Session token lost in async context** — `Store.Create()` now accepts `parentCtx` and uses `context.WithoutCancel` to propagate session ID (and all other context values) without inheriting the request's cancellation signal. Async operations (`run_allure_launch`, `copy_launch`, `merge_launches`, `bulk_run_*`, `bulk_clone_test_cases`) now correctly resolve per-user Allure tokens in multi-user deployments.
- **Data race in task reads** — `Store.Get()` and `Store.List()` now return struct copies under `RLock` instead of raw pointers; `taskToMap` no longer races against concurrent `Update` writes.
- **Async task timeout** — `taskCtx` carries a 30-minute hard deadline via `context.WithTimeout`; hung Allure calls can no longer leak goroutines and memory indefinitely.
- **Task store memory leak** — Added background janitor (`StartJanitor`) that purges succeeded/failed/cancelled tasks older than 1 hour; runs every 5 minutes.

#### MCP Protocol — Elicitation & Sampling
- **Standard JSON-RPC response routing** — `elicitation/create` is now sent as a proper JSON-RPC *request* (with `id`); `sampling/createMessage` already had an `id`. Both now receive the client's standard JSON-RPC response `{id, result}` routed through a new `handleJSONRPCResponse` dispatcher, keyed by the request ID. The previous non-standard methods `notifications/elicitation/complete` and `sampling/createMessage/response` are removed. Compatible with Claude Desktop and any spec-compliant MCP client.
- **Silent deletion on stdio** — `delete_test_case` and `bulk_delete_test_cases` on the stdio transport previously bypassed the confirmation guard (no `ElicitFunc` in context → the `if ok` block was skipped, and data was deleted silently). Now they return an error: *"deletion requires user confirmation but no interactive session is available"*. Also, transport errors from `elicit()` are now returned as errors instead of being masked as a user cancellation.

#### Resource Subscriptions
- **Push notifications never arrived** — `handleResourcesSubscribe` was passing the HTTP request context to `OnSubscribe → StartLaunchWatch → watchLaunch`. The request's `r.Context()` was cancelled when the `POST /messages` handler returned (202 Accepted), so the watcher exited before the first 10-second tick. Now `sess.ctx` (session lifetime) is used — the watcher runs until the client disconnects.

#### Other
- **Unstable pagination** — `resources/list` and `prompts/list` now sort by URI and name respectively before slicing, matching `tools/list` behaviour. Previously, cursor-based pagination over Go maps produced non-deterministic results.
- **Completion hits Allure on every keystroke** — `Complete` now maintains a 30-second in-process cache of project/launch ID lists fetched with a 5-second timeout derived from the request context. Completion is gated on `ref.type == "ref/prompt"` to avoid unintended Allure calls from resources with identically-named arguments.
- **UTF-8 truncation in analysis** — `analyzeLaunchFailures` truncated error messages and stack traces by byte index (`msg[:300]`), which split multi-byte runes (Cyrillic, CJK, emoji). Now uses `truncateRunes` which truncates at rune boundary.

### Changed
- Tool count: 100 → 104 (added `get_task_status`, `list_running_tasks`, `cancel_task`, `analyze_launch_failures`).
- `initialize` response now includes `logging` capability.
- `resources.subscribe` is now `true` in `initialize` capabilities.

## [1.9.0] - 2026-05-31 - MCP Prompts & Resources

### Added

#### MCP Prompts
- **`analyze-test-failures`** — prompt template that instructs Claude to retrieve launch details, list failed/broken results, group by error pattern, and surface critical issues. Argument: `launch_id` (required), `project_id` (optional).
- **`launch-report-summary`** — prompt template that instructs Claude to generate a concise executive summary (pass rate, counts, duration, environment, top failures). Argument: `launch_id` (required), `project_id` (optional).
- Server now advertises `prompts` capability in the MCP `initialize` response.
- `prompts/list` and `prompts/get` JSON-RPC methods are fully handled.

#### MCP Resources
- **`allure://docs/quickstart`** — static Markdown resource listing tool groups, prompt templates, and quick-start steps. Available even when no Allure token is configured; clients can attach it as context.

## [1.8.1] - 2026-05-30 - Doc Fixes

### Fixed
- **Claude Desktop client config** — replaced `url` + `headers` format (not supported by Claude Desktop for custom headers) with `mcp-remote` proxy pattern. All documentation examples now show `npx mcp-remote <url> --header Key:Value` for both Windows (`cmd /c npx ...`) and macOS/Linux (`npx ...`). Updated README, DEPLOYMENT.md, QUICKSTART.md, llms.txt, llms-full.txt.

## [1.8.0] - 2026-05-30 - Multi-User Auth & Audit Log

### Added

#### Multi-User Authentication
- **`MCP_AUTH_TOKENS`** environment variable — configure named user tokens in `name:token,...` format (e.g. `alice:abc123,bob:xyz789`). Each user authenticates with their own bearer token; requests are attributed to the user by name in logs and audit records.
- **Backward-compatible** — existing `MCP_AUTH_TOKEN` (single token) still works and is treated as user `"default"`.
- Startup log now shows the configured user count instead of a boolean auth flag.

#### Audit Log
- **Daily JSONL audit files** written to a configurable directory (`AUDIT_LOG_PATH`, default `audit/`).
- Each entry records: `timestamp`, `user`, `session_id`, `remote_addr`, `method`, `tool` (for `tools/call`), `status` (`ok`/`error`), `duration_ms`.
- **Automatic retention** — files older than `AUDIT_RETENTION_DAYS` days (default 30) are deleted nightly.
- Docker Compose mounts `./audit` as a host volume so logs survive container restarts.
- Audit logger is disabled gracefully (warning logged) if the directory cannot be created.

#### New Environment Variables
| Variable | Default | Description |
|---|---|---|
| `MCP_AUTH_TOKENS` | — | Named user tokens: `alice:tok1,bob:tok2` |
| `AUDIT_LOG_PATH` | `audit` | Directory for daily audit JSONL files |
| `AUDIT_RETENTION_DAYS` | `30` | Days to keep audit files |

### Fixed

#### GitHub Actions — Docker Release
- Added **QEMU** setup step (`docker/setup-qemu-action@v3`) required for multi-platform (`linux/amd64`, `linux/arm64`) builds — previously the multi-arch build was silently skipped.
- Added `platforms: linux/amd64,linux/arm64` to `build-push-action`.
- Added `type=raw,value=latest` tag so the `:latest` image is correctly pushed on tag releases.
- Fixed `publish-release` step using `--notes-append` instead of `--notes-file` (which was overwriting the full release body).
- Pinned action versions to stable releases (`actions/checkout@v4`, `docker/setup-buildx-action@v3`, `docker/login-action@v3`, `docker/metadata-action@v5`).

### Changed
- `docker-compose.yml` — removed duplicated `environment` keys that were already covered by `env_file`; added `AUDIT_LOG_PATH` and `AUDIT_RETENTION_DAYS` with defaults; added `./audit:/app/audit` volume mount.

## [1.7.1] - 2026-05-29 - MCP Compliance Fixes & Model Expansions

### Fixed

#### MCP Protocol Compliance
- **Tool annotations missing on 3 tools** — `configure_allure_token`, `search_testops_operations`, and `execute_testops_operation` were registered after the annotation loop and sent `null` annotations to Claude. All three now receive correct `readOnlyHint`/`destructiveHint` values.
- **`execute_testops_operation` now marked `destructiveHint: true`** — the tool can execute any API operation including DELETE/PUT, which requires the destructive hint per MCP spec.
- **CORS headers missing custom headers** — `Mcp-Session-Id` and `X-Allure-Token` added to `Access-Control-Allow-Headers`; `DELETE` added to `Access-Control-Allow-Methods`. Previously, browsers would block cross-origin preflight requests using these headers.
- **Streamable HTTP Content-Type validation** — POST `/mcp` now rejects requests with a non-JSON `Content-Type` with `415 Unsupported Media Type`, as required by MCP spec 2025-03-26.
- **`GetLaunchStatistics` response parsing** — the Allure API returns a `[]{"status", "count"}` array, not a single object. Client now aggregates the array into `StatisticsResponse` correctly; the Launch Dashboard widget now shows real pass/fail/broken counts.

#### Data Models
- **`ClientCapabilities`** — `InitializeRequest.Capabilities` was an empty `struct{}`, silently discarding client capability flags. Replaced with typed struct parsing `elicitation`, `sampling`, and `roots` fields — required for future elicitation support.
- **`LaunchCreateRequest`** expanded with `AutoClose`, `External`, `Issues`, `Links`, `Tags` fields matching the full API schema.
- **`LaunchResponse`** enriched with `UUID`, `CreatedDate`, `LastModifiedDate`, `AutoClose`, `Closed`, `External`, `Issues`, `Links`, `Tags`.
- **`StatisticsResponse`** expanded with `Unknown` field; new `StatisticItem` DTO for the raw array items from the statistics endpoint.

### Added

#### New Model DTOs
- `CategoryDto` — test result category reference
- `TestLayerDto` — test layer reference  
- `JobRunDto` — CI/CD job run linked to a test result (with URL)
- `IdAndNameOnlyDto` — lightweight id+name reference
- `IntegrationTypeDto` — issue tracker integration type
- `RoleDto` — user role in a project
- `StatusDto` — named status object for test cases
- `WorkflowRowDto` — workflow attached to a test case
- `CustomFieldValueWithCfDto` — payload for setting custom field values via the `/cfv` endpoint

### Notes
- No tool count changes — all fixes are internal correctness and protocol compliance improvements
- Fully backwards-compatible: existing Claude Desktop and claude.ai configurations require no changes

## [1.7.0] - 2026-05-27 - Interactive API Discovery Widgets

### Added

#### MCP App Widgets for Search/Execute Tools
- **Action Picker Widget** — interactive searchable picker for `search_testops_operations` results
  - Real-time filtering by operation name, description, or API path
  - HTTP method color-coding (GET green, POST blue, PUT orange, DELETE red)
  - Click to select operation and inject operation_id into chat
- **Results Display Widget** — formatted result viewer for `execute_testops_operation` 
  - Status indicator (Success/Error) with color
  - Auto-formatted JSON with proper indentation
  - Scrollable body for large API responses
  - Graceful fallback for raw text responses

### Changed
- **`search_testops_operations`** — updated to render Action Picker widget; description now mentions widget support
- **`execute_testops_operation`** — updated to render Results Display widget; description now mentions widget support

### Notes
- All widgets support light/dark mode via host theme detection
- Compatible with Claude Desktop and claude.ai
- Uses ext-apps bundle for cross-platform rendering

## [1.6.0] - 2026-05-26 - Full Test Case API Coverage

### Added

#### Test case single-item operations (25 new tools)
- **`get_test_case_tags`** / **`set_test_case_tags`** — read and replace tags on a test case
- **`get_test_case_issues`** / **`set_test_case_issues`** — read and replace linked bug-tracker issues
- **`get_test_case_examples`** / **`set_test_case_examples`** — parametrized data table rows (full replace)
- **`list_test_case_versions`** / **`create_test_case_version`** / **`restore_test_case_version`** — version snapshot management
- **`get_test_case_version_data`** / **`delete_test_case_version`** — version content and cleanup
- **`get_test_case_attachments`** / **`delete_test_case_attachment`** — attachment listing and removal
- **`search_test_cases`** — AQL/RQL full-text search across project test cases
- **`list_deleted_test_cases`** — browse soft-deleted test cases
- **`list_muted_test_cases`** — browse muted test cases
- **`delete_test_case_scenario`** — remove the entire step scenario from a test case
- **`move_test_case_step`** / **`copy_test_case_step`** — reposition steps within a scenario
- **`get_test_case_relations`** / **`set_test_case_relations`** — test-case-to-test-case relations
- **`get_test_case_custom_fields`** / **`update_test_case_custom_fields`** — custom field values via dedicated `/cfv` endpoint
- **`get_test_case_workflow`** — workflow definition for a test case
- **`get_test_case_keys`** / **`set_test_case_keys`** — integration test keys (Jira, Azure DevOps, etc.)
- **`get_test_case_scenario_from_run`** — scenario captured from the last automated run
- **`detach_test_case_automation`** — convert an automated test case back to manual
- **`get_test_case_audit`** — change history / audit log
- **`validate_test_case_query`** — validate AQL/RQL expression without executing it
- **`suggest_test_cases`** — autocomplete suggestions by name

#### Bulk test-case operations (14 new tools)
- **`bulk_add_test_case_members`** / **`bulk_remove_test_case_members`**
- **`bulk_add_test_case_custom_fields`** / **`bulk_remove_test_case_custom_fields`**
- **`bulk_add_test_case_external_links`**
- **`bulk_add_test_case_issues`** / **`bulk_remove_test_case_issues`**
- **`bulk_set_test_case_layer`**
- **`bulk_move_test_cases`** — move test cases to another project
- **`bulk_delete_test_cases`** — permanent bulk delete
- **`bulk_run_test_cases_new_launch`** / **`bulk_run_test_cases_existing_launch`**
- **`bulk_create_test_plan`** — create a test plan from selected test cases
- **`bulk_mute_test_cases`**

### Fixed
- **`execute_testops_operation`** parameter routing — path parameters no longer leaked into the request body; unknown parameters are now collected separately
- **`execute_testops_operation` array bodies** — new explicit `body` key lets callers pass any JSON value (array or object) directly as the HTTP request body

### Changed
- Total MCP tools: **57 → 102** (45 new tools, all backed by proper client methods — no OpenAPI dynamic execution fallback)
- All new client methods added to `internal/adapters/allure/client.go` with typed request/response models in `models.go`
- New file `internal/tools/tools_testcases_extra.go` for the extended single-item test case tools

## [1.5.0] - Security & Architecture Hardening

### Added
- **Per-session Allure token isolation** — Each SSE session stores its own token; concurrent users on a shared server never mix credentials
- **`X-Allure-Token` header support** — Users pass their personal Allure token from Claude Desktop config via HTTP header; no need to call `configure_allure_token` manually
- **Token-keyed JWT cache** — JWT tokens cached per API key, not globally; separate users get separate JWTs
- **`internal/session` package** — Shared context helpers for session ID propagation across packages
- **Server startup warning** — Logged when HTTP mode runs without `MCP_AUTH_TOKEN`
- **Build-time version injection** — Server version set via `-ldflags` instead of hardcoded `"1.0.0"`

### Changed
- **`registry.go` split into domain files** — `tools_launches.go`, `tools_results.go`, `tools_testcases.go`, `tools_projects.go`, `tools_analytics.go`, `tools_bulk.go`, `tools_relations.go`
- **`GetSessionToken` callback now context-aware** — Receives `context.Context` to resolve the correct per-session token
- **Dockerfile** — Pinned `alpine:3.21`, removed `.exe` suffix from Linux binary
- **HTTP server timeouts** — Added `ReadTimeout` and `IdleTimeout`; `WriteTimeout` disabled for SSE long-lived streams
- **Documentation** — All Claude Desktop config examples corrected to use `url` + `headers` for remote servers; `Authorization` marked optional

### Fixed
- **Data race on JWT cache** — Added `sync.Mutex` protecting `jwtToken`/`jwtExpiresAt` fields
- **Cross-user token contamination** — Global `sessionToken` field replaced with per-session map
- **Invalid JSON on marshal error** — `resultToJSON` error path now uses `json.Marshal` instead of `fmt.Sprintf`
- **Predictable session ID fallback** — `crypto/rand` failure now panics instead of using `time.Now().UnixNano()`
- **`SetSessionTokenFunc` visibility** — Exported function field replaced with proper setter method

## [1.4.0] - Token Priority & Shared Server Auth

### Added
- **Improved token handling for shared servers** — Users can now provide personal tokens on shared server deployments
- **Session token priority** — Personal tokens override server tokens, enabling per-user authentication on shared setups

### Changed
- **ALLURE_TOKEN handling** — Server token is now truly optional (session token takes priority)
- **README reorganized** — Cleaner quick start for end users, separated setup guides for per-user and shared server setups
- **.env.example updated** — Clear documentation of per-user vs shared server setup modes with helpful comments

### Fixed
- Token priority logic in `client.go` — session/user tokens now correctly override fallback server tokens
- Documentation clarity on shared server setup without mandatory ALLURE_TOKEN

## [1.3.0] - 2026-05-12

### Added
- **Search + Execute tools** — Query and execute any of 600+ TestOps API endpoints dynamically
- **Per-user authentication** — Users can authenticate with their own tokens in chat, no shared credentials needed
- **Session-based token configuration** — New `configure_allure_token` tool for chat-based auth setup with security warnings
- **Go 1.26 support** — Updated all tooling, CI/CD, and Docker builds to Go 1.26

### Changed
- **ALLURE_TOKEN is now optional** — Shared server deployments no longer require pre-configured tokens
- **OpenAPI spec bundled** — Complete Allure TestOps API specification included in repo (spec/testops.json)
- **Improved documentation** — Added per-user vs shared server setup guides, lazy setup instructions

### Improved
- Authentication flexibility for different deployment scenarios
- API coverage expanded from 55 explicit tools to 55 + dynamic search/execute for 600+ operations
- Development experience with updated Go dependencies across the board

## [1.2.1] - 2026-05-08

### Added
- **GitHub Actions CI/CD pipeline** — Automated multi-platform builds for Windows, macOS, and Linux
- **CLAUDE.md** — Developer documentation with release procedures and project guidelines

### Changed
- **Simplified README** — Local Claude Desktop setup now the primary onboarding flow
- **Improved getting-started experience** — 3-minute setup instead of 10+ minutes
- **Marketing-focused documentation** — Value-driven introduction, use cases, and comparisons
- **Better documentation structure** — Clearer navigation and links throughout

### Improved
- README with benefit-driven messaging
- Deployment option clarity with examples
- Contributing guidelines
- SEO optimization for discoverability
- Build artifact management (.gitignore updates)

### Infrastructure
- Automated release process: `git tag vX.Y.Z && git push origin vX.Y.Z`
- Pre-built binaries attached automatically to GitHub releases
- CI/CD pipeline ready for future releases

## [1.2.0] - 2026-05-08

### Added
- **21 new MCP tools** for comprehensive test case management
- Complete defect management: `add_test_case_defect`, `remove_test_case_defect`, `get_test_case_defects`, `get_launch_defects`
- Team collaboration: `get_test_case_members`, `add_test_case_members`, `remove_test_case_members`
- External links: `get_test_case_external_links`, `add_test_case_external_link`, `delete_test_case_external_link`
- Enhanced test case operations: `clone_test_case`, `restore_test_case`, `get_test_case_history`
- Advanced launch operations: `copy_launch`, `merge_launches`
- Environment management: `get_launch_environment`, `update_launch_environment`
- Bulk operations: `bulk_clone_test_cases`
- Single test result operations: `resolve_test_result`, `unmute_test_result`
- Manual scenario update support via `update_test_case` tool

### Changed
- Tool count: 35 → 55 (added 20 new tools)
- Release distribution: pre-built binaries for Windows, macOS (Intel & ARM), and Linux
- Updated README with quick start for pre-built binaries

### Added Documentation
- New `RELEASES.md` file with detailed setup instructions for all platforms
- Platform-specific guides: Windows, macOS, Linux
- Troubleshooting section for common issues
- API token retrieval instructions

### Features
- ✅ Support for `manual_scenario` update in test case updates
- ✅ Comprehensive test case lifecycle management
- ✅ Full defect tracking integration
- ✅ Team member assignment and management
- ✅ External link/relation support (GitHub, Jira, etc.)
- ✅ Launch environment variable management
- ✅ Test case cloning and restoration
- ✅ Launch merging capabilities

## [1.1.2] - 2026-05-08

### Fixed
- `get_test_case` now includes `manual_scenario` field with complete test execution steps
- Fetches scenario via dedicated `/api/testcase/{id}/scenario` endpoint
- Returns scenario attachments, step structure, and expected results

### Changed
- Tool count: 32 → 35 (added 3 step management tools)
- Tool description clarified to explicitly mention `manual_scenario` field
- README updated with note about accessing test execution steps

### Documentation
- Added explicit note that `get_test_case` includes all test execution steps in `manual_scenario` field
- Updated tools table to include step management tools (`create_test_case_step`, `update_test_case_step`, `delete_test_case_step`)

## [1.1.1] - 2026-05-08

### Fixed
- `get_test_case` now returns complete test case overview from `/api/testcase/{id}/overview` endpoint
- Added execution steps (scenario), tags, members, custom fields, and all metadata to test case details
- Resolved missing Execution section that was not available in basic endpoint

## [1.1.0] - 2026-05-08

### Added
- **Full test case field editing** — Extended `update_test_case` tool to support all fields from TestCasePatchV2Dto:
  - Text fields: `description`, `precondition`, `expected_result`, `full_name`
  - Boolean flags: `automated`, `external`, `deleted`
  - Resource IDs: `status_id`, `test_layer_id`, `workflow_id`
  - Collections: `tags` (array of tag objects), `members` (array of member objects), `links` (array of external links)
- **Step management tools** — Full CRUD operations for test case steps:
  - `create_test_case_step` — Create a new step with optional positioning (`after_id`) and nesting (`parent_id`)
  - `update_test_case_step` — Update step body and expected results
  - `delete_test_case_step` — Delete a step from a test case
- **Extended models** — New DTOs for step operations and external links support

### Changed
- `update_test_case` handler refactored to accept struct-based request instead of individual parameters
- Tool count increased from 32 to 35

## [1.0.0] - 2026-04-23

### Added
- **32 MCP tools** covering the full Allure TestOps workflow:
  - Launch management: `run_allure_launch`, `get_launch_status`, `get_launch_report`, `list_launches`, `get_launch_details`, `close_launch`, `reopen_launch`, `add_test_cases_to_launch`, `add_test_plan_to_launch`
  - Test results: `list_test_results`, `get_test_result`, `assign_test_result`, `mute_test_result`, `bulk_assign_test_results`, `bulk_mute_test_results`, `bulk_unmute_test_results`, `bulk_resolve_test_results`
  - Test cases: `list_test_cases`, `get_test_case`, `create_test_case`, `update_test_case`, `delete_test_case`, `run_test_case`, `bulk_set_test_case_status`, `bulk_add_test_case_tags`, `bulk_remove_test_case_tags`
  - Projects & analytics: `list_projects`, `get_project`, `get_project_stats`, `get_launch_trend_analytics`, `get_launch_duration_analytics`, `get_test_success_rate`
- **Dual transport modes**: stdio (Claude Desktop) and HTTP/SSE (team deployment)
- **Bearer-token authentication** via `MCP_AUTH_TOKEN` for HTTP mode
- **CORS support** with configurable `CORS_ALLOWED_ORIGIN`
- **Structured JSON logging** to stderr with configurable level (`LOG_LEVEL`)
- **Multi-stage Dockerfile** with non-root user and health check
- **Docker Compose** configuration for team deployment with resource limits
- **Kubernetes manifest** (`k8s-manifest.yaml`) with deployment, service, and resource constraints
- **Caddy reverse proxy** config for automatic HTTPS
- **Systemd service** example for Linux deployments
- **MCP protocol 2024-11-05** with full JSON-RPC 2.0 compliance
- Comprehensive documentation: Installation, Deployment, API Reference, Security guides
- `.env.example` configuration template

[Unreleased]: https://github.com/MimoJanra/TestOpsMCP/compare/v2.0.0...HEAD
[2.0.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.9.0...v2.0.0
[1.9.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.8.1...v1.9.0
[1.8.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.8.0...v1.8.1
[1.8.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.7.1...v1.8.0
[1.7.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.7.0...v1.7.1
[1.7.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.2.1...v1.3.0
[1.2.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.1.2...v1.2.0
[1.1.2]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/MimoJanra/TestOpsMCP/releases/tag/v1.0.0
