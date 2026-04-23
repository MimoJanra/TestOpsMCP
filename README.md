# Allure MCP Server

A [Model Context Protocol](https://spec.modelcontextprotocol.io/) (MCP) server that integrates with
[Allure TestOps](https://qameta.io/allure-testops/) to launch test runs and fetch execution reports.

The server uses the **MCP SSE transport over HTTP**: clients connect to `GET /sse` to receive a
per-session endpoint URL, then POST JSON-RPC 2.0 requests to `POST /messages?sessionId=<id>`.
Responses are delivered back through the SSE stream.

## Tools

| Tool | Description |
|------|-------------|
| `run_allure_launch` | Start a test launch in Allure TestOps |
| `get_launch_status` | Get the current status of a launch |
| `get_launch_report` | Get test execution statistics (total / passed / failed / broken) |

## Requirements

- Go 1.22+
- Access to an Allure TestOps instance
- Valid Allure API token

## Build & run

```bash
go build -o bin/server ./cmd/server
cp .env.example .env        # fill in ALLURE_BASE_URL and ALLURE_TOKEN
set -a && source .env && set +a
./bin/server
```

On Windows with PowerShell:

```powershell
go build -o bin/server.exe ./cmd/server
Copy-Item .env.example .env
# edit .env, then export variables into the process or use a loader
./bin/server.exe
```

## Configuration

All configuration is via environment variables.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ALLURE_BASE_URL` | yes | — | http(s) URL of your Allure TestOps instance |
| `ALLURE_TOKEN` | yes | — | API token used as `Authorization: Bearer <token>` |
| `REQUEST_TIMEOUT` | no | `30` | HTTP timeout for Allure calls, in seconds (1..600) |
| `PORT` | no | `3000` | Port the MCP server listens on; accepts `3000` or `:3000` |
| `LOG_LEVEL` | no | `INFO` | One of `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `MCP_AUTH_TOKEN` | no | — | If set, clients must send `Authorization: Bearer <token>` to `/sse` and `/messages` |
| `CORS_ALLOWED_ORIGIN` | no | `*` | Value for `Access-Control-Allow-Origin`; empty disables CORS headers |

The server fails fast on startup if a required variable is missing or invalid.

## HTTP endpoints

### `GET /sse`

Opens a Server-Sent Events stream. The first event carries the per-session message endpoint:

```
event: endpoint
data: /messages?sessionId=<hex-id>
```

Subsequent JSON-RPC responses are delivered as:

```
event: message
data: {"jsonrpc":"2.0","id":1,"result":{...}}
```

The stream also emits `:` ping comments every 25s as heartbeat.

### `POST /messages?sessionId=<id>`

Accepts a single JSON-RPC 2.0 request. Responses are *not* returned in the HTTP body — the server
replies with `202 Accepted` and pushes the JSON-RPC response to the SSE stream bound to the session.

Missing or unknown `sessionId` yields `400` / `404` respectively. Payloads are limited to 1 MiB.

### `OPTIONS`

Both endpoints respond to CORS preflight when `CORS_ALLOWED_ORIGIN` is set.

## Protocol

The server implements MCP protocol version `2024-11-05`. The expected client sequence is:

1. Open `GET /sse`, read the `endpoint` event.
2. POST `initialize` request, wait for `initialize` result via SSE.
3. POST `notifications/initialized` (no response expected).
4. POST `tools/list` to discover tools; POST `tools/call` to invoke them.

### Example — `tools/call`

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "run_allure_launch",
    "arguments": { "project_id": 1, "launch_name": "Smoke Tests" }
  }
}
```

### Tool errors

Tool-level failures are returned as a successful JSON-RPC result with `isError: true`:

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [{ "type": "text", "text": "Tool execution failed: project_id must be positive" }],
    "isError": true
  }
}
```

Protocol errors (parse error, method not found, invalid params) use the standard JSON-RPC `error`
object with codes `-32700`, `-32600`, `-32601`, `-32602`.

## Architecture

```
cmd/
  server/
    main.go              # entry point, graceful shutdown
internal/
  adapters/allure/
    client.go            # Allure TestOps HTTP client (with timeout)
    models.go            # request/response DTOs
  config/
    config.go            # env parsing & validation
  core/
    logger.go            # leveled structured JSON logger (stderr)
  mcp/
    protocol.go          # JSON-RPC 2.0 & MCP types
    server.go            # SSE transport, sessions, auth, CORS
  tools/
    registry.go          # tool registration & handlers
```

## Security notes

- `MCP_AUTH_TOKEN` is the only built-in authentication. The server is **not** hardened for public
  exposure — run it behind a reverse proxy or on a trusted network.
- When `CORS_ALLOWED_ORIGIN=*`, any site a browser visits can call your server. Use a concrete origin
  (or empty) for anything other than local development.
- The Allure token in `.env` is sensitive. `.env` is git-ignored; rotate the token if it ever leaks.

## Allure TestOps endpoints used

- `POST /api/rs/launch` — create launch
- `GET /api/rs/launch/{id}` — launch details
- `GET /api/rs/launch/{id}/statistic` — launch statistics

See [Allure TestOps API](https://docs.qameta.io/allure-testops/advanced/api/) for details.

## Development

```bash
go vet ./...
go test ./...
go build ./cmd/server
```

Logs go to stderr as one JSON object per line. stdout is unused.

## License

Apache License 2.0. See [LICENSE](LICENSE).
