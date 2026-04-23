# API Reference

Complete reference for Allure MCP Server tools and endpoints.

## Table of Contents

- [Tools](#tools)
- [HTTP Endpoints](#http-endpoints)
- [Protocol](#protocol)
- [Examples](#examples)
- [Error Handling](#error-handling)

## Tools

The server exposes **32+ tools** across 4 categories:

- **Launch Management** (9 tools) — Create, monitor, close launches
- **Test Results** (8 tools) — Manage test execution results
- **Test Cases** (9 tools) — CRUD operations on test cases
- **Analytics** (6 tools) — Projects, statistics, trends

---

## Launch Management Tools

### 1. `run_allure_launch`

Start a new test launch in Allure TestOps.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Allure project ID |
| `launch_name` | string | ✓ | Human-readable name for the launch |
| `auto_submit` | boolean | | Auto-submit when test collection complete (default: false) |

#### Response

```json
{
  "type": "text",
  "text": "Launch started: ID=12345, Name='Smoke Tests'"
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

Get current launch status.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |

#### Response: `CREATED`, `RUNNING`, `PAUSED`, `COMPLETED`, `SUBMITTED`

---

### 3. `get_launch_report`

Get execution statistics for a launch.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |

#### Response Fields: `total`, `passed`, `failed`, `broken`, `skipped`

---

### 4. `list_launches`

List launches in a project with pagination.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |
| `page` | integer | | Page number (0-based, default: 0) |
| `size` | integer | | Items per page (default: 10, max: 100) |

---

### 5. `get_launch_details`

Get comprehensive launch information.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |

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

## Test Results Management Tools

### 10. `list_test_results`

List test results in a launch with optional status filter.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | Launch ID |
| `status` | string | | Filter: PASSED, FAILED, BROKEN, SKIPPED |
| `page` | integer | | Page number (0-based) |
| `size` | integer | | Items per page |

---

### 11. `get_test_result`

Get detailed information about a single test result.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_result_id` | integer | ✓ | Test result ID |

---

### 12. `assign_test_result`

Assign a test result to a team member.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_result_id` | integer | ✓ | Test result ID |
| `username` | string | ✓ | Username to assign to |

---

### 13. `mute_test_result`

Mute a failing test result (mark as known issue).

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_result_id` | integer | ✓ | Test result ID |
| `reason` | string | | Reason for muting |

---

### 14-17. Bulk Test Result Operations

- **`bulk_assign_test_results`** — Assign multiple results at once
- **`bulk_mute_test_results`** — Mute multiple results
- **`bulk_unmute_test_results`** — Unmute multiple results
- **`bulk_resolve_test_results`** — Resolve multiple results

All take: `launch_id`, `test_result_ids` (array), and optional parameters.

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

---

### 19. `get_test_case`

Get test case details and steps.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_case_id` | integer | ✓ | Test case ID |

---

### 20. `create_test_case`

Create a new test case in a project.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |
| `name` | string | ✓ | Test case name |
| `description` | string | | Description (optional) |

---

### 21. `update_test_case`

Update an existing test case.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `test_case_id` | integer | ✓ | Test case ID |
| `name` | string | | New name |
| `description` | string | | New description |

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

### 28. `get_project`

Get project details and settings.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

### 29. `get_project_stats`

Get project statistics (test count, runs, automation %).

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

### 30. `get_launch_trend_analytics`

Get launch trend data over time (passed/failed/broken/skipped).

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

### 31. `get_launch_duration_analytics`

Get launch execution time distribution.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

### 32. `get_test_success_rate`

Get test case success rate metrics.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `project_id` | integer | ✓ | Project ID |

---

## HTTP Endpoints

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
- `Authorization: Bearer <MCP_AUTH_TOKEN>` (if `MCP_AUTH_TOKEN` is set)

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
- `Authorization: Bearer <MCP_AUTH_TOKEN>` (if required)
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

The server implements **MCP Protocol 2024-11-05** (JSON-RPC 2.0 subset).

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
| `401` | Unauthorized | Missing/invalid `MCP_AUTH_TOKEN` |
| `404` | Not Found | Unknown sessionId |
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
