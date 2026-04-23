# API Reference

Complete reference for Allure MCP Server tools and endpoints.

## Table of Contents

- [Tools](#tools)
- [HTTP Endpoints](#http-endpoints)
- [Protocol](#protocol)
- [Examples](#examples)
- [Error Handling](#error-handling)

## Tools

The server exposes three main tools:

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

Get the current status of a launch.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | ID of the launch to check |

#### Response

```json
{
  "type": "text",
  "text": "Launch ID=12345 Status=RUNNING, Tests: 42 total"
}
```

Status values:
- `CREATED` — Launch created but tests not yet collected
- `RUNNING` — Tests are executing
- `PAUSED` — Launch paused
- `COMPLETED` — All tests executed
- `SUBMITTED` — Results submitted to Allure Dashboard

#### Example

```bash
curl -X POST http://localhost:3000/messages?sessionId=abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "get_launch_status",
      "arguments": {
        "launch_id": 12345
      }
    }
  }'
```

---

### 3. `get_launch_report`

Get detailed execution statistics for a launch.

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `launch_id` | integer | ✓ | ID of the launch |

#### Response

```json
{
  "type": "text",
  "text": "Launch 12345 Results: Total=42, Passed=38, Failed=3, Broken=1"
}
```

#### Fields

- `total` — Total number of tests
- `passed` — Tests passed
- `failed` — Tests failed
- `broken` — Tests broken (setup errors)
- `skipped` — Tests skipped
- `unknown` — Tests with unknown status

#### Example

```bash
curl -X POST http://localhost:3000/messages?sessionId=abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "get_launch_report",
      "arguments": {
        "launch_id": 12345
      }
    }
  }'
```

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

### Complete Flow: Create Launch and Check Status

```bash
#!/bin/bash

BASE_URL="http://localhost:3000"
AUTH_TOKEN="your_mcp_auth_token"

# 1. Open SSE stream (in background)
SESSION_ID=$(curl -s "$BASE_URL/sse" \
  -H "Authorization: Bearer $AUTH_TOKEN" | \
  grep -oP '(?<=sessionId=)[^"]+' | head -1)

echo "Session ID: $SESSION_ID"

# 2. Create launch
curl -s -X POST "$BASE_URL/messages?sessionId=$SESSION_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "run_allure_launch",
      "arguments": {
        "project_id": 1,
        "launch_name": "E2E Tests"
      }
    }
  }'

# 3. Wait for response and extract launch ID
sleep 2

# 4. Check status (assuming launch_id=12345)
curl -s -X POST "$BASE_URL/messages?sessionId=$SESSION_ID" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "get_launch_status",
      "arguments": {
        "launch_id": 12345
      }
    }
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
