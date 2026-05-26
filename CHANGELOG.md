# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.4.0...HEAD
[1.4.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.2.1...v1.3.0
[1.2.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.1.2...v1.2.0
[1.1.2]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/MimoJanra/TestOpsMCP/releases/tag/v1.0.0
