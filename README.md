# Allure MCP Server

A [Model Context Protocol](https://spec.modelcontextprotocol.io/) (MCP) server that integrates with
[Allure TestOps](https://qameta.io/allure-testops/) to launch test runs and fetch execution reports.

The server supports **two transport modes**:
- **stdio** (default) — reads JSON-RPC 2.0 requests from stdin, writes responses to stdout
- **HTTP** (with `--http` flag) — exposes SSE + JSON-RPC endpoints over HTTP

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

## Build

```bash
go build -o bin/server ./cmd/server
```

## Running

### Stdio mode (default) — for Claude Desktop (local development)

```bash
cp .env.example .env        # fill in ALLURE_BASE_URL and ALLURE_TOKEN
source .env                 # load env vars
./bin/server
```

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "allure": {
      "command": "/path/to/bin/server",
      "env": {
        "ALLURE_BASE_URL": "https://your-allure-domain.com",
        "ALLURE_TOKEN": "your_token"
      }
    }
  }
}
```

Then restart Claude Desktop and the tools will appear.

### HTTP mode — for shared server deployment

```bash
source .env
./bin/server --http
# listens on :3000 (configurable via PORT env)
```

**Recommended setup for team use:**

1. **On your server**, create a `.env` file with Allure credentials:
   ```bash
   ALLURE_BASE_URL=https://allure.example.com
   ALLURE_TOKEN=your_allure_token
   REQUEST_TIMEOUT=30
   MCP_AUTH_TOKEN=your_shared_secret_for_team
   CORS_ALLOWED_ORIGIN=*  # or restrict to specific domains
   PORT=3000
   LOG_LEVEL=INFO
   ```

2. **Run the server** (e.g., with systemd, docker, or process manager):
   ```bash
   ./bin/server --http
   ```

3. **Share the server URL with your team**: `http://your-server:3000`

4. **Team members configure their Claude Desktop** to connect to the shared server:
   ```json
   {
     "mcpServers": {
       "allure": {
         "url": "http://your-server:3000",
         "env": {
           "MCP_AUTH_TOKEN": "your_shared_secret_for_team"
         }
       }
     }
   }
   ```

5. **For production**, use HTTPS (self-signed cert + trusted CA, or ngrok):
   ```bash
   # Example with ngrok (https tunneling)
   ngrok http 3000
   # Share https://your-unique-id.ngrok.io with your team
   ```

## Configuration

All configuration is via environment variables.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ALLURE_BASE_URL` | yes | — | http(s) URL of your Allure TestOps instance |
| `ALLURE_TOKEN` | yes | — | API token used as `Authorization: Bearer <token>` |
| `REQUEST_TIMEOUT` | no | `30` | HTTP timeout for Allure calls, in seconds (1..600) |
| `PORT` | no | `3000` | Port the HTTP server listens on (stdio mode ignores this); accepts `3000` or `:3000` |
| `LOG_LEVEL` | no | `INFO` | One of `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `MCP_AUTH_TOKEN` | no | — | If set in HTTP mode, clients must send `Authorization: Bearer <token>` |
| `CORS_ALLOWED_ORIGIN` | no | `*` | CORS `Access-Control-Allow-Origin` header (HTTP mode only); empty disables |

The server fails fast on startup if a required variable is missing or invalid.

## HTTP transport (--http mode)

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

Accepts a single JSON-RPC 2.0 request. The server replies with `202 Accepted` and pushes the
JSON-RPC response to the SSE stream bound to the session.

Missing or unknown `sessionId` yields `400` / `404` respectively. Payloads are limited to 1 MiB.

### `OPTIONS`

Both endpoints respond to CORS preflight when `CORS_ALLOWED_ORIGIN` is set.

## Stdio transport (default mode)

Reads line-delimited JSON-RPC 2.0 from stdin, writes responses to stdout:

```bash
# echo a request
{ "jsonrpc":"2.0", "id":1, "method":"initialize", "params":{...} }
# read response
{ "jsonrpc":"2.0", "id":1, "result":{...} }
```

Each line must be valid JSON. Lines are processed sequentially; parsing errors receive an error response.

## Protocol

The server implements MCP protocol version `2024-11-05`. The expected client sequence is:

1. (HTTP only) Open `GET /sse`, read the `endpoint` event.
2. POST/write `initialize` request, wait for `initialize` result.
3. POST/write `notifications/initialized` (no response expected).
4. POST/write `tools/list` to discover tools; POST/write `tools/call` to invoke them.

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
    main.go              # entry point, mode dispatch (stdio vs HTTP)
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
    server.go            # MCP server core (shared by both transports)
    stdio.go             # stdio transport handler
  tools/
    registry.go          # tool registration & handlers
```

## Deployment

### Docker

Create a `Dockerfile`:

```dockerfile
FROM golang:1.22 AS builder
WORKDIR /build
COPY . .
RUN go build -o server ./cmd/server

FROM scratch
COPY --from=builder /build/server /server
ENTRYPOINT ["/server"]
```

Build and run:

```bash
docker build -t allure-mcp .
docker run -e ALLURE_BASE_URL=https://... \
           -e ALLURE_TOKEN=... \
           -e MCP_AUTH_TOKEN=... \
           -e LOG_LEVEL=INFO \
           -p 3000:3000 \
           allure-mcp --http
```

### Systemd service

Create `/etc/systemd/system/allure-mcp.service`:

```ini
[Unit]
Description=Allure MCP Server
After=network.target

[Service]
Type=simple
User=allure-mcp
WorkingDirectory=/opt/allure-mcp
ExecStart=/opt/allure-mcp/bin/server --http
Restart=on-failure
RestartSec=10
EnvironmentFile=/opt/allure-mcp/.env
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable allure-mcp
sudo systemctl start allure-mcp
sudo journalctl -u allure-mcp -f  # tail logs
```

## Allure TestOps integration

The server communicates with Allure TestOps using the Report Service API:

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

Logs go to stderr as one JSON object per line.

## Security notes

### Stdio mode (local development)

- No auth: runs as a subprocess with inherited privileges and direct stdin/stdout
- Suitable only for local development (Claude Desktop)
- No network exposure

### HTTP mode (team/shared server)

**Critical:** Always set `MCP_AUTH_TOKEN` when exposing the server over HTTP.

- `MCP_AUTH_TOKEN`: Bearer-token auth on `/sse` and `/messages`; clients must send `Authorization: Bearer <token>`
- Use **HTTPS** (or equivalent like ngrok) for production; HTTP plaintext exposes credentials
- CORS: `CORS_ALLOWED_ORIGIN=*` allows any site to call your server; use a concrete origin for production
  (e.g., `https://claude.ai` or your internal domain)
- Place the server behind a **reverse proxy** (nginx, Caddy) with additional auth (mTLS, IP whitelist)
- **Never** commit `.env` to git; use a secrets manager or `.env.local` (in .gitignore)
- **Rotate credentials** regularly; if Allure token or `MCP_AUTH_TOKEN` leaks, regenerate immediately
- **Monitor logs** for unauthorized access attempts

### Example production setup (nginx reverse proxy with HTTPS)

```nginx
server {
    listen 443 ssl;
    server_name allure-mcp.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    auth_request /auth;

    location /sse {
        proxy_pass http://localhost:3000/sse;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Authorization $http_authorization;
        proxy_pass_header Authorization;
    }

    location /messages {
        proxy_pass http://localhost:3000/messages;
        proxy_set_header Authorization $http_authorization;
        proxy_pass_header Authorization;
    }
}
```

Share `https://allure-mcp.example.com` with your team; they set it in their Claude Desktop config.

## License

Apache License 2.0. See [LICENSE](LICENSE).
