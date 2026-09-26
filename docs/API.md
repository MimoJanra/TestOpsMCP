# API Reference

Complete reference for Allure MCP Server tools and endpoints.

## Table of Contents

- [Tools](#tools)
- [Prompts](#prompts)
- [Resources](#resources)
- [HTTP Endpoints](#http-endpoints)
- [Protocol](#protocol)
- [Examples](#examples)
- [Error Handling](#error-handling)

## Tools

The server exposes **148 tools** across multiple categories covering launches, test results, test cases, bulk operations, custom fields, analytics, async tasks, and AI analysis. See [llms-full.txt](../llms-full.txt) for the complete reference.

---

## Prompts

Built-in prompt templates are available via `prompts/list` and `prompts/get`. Clients can invoke them as slash commands or workflow shortcuts.

### `analyze-test-failures`

Instructs Claude to retrieve launch details, list all failed and broken results, group failures by error message or test suite, and surface the most critical issues.

**Arguments:**

| Name | Required | Description |
|------|----------|-------------|
| `launch_id` | ✓ | Allure launch ID to analyze |
| `project_id` | | Allure project ID for additional context |

### `launch-report-summary`

Instructs Claude to generate a concise executive summary including pass rate, total counts, duration, environment, and the top 3 failures.

**Arguments:**

| Name | Required | Description |
|------|----------|-------------|
| `launch_id` | ✓ | Allure launch ID |
| `project_id` | | Allure project ID |

---

## Resources

Resources are available via `resources/list` and `resources/read`.

| URI | MIME type | Description |
|-----|-----------|-------------|
| `allure://docs/quickstart` | `text/markdown` | Quickstart guide — tool groups, prompts, setup steps. Always available. |
| `ui://widgets/launch-dashboard` | `text/html;profile=mcp-app` | Interactive dashboard with pass/fail bar and stats (requires Allure token). |
| `ui://widgets/action-picker` | `text/html;profile=mcp-app` | Filterable picker for 600+ OpenAPI operations. |
| `ui://widgets/results-display` | `text/html;profile=mcp-app` | Formatted JSON viewer for `execute_testops_operation` results. |

The `launch-dashboard` resource supports **subscriptions** (`resources/subscribe`). When subscribed, the server polls launch status every 10 s and sends `notifications/resources/updated` when it changes.

---

## MCP Protocol (2025-11-25)

| Feature | Details |
|---------|---------|
| **Version** | `2025-11-25` (negotiated — accepts older client versions) |
| **Pagination** | `tools/list`, `resources/list`, `prompts/list` paginate at 50 items. Response includes `nextCursor`. |
| **Completion** | `completion/complete` — `project_id` and `launch_id` arguments return live suggestions from Allure API |
| **Elicitation** | Server confirms destructive operations via `elicitation/create`. Applied to `delete_test_case`, `bulk_delete_test_cases`, `remove_test_cases_from_launch` (mode=delete), `delete_launch`, `delete_test_plan`, `delete_defect`. |
| **Sampling** | Server can ask Claude via client with `sampling/createMessage`. Used by `analyze_launch_failures`. |
| **Subscriptions** | `resources/subscribe` / `resources/unsubscribe`. `notifications/resources/updated` on launch status change. |
| **Capabilities** | `tools`, `resources.subscribe=true`, `prompts`, `logging`, `elicitation` |
| **Client capabilities** | Tracked from `initialize`. Sampling/elicitation requests to a client that didn't declare them fail immediately instead of waiting for the timeout. |
| **Argument validation** | Each tool's schema `required` lists (top level and inside array items) are enforced before the tool runs. |
| **Message size** | Max 32 MiB per JSON-RPC message on stdio and HTTP. On stdio an oversized line gets a JSON-RPC error and the server keeps running. |

---

## Launch Management Tools

### 1. `run_allure_launch`

Start a new test launch in Allure TestOps. **Returns immediately** with a `task_id`; the launch is created asynchronously.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `launch_name` | string | ✓ | Human-readable name for the launch |

#### Response

```json
{
  "task_id": "a1b2c3d4e5f6...",
  "message": "Launch creation started. Use get_task_status to track progress."
}
```

Use `get_task_status` with the returned `task_id`. On completion, task result contains `{"launch_id": 123, "status": "started"}`.

#### Example

```json
{
  "task_id": "a1b2c3...",
  "message": "Launch creation started. Use get_task_status to track progress."
}
```

#### Example

```bash
curl -X POST http://localhost:3000/messages?sessionId=abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "run_allure_launch",
      "arguments": {
        "project_id": 1,
        "launch_name": "Smoke Tests"
      }
    }
  }'
```

---

### 2. `get_launch_status`

Get whether a launch is open or closed, plus its per-status result counts. (The launch DTO has no status field; this tool used to always return null.)

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |

#### Response

```json
{
  "launch_id": 123,
  "status": "OPEN",
  "closed": false,
  "statistic": {"total": 10, "passed": 7, "failed": 1, "broken": 1, "skipped": 0, "unknown": 1}
}
```

`status` is `OPEN` or `CLOSED`.

---

### 3. `get_launch_report`

Get execution statistics for a launch.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |

#### Response Fields: `total`, `passed`, `failed`, `broken`, `skipped`, `unknown`

---

### 4. `list_launches`

List launches in a project with pagination.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |
| `page` | integer | | Page number (0-based, default: 0) |
| `size` | integer | | Items per page (default: 10, max: 100) |

Each launch: `id`, `name`, `status` (`OPEN`/`CLOSED`), `closed`, `project_id`, `created_date`, `tags`.

---

### 5. `get_launch_details`

Get comprehensive launch information.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |

Returns `id`, `name`, `status` (`OPEN`/`CLOSED`), `closed`, `autoclose`, `external`, `project_id`, `created_date`, `last_modified_date`, `tags`, `links`, `issues`, and `statistic` (per-status counts).

---

### 6. `close_launch`

Close/finish an active launch.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |

---

### 7. `reopen_launch`

Reopen a closed launch for additional test results.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |

---

### 7a. `copy_launch`

Copy an existing launch, including its test results, into a new launch.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID to copy |
| `launch_name` | string | | Name for the new launch. Defaults to `"<original name> (copy)"` if omitted. |

Async — returns a `task_id`; poll with `get_task_status`.

---

### 7b. `update_launch`

Rename a launch or change its `autoClose`/`external` flags.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |
| `name` | string | | New launch name (optional) |
| `autoclose` | boolean | | Auto-close flag (optional) |
| `external` | boolean | | External flag (optional) |

---

### 8. `add_test_cases_to_launch`

Add test cases to a launch.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |
| `project_id` | integer | ✓ | Project ID |
| `test_case_ids` | array | ✓ | Test case IDs |
| `assignees` | array | | Usernames to assign to |

---

### 9. `add_test_plan_to_launch`

Add a test plan to a launch.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |
| `test_plan_id` | integer | ✓ | Test plan ID |

---

### 10. `remove_test_cases_from_launch`

Remove test cases from a launch. Resolves each `test_case_id` to its test result(s) in the launch (including retries) and removes them — use to trim a launch that has too many cases or duplicates.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID to remove test cases from |
| `test_case_ids` | integer[] | ✓ | Test case IDs to remove |
| `mode` | string | | `hide` (default) excludes results from the report but keeps the data; `delete` permanently deletes them |

Returns `removed_count`, `removed_result_ids`, `not_found_test_case_ids`, and `truncated`. In `delete` mode, any per-result failures are listed under `failed`. No bulk-delete API exists, so `delete` issues one `DELETE` per result.

---

### 10a. `delete_launch`

Permanently delete a launch and all its test results. Cannot be undone. Asks the user to confirm via elicitation.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |

Returns `{"status": "deleted"}`.

---

## Test Results Management Tools

### 11. `list_test_results`

List test results in a launch with optional status filter. The underlying API has no server-side status filter,
so a filtered request scans the launch's results client-side (capped at 20,000; see `truncated` in the response)
and paginates over the matches — `page`/`size` apply to the filtered list, not the launch's raw result order.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |
| `status` | string | | Filter: passed, failed, broken, skipped, unknown |
| `page` | integer | | Page number (0-based) |
| `size` | integer | | Items per page (default 10, max 1000) |

---

### 12. `get_test_result`

Get detailed information about a single test result, including `tested_by`, `hidden`, `manual`, `category`, `layer` and `links`.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_result_id` | integer | ✓ | Test result ID |

---

### 13. `assign_test_result`

Assign a test result to a team member.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_result_id` | integer | ✓ | Test result ID |
| `username` | string | ✓ | Username to assign to |

The stored assignee is read back after the call; the tool errors when Allure ignores the assignment (e.g. on already resolved results).

---

### 14. `mute_test_result`

Mute a failing test result (mark as known issue).

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_result_id` | integer | ✓ | Test result ID |
| `reason` | string | | Reason for muting |

The mute name column holds 255 characters: a longer reason is truncated for the name and kept whole as the reason (it used to 500).

---

### 14-17. Bulk Test Result Operations

- **`bulk_assign_test_results`** — Assign multiple results at once
- **`bulk_mute_test_results`** — Mute multiple results
- **`bulk_unmute_test_results`** — Unmute multiple results
- **`bulk_resolve_test_results`** — Resolve multiple results

All take: `launch_id`, `test_result_ids` (array), and optional parameters. `bulk_assign_test_results` distributes several assignees round-robin, unassigns when assignees are omitted, and skips not-run results. `bulk_unmute_test_results` also removes the test cases' mute records from the project's muted list.

---

## Test Cases Management Tools

### 18. `list_test_cases`

List test cases in a project.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |
| `page` | integer | | Page (0-based) |
| `size` | integer | | Items per page |

Each item includes `automation_status` (`automated` / `manual`).

---

### 19. `get_test_case`

Get test case details and steps.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_case_id` | integer | ✓ | Test case ID |

---

### 20. `create_test_case`

Create a new test case in a project, optionally with all its content in one call.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |
| `name` | string | ✓ | Test case name |
| `description` | string | | Description (markdown) |
| `precondition` | string | | Precondition |
| `expected_result` | string | | Overall expected result |
| `full_name` | string | | Full name, e.g. the automated test's qualified name |
| `automated` | boolean | | Mark as automated (default manual) |
| `status_id` | integer | | Status ID — must be passed together with `workflow_id` |
| `workflow_id` | integer | | Workflow ID — must be passed together with `status_id` |
| `test_layer_id` | integer | | Test layer ID |
| `tags` | array | | `[{id}]` or `[{name}]`; an unknown name creates the tag |
| `links` | array | | `[{name, type, url}]`, `url` required |
| `members` | array | | `[{id, name}]`, `id` is a project member id |
| `custom_fields` | array | | `[{custom_field_id, value_id}]` or `[{custom_field_id, name}]` — `name` finds or creates the value |
| `steps` | array | | `[{body, expected_result, steps: [...]}]` — `body` required, `steps` are sub-steps of the same shape |

Returns the new test case's id and `steps_created`.

---

### 21. `update_test_case`

Update an existing test case. Only the fields you pass change. `""` clears a text field and `[]` clears tags/members/links; tags, members and links replace the whole list.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_case_id` | integer | ✓ | Test case ID |
| `name` | string | | New name |
| `description` | string | | New description |
| `precondition` | string | | Precondition |
| `expected_result` | string | | Expected result |
| `full_name` | string | | Full name |
| `automated` / `external` / `deleted` | boolean | | Flags |
| `status_id` / `workflow_id` / `test_layer_id` | integer | | Status, workflow, test layer |
| `tags` / `members` / `links` | array | | Same shapes as `create_test_case` |
| `manual_scenario` | object | | `{"steps": [{body, expected_result, steps}]}` — REPLACES all steps (and their attachments); `{"steps": []}` removes them. The legacy `{type: "body"\|"expected", body}` entries are still accepted. |

---

### 22. `delete_test_case`

Delete a test case.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_case_id` | integer | ✓ | Test case ID |

---

### 23. `run_test_case`

Start a test run for a specific test case.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_case_id` | integer | ✓ | Test case ID |
| `launch_id` | integer | ✓ | Launch ID to run in |

---

### 24-26. Bulk Test Case Operations

- **`bulk_set_test_case_status`** — Update status for multiple cases
- **`bulk_add_test_case_tags`** — Add tags to multiple cases
- **`bulk_remove_test_case_tags`** — Remove tags from multiple cases
- **`bulk_clone_test_cases`** — Clone cases; optional `name_suffix`, `ignore_tags`. The API clones asynchronously and returns no ids — find the clones with `search_test_cases`.
- **`bulk_run_test_cases_new_launch`** — `launch_name` is required; the task result includes the new `launch_id`.
- **`bulk_create_test_plan`** — Returns `test_plan_id`, `name`, `test_cases_count`.

---

## Test Case Steps, Folders & Attachments

| Tool | Parameters | Notes |
|------|------------|-------|
| `create_test_case_step` | `test_case_id`✓, `body`✓, `expected_result`, `parent_id`, `after_id`, `before_id` | The parent of a nested `after_id`/`before_id` anchor is resolved automatically. |
| `move_test_case_step` | `step_id`✓, `test_case_id`, `after_id`, `before_id`, `parent_id` | One position is required. Pass `test_case_id` with a nested anchor. |
| `copy_test_case_step` | `step_id`✓, `test_case_id`, `after_id`, `before_id`, `parent_id` | Returns the copy's `step_id` when `test_case_id` is given. |
| `detach_test_case_automation` | `test_case_id`✓, `status_id`, `workflow_id`, `use_scenario_from_test_result` | Omitted status/workflow keep the current values. |
| `add_test_case_members` | `test_case_id`✓, `members`✓ (`[{id, name, role?}]`) | Additive; `role` optional. |
| `remove_test_case_members` | `test_case_id`✓, `member_ids`✓, `project_id` | `project_id` looked up; a mismatching one is refused. |
| `rename_test_case_folder` | `project_id`✓, `tree_id`, `node_id`✓, `name`✓ | Folder keeps its node id; test cases follow. |
| `delete_test_case_folder` | `project_id`✓, `tree_id`, `node_id`✓ | Test cases are unassigned, not deleted. Status `emptied` when the folder's value backs another folder. |
| `upload_test_case_attachment` | `test_case_id`✓, `file_path` (stdio only) or `content_base64` + `file_name`, `content_type`, `step_id`, `target` | Max 20 MiB. |
| `get_test_case_attachment_content` | `attachment_id`✓, `save_to` (stdio only) | Text as `content`, binary as `content_base64`; `save_to` never overwrites. Max 20 MiB. |

`tree_id` is optional when the project has exactly one tree (`list_test_case_trees`).

---

## Test Plan Tools

| Tool | Parameters | Notes |
|------|------------|-------|
| `list_test_plans` | `project_id`✓, `name`, `page`, `size` (default 20, max 100) | Newest first. |
| `get_test_plan` | `test_plan_id`✓ | Includes the AQL it selects by (`base_rql`). |
| `run_test_plan` | `test_plan_id`✓, `launch_name`✓ | Starts a new launch; returns `launch_id`. |
| `rename_test_plan` | `test_plan_id`✓, `name`✓ | |
| `delete_test_plan` | `test_plan_id`✓ | Asks for confirmation. Test cases and past launches are not affected. |

Create plans with `bulk_create_test_plan`; add one to an existing launch with `add_test_plan_to_launch`.

---

## Defect Tools

| Tool | Parameters | Notes |
|------|------------|-------|
| `list_defects` | `project_id`✓, `name_filter`, `status` (`open`/`closed`), `page`, `size` (default 20, max 100) | Newest first. |
| `get_defect` | `defect_id`✓ | Name, description, open/closed state, linked issue. |
| `create_defect` | `project_id`✓, `name`✓, `description` | Link to test cases with `add_test_case_defect`. |
| `update_defect` | `defect_id`✓, `name`, `description`, `closed` | `closed: true` closes, `false` reopens. |
| `delete_defect` | `defect_id`✓ | Asks for confirmation. Prefer `update_defect` with `closed: true`. |

---

## Projects & Analytics Tools

### 27. `list_projects`

List all accessible projects.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `page` | integer | | Page (0-based) |
| `size` | integer | | Items per page |

---

### 28. `find_project`

Find a project by name or code (case-insensitive substring match). Resolves a human-readable name/code (e.g. `TSi`) to its numeric project ID — instead of paging through `list_projects` or guessing IDs.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | ✓ | Name or code to match (case-insensitive substring) |
| `limit` | integer | | Max matches to return (1–100, default 20) |

Returns `matches` (array of `{id, name, code}`), `count`, `scanned`, and `truncated` (`true` if more matches may exist beyond `limit` or the scan cap).

---

### 29. `get_project`

Get project details and settings.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

### 30. `get_project_stats`

Get project statistics (test count, runs, automation %).

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

### 31. `get_launch_trend_analytics`

Get launch trend data over time (passed/failed/broken/skipped).

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

### 32. `get_launch_duration_analytics`

Get launch execution time distribution.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

### 33. `get_test_success_rate`

Get test case success rate metrics.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

## HTTP Endpoints

### `GET /health`

Liveness probe — returns `200 ok` with no authentication required. Used by Docker, Kubernetes, and uptime monitors.

```bash
curl http://localhost:3000/health
# ok
```

---

### `GET /sse`

Opens a Server-Sent Events (SSE) stream for receiving async responses.

**Request:**

```bash
curl http://localhost:3000/sse
```

**Response Stream:**

```
event: endpoint
data: /messages?sessionId=e8f2a3c9d4e1b6f2

event: message
data: {"jsonrpc":"2.0","id":1,"result":{"content":[...]}}

: heartbeat comment
```

**Headers:**
- `Authorization: Bearer <your-token>` — required when `MCP_AUTH_TOKENS` is configured on the server

---

### `POST /messages`

Send a JSON-RPC 2.0 request and receive async response on SSE stream.

**Query Parameters:**

| Parameter | Required | Description |
|-----------|----------|-------------|
| `sessionId` | ✓ | Session ID from SSE `/endpoint` event |

**Request Body:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "run_allure_launch",
    "arguments": { "project_id": 1, "launch_name": "Tests" }
  }
}
```

**Response:**

```
HTTP/1.1 202 Accepted
Content-Length: 0
```

The actual response is delivered to the SSE stream.

**Headers:**
- `Authorization: Bearer <your-token>` — required when `MCP_AUTH_TOKENS` is configured on the server
- `Content-Type: application/json`

---

### `OPTIONS /sse` and `OPTIONS /messages`

CORS preflight handling (when `CORS_ALLOWED_ORIGIN` is set).

**Response Headers:**

```
Access-Control-Allow-Origin: <CORS_ALLOWED_ORIGIN>
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Max-Age: 86400
```

---

## Protocol

The server implements **MCP Protocol 2025-11-25** (JSON-RPC 2.0 subset).

### Initialization Sequence

```
1. Client opens GET /sse
   ↓
2. Server sends: event: endpoint, data: /messages?sessionId=...
   ↓
3. Client sends initialize request
4. Server responds with initialize result
   ↓
5. Client sends notifications/initialized
   ↓
6. Client can now call tools via tools/call
```

### JSON-RPC 2.0 Format

**Request:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "run_allure_launch",
    "arguments": { "project_id": 1, "launch_name": "Tests" }
  }
}
```

**Success Response:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Launch started: ID=12345"
      }
    ]
  }
}
```

**Error Response (MCP Protocol Error):**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": { "reason": "project_id must be positive" }
  }
}
```

**Tool-Level Error:**

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Error: Allure API returned 401 Unauthorized"
      }
    ],
    "isError": true
  }
}
```

---

## Examples

### Example: Launch, Run Tests, and Report

```bash
#!/bin/bash
# See examples/launch-tests.sh for complete automation script

BASE_URL="http://localhost:3000"
AUTH_TOKEN="your_mcp_auth_token"

# 1. Open SSE stream
SESSION_ID=$(curl -s "$BASE_URL/sse" \
  -H "Authorization: Bearer $AUTH_TOKEN" | \
  grep -oP '(?<=sessionId=)[^"]+' | head -1)

# 2. Create launch
LAUNCH=$(curl -s -X POST "$BASE_URL/messages?sessionId=$SESSION_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "run_allure_launch",
      "arguments": {"project_id": 1, "launch_name": "Regression"}
    }
  }')

LAUNCH_ID=$(echo $LAUNCH | grep -o '"launch_id":[0-9]*' | grep -o '[0-9]*')

# 3. Add test cases
curl -s -X POST "$BASE_URL/messages?sessionId=$SESSION_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "add_test_cases_to_launch",
      "arguments": {
        "launch_id": '$LAUNCH_ID',
        "project_id": 1,
        "test_case_ids": [1, 2, 3]
      }
    }
  }'

# 4. Monitor and close
sleep 60  # Wait for execution
curl -s -X POST "$BASE_URL/messages?sessionId=$SESSION_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {"name": "close_launch", "arguments": {"launch_id": '$LAUNCH_ID'}}
  }'
```

### Python Client Example

```python
import httpx
import json
import asyncio

async def call_tool(session_id, auth_token, tool_name, arguments):
    """Call a tool via Allure MCP server."""
    async with httpx.AsyncClient() as client:
        # Send request
        response = await client.post(
            f"http://localhost:3000/messages?sessionId={session_id}",
            headers={"Authorization": f"Bearer {auth_token}"},
            json={
                "jsonrpc": "2.0",
                "id": 1,
                "method": "tools/call",
                "params": {
                    "name": tool_name,
                    "arguments": arguments
                }
            }
        )
        return response.status_code, response.text

async def main():
    session_id = "abc123"
    auth_token = "your_token"
    
    # Run launch
    status, response = await call_tool(
        session_id, auth_token,
        "run_allure_launch",
        {"project_id": 1, "launch_name": "Tests"}
    )
    print(f"Status: {status}, Response: {response}")

asyncio.run(main())
```

---

## Error Handling

### HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| `200` | OK (SSE stream opened) | GET /sse |
| `202` | Accepted (request queued) | POST /messages |
| `400` | Bad Request | Missing sessionId, invalid JSON |
| `401` | Unauthorized | Missing or invalid bearer token |
| `404` | Not Found | Unknown sessionId |
| `413` | Request Entity Too Large | Body over 32 MiB |
| `500` | Server Error | Crash, internal bug |

### JSON-RPC Error Codes

| Code | Description |
|------|-------------|
| `-32700` | Parse error |
| `-32600` | Invalid Request |
| `-32601` | Method not found |
| `-32602` | Invalid params |
| `-32603` | Internal error |

### Common Errors

#### "project_id must be positive"

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": { "reason": "project_id must be positive" }
  }
}
```

**Fix:** Use a valid Allure project ID (integer > 0)

#### "Allure API returned 401 Unauthorized"

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [{ "type": "text", "text": "Tool execution failed: ..." }],
    "isError": true
  }
}
```

**Fix:** Verify `ALLURE_BASE_URL` and `ALLURE_TOKEN` are correct

#### "launch not found"

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [{ "type": "text", "text": "Tool execution failed: launch_id=99999 not found" }],
    "isError": true
  }
}
```

**Fix:** Use correct `launch_id` from `run_allure_launch` response

---

## Async Task Tools

### `get_task_status`

Get the current status of a background task.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_id` | string | ✓ | Task ID returned by an async tool |

**Response:**
```json
{
  "id": "a1b2c3...",
  "tool": "run_allure_launch",
  "status": "working",
  "created_at": "2026-05-31T12:00:00Z",
  "updated_at": "2026-05-31T12:00:01Z"
}
```

`status` values: `working` / `succeeded` / `failed` / `cancelled`. On `succeeded`, includes `result` field with the tool's output. On `failed`, includes `error` string.

### `list_running_tasks`

No parameters. Returns `{"tasks": [...], "count": N}` — only tasks with `status=working`.

### `cancel_task`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `task_id` | string | ✓ | Task ID to cancel |

Returns `{"status": "cancelled", "task_id": "..."}`. Signals the background goroutine via context cancellation.

---

## AI Analysis Tools

### `analyze_launch_failures`

Fetches failed and broken test results and uses MCP sampling to ask Claude for root-cause analysis. Errors on a nonexistent launch.

**Requires:** MCP client that supports `sampling/createMessage` (Claude Desktop, claude.ai).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Allure launch ID to analyze |
| `max_failures` | integer | | Max failures to analyze (default 20, max 50) |

**Response:**
```json
{
  "launch_id": 12345,
  "failures": 8,
  "analysis": "Root causes identified:\n1. Auth token expiry (4 tests)...\n\nSuggested fixes:\n..."
}
```

If sampling is unavailable, returns `summary` of raw failures with an `error` field.

---

## Custom Field Management Tools

Three layers: a custom field **definition** (org-wide, e.g. "Priority"), its **project attachment** (a project opts in, and sets required/locked/default there), and its **value catalog** within a project (the selectable options, e.g. "High"/"Low"). Setting values *on a test case* is `update_test_case_custom_fields` (see Test Cases Management Tools) — these tools manage the fields and their options, not test-case assignments.

### `create_custom_field`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | ✓ | Field name (e.g. "Severity") |
| `required` | boolean | ✓ | Whether a value must be set on every test case it's attached to |
| `single_select` | boolean | | At most one value per test case (default false) |

### `get_custom_field` / `delete_custom_field`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `custom_field_id` | integer | ✓ | Custom field ID |

`delete_custom_field` is permanent — prefer `set_custom_field_archived` for a reversible removal. The API refuses while the field is attached to a project: clear its values from test cases, `remove_custom_field_from_project`, then delete.

Built-in fields (negative ids: Epic/Feature/Story/Suite/Component) are accepted by `get_custom_field`, `get_project_custom_field`, `update_project_custom_field`, `create_custom_field_value` and `list_custom_field_values`.

### `update_custom_field`

All fields optional — only the ones passed are changed.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `custom_field_id` | integer | ✓ | Custom field ID |
| `name` | string | | New field name |
| `required` | boolean | | Whether a value must be set |
| `single_select` | boolean | | At most one value per test case |
| `locked` | boolean | | Locked flag. Does not block renaming the field or adding/renaming values (confirmed live). |

### `set_custom_field_archived`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `custom_field_id` | integer | ✓ | Custom field ID |
| `archived` | boolean | ✓ | `true` to archive, `false` to restore |

### `list_project_custom_fields`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `query` | string | | Filter by field name |
| `page` | integer | | Page number (0-based), default 0 |
| `size` | integer | | Items per page, default 10 |

### `get_project_custom_field`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `custom_field_id` | integer | ✓ | Custom field ID |

Returns the field's project-scoped required/locked/default settings plus the field definition. A field not attached to the project returns a clear error.

### `add_custom_fields_to_project`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `custom_field_ids` | integer[] | ✓ | IDs of existing custom field definitions to attach |

Reports ids that were not attached (the API silently ignores unknown ids).

### `remove_custom_field_from_project`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `custom_field_id` | integer | ✓ | Custom field ID |

Detaches the field from the project only — does not delete the field definition or its values. The API refuses while any test case in the project has a value for it; clear those first with `bulk_remove_test_case_custom_fields`.

### `update_project_custom_field`

The current replacement for the API's deprecated set-required/set-default endpoints. All fields optional — only the ones passed are changed.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `custom_field_id` | integer | ✓ | Custom field ID |
| `required` | boolean | | Whether a value must be set on this project's test cases |
| `locked` | boolean | | Lock this project's use of the field |
| `default_custom_field_value_id` | integer | | Value ID to use as the default in this project |

### `create_custom_field_value`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `custom_field_id` | integer | ✓ | Custom field to add the value to |
| `name` | string | ✓ | The new value's display name |
| `default` | boolean | | Make this the field's default value in this project |

Use `list_custom_field_values` first to check the value doesn't already exist.

### `update_custom_field_value`

All fields optional — only the ones passed are changed. Renaming gives the value a new `value_id` (the old id stops existing); test cases that had the value follow it. The API returns no body — call `list_custom_field_values` for the new id.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `value_id` | integer | ✓ | Custom field value ID |
| `name` | string | | New display name |
| `default` | boolean | | Make this the field's default value in this project |
| `global` | boolean | | Share this value across all projects. Cannot be undone. |

### `delete_custom_field_value`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `value_id` | integer | ✓ | Custom field value ID |

Test cases currently set to this value will lose it.
