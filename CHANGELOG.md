# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
