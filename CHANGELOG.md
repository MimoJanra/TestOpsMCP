# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/MimoJanra/TestOpsMCP/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/MimoJanra/TestOpsMCP/releases/tag/v1.0.0
