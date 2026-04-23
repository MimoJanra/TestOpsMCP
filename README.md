# Allure MCP Server

**AI-powered test orchestration for Allure TestOps via Claude**

A production-ready [Model Context Protocol](https://spec.modelcontextprotocol.io/) (MCP) server that seamlessly integrates Claude with [Allure TestOps](https://qameta.io/allure-testops/), enabling AI-assisted test launch management and execution reporting.

**Key Features:**
- 🚀 **Launch Test Runs** — Start Allure TestOps launches directly from Claude
- 📊 **Real-time Status Tracking** — Monitor launch progress and test execution
- 📈 **Execution Reports** — Get detailed statistics (pass/fail/broken/skipped rates)
- 🔌 **Two Transport Modes** — stdio (local) and HTTP (team/shared deployment)
- 🐳 **Docker Ready** — Production-grade Dockerfile and docker-compose config
- 🔐 **Enterprise Security** — Comprehensive auth, CORS, and TLS support
- 🌐 **SEO-Friendly** — Discoverable documentation and API reference

**Supported Platforms:**
- Claude Desktop (macOS, Windows, Linux)
- Claude Web (claude.ai)
- Custom MCP clients

---

## Quick Start (2 Minutes)

### 1. Clone & Build

```bash
git clone https://github.com/MimoJanra/TestOpsMCP.git
cd TestOpsMCP
make build
```

### 2. Configure

```bash
cp .env.example .env
# Edit .env with your Allure credentials
```

### 3. Run

```bash
make run  # Local development
# or
make run-http  # Team/server mode
```

### 4. Connect Claude Desktop

Edit your Claude config (Windows: `%APPDATA%\Claude\claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "allure": {
      "command": "C:\\Users\\YourName\\TestOpsMCP\\bin\\server.exe",
      "env": {
        "ALLURE_BASE_URL": "https://your-allure.com",
        "ALLURE_TOKEN": "your_token"
      }
    }
  }
}
```

Restart Claude Desktop — Allure tools appear in the dropdown.

---

## Tools Available

**32+ tools** for complete Allure TestOps automation:

### Launch Management
| Tool | Purpose |
|------|---------|
| `run_allure_launch` | Start a new test launch |
| `get_launch_status` | Check launch progress and status |
| `get_launch_report` | Get test statistics (pass/fail/broken) |
| `list_launches` | List launches with pagination |
| `get_launch_details` | Get comprehensive launch info |
| `close_launch` | Close/finish an active launch |
| `reopen_launch` | Reopen a closed launch |
| `add_test_cases_to_launch` | Add test cases to a launch |
| `add_test_plan_to_launch` | Add test plan to a launch |

### Test Results Management
| Tool | Purpose |
|------|---------|
| `list_test_results` | List test results with filter |
| `get_test_result` | Get detailed test result info |
| `assign_test_result` | Assign result to team member |
| `mute_test_result` | Mute failing tests (mark known issues) |
| `bulk_assign_test_results` | Bulk assign multiple results |
| `bulk_mute_test_results` | Bulk mute results |
| `bulk_unmute_test_results` | Bulk unmute results |
| `bulk_resolve_test_results` | Bulk resolve results |

### Test Cases & Management
| Tool | Purpose |
|------|---------|
| `list_test_cases` | List test cases in project |
| `get_test_case` | Get test case details |
| `create_test_case` | Create new test case |
| `update_test_case` | Update existing test case |
| `delete_test_case` | Delete test case |
| `run_test_case` | Start test case execution |
| `bulk_set_test_case_status` | Bulk update test case status |
| `bulk_add_test_case_tags` | Bulk add tags to test cases |
| `bulk_remove_test_case_tags` | Bulk remove tags from test cases |

### Projects & Analytics
| Tool | Purpose |
|------|---------|
| `list_projects` | List all accessible projects |
| `get_project` | Get project details |
| `get_project_stats` | Get project statistics |
| `get_launch_trend_analytics` | Get launch trend data (passed/failed/broken) |
| `get_launch_duration_analytics` | Get execution time distribution |
| `get_test_success_rate` | Get test case success rate metrics |

---

## Documentation

- **[Installation Guide](./docs/INSTALLATION.md)** — Detailed setup for local, Docker, Kubernetes
- **[Deployment Guide](./docs/DEPLOYMENT.md)** — Production patterns, reverse proxy, monitoring
- **[API Reference](./docs/API.md)** — Complete tool & endpoint documentation
- **[Architecture](./README.md#architecture)** — Code organization and design

---

## Requirements

- **Go 1.22+** — [Download](https://golang.org/dl/)
- **Allure TestOps** — Instance with API access
- **Claude Desktop** or MCP-compatible client

Optional:
- **Docker** — For containerized deployment
- **Docker Compose** — For team deployment

## Getting Started by Deployment Type

### 🖥️ Local Development (Claude Desktop)

See [Installation Guide](./docs/INSTALLATION.md#local-development) for step-by-step setup.

```bash
make build && make run
```

### 🐳 Docker (Single Instance)

```bash
docker build -t allure-mcp .
docker run -e ALLURE_BASE_URL=https://your-allure.com \
           -e ALLURE_TOKEN=your_token \
           -p 3000:3000 \
           allure-mcp --http
```

### 🔄 Docker Compose (Team)

```bash
cp .env.example .env
# Edit .env with your credentials
docker-compose up -d
```

### ☁️ Production (Kubernetes, Systemd, etc.)

See [Deployment Guide](./docs/DEPLOYMENT.md) for:
- Nginx reverse proxy with HTTPS
- Kubernetes manifests with auto-scaling
- Systemd service files
- Monitoring & health checks

## Team Deployment (HTTP Mode)

For shared server or team use, run in HTTP mode:

```bash
docker-compose up -d
# Server listens on :3000
```

**Team members connect via:**

```json
{
  "mcpServers": {
    "allure": {
      "url": "http://your-server:3000",
      "env": {
        "MCP_AUTH_TOKEN": "your_shared_secret_from_team"
      }
    }
  }
}
```

**For production HTTPS:**

Use Nginx reverse proxy (nginx config in [Deployment Guide](./docs/DEPLOYMENT.md#reverse-proxy-setup)):

```bash
# With Caddy (automatic HTTPS)
caddy run  # Reads from Caddyfile
```

Or use ngrok for quick HTTPS tunneling:

```bash
ngrok http 3000
# Share https://your-unique-id.ngrok.io with your team
```

📚 **Full setup:** [Deployment Guide](./docs/DEPLOYMENT.md)

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

## Deployment Options

| Method | Best For | Setup Time |
|--------|----------|-----------|
| **Docker Compose** | Teams, quick setup | ~5 min |
| **Kubernetes** | Large deployments, scaling | ~15 min |
| **Systemd** | Linux servers | ~10 min |
| **ngrok** | Quick HTTPS testing | ~1 min |
| **Nginx reverse proxy** | Production, custom domains | ~10 min |

**Docker Compose** (Recommended for teams):

```bash
docker-compose up -d
docker-compose logs -f
```

**Kubernetes:**

```bash
kubectl apply -f <(curl -s https://raw.githubusercontent.com/MimoJanra/TestOpsMCP/main/k8s-manifest.yaml)
```

**Systemd (Linux):**

```bash
sudo cp allure-mcp.service /etc/systemd/system/
sudo systemctl daemon-reload && sudo systemctl start allure-mcp
```

📚 **Full deployment guide with reverse proxy, monitoring, and scaling:** [Deployment Guide](./docs/DEPLOYMENT.md)

## API Reference

### Tools

Complete tool parameter reference and examples in [API.md](./docs/API.md#tools):

- **`run_allure_launch(project_id, launch_name)`** → Starts a test launch
- **`get_launch_status(launch_id)`** → Returns current status (CREATED, RUNNING, COMPLETED, etc.)
- **`get_launch_report(launch_id)`** → Returns statistics (total, passed, failed, broken, skipped)

### HTTP Endpoints

**For HTTP mode (`--http`):**

- `GET /sse` — Opens SSE stream for responses
- `POST /messages?sessionId=<id>` — Sends JSON-RPC requests
- `OPTIONS *` — CORS preflight

See [API.md](./docs/API.md#http-endpoints) for details.

### Allure TestOps Integration

The server uses the Allure Report Service API:

- `POST /api/rs/launch` — Create launch
- `GET /api/rs/launch/{id}` — Fetch launch details
- `GET /api/rs/launch/{id}/statistic` — Get statistics

📚 [Allure TestOps API Docs](https://docs.qameta.io/allure-testops/advanced/api/)

## Development

```bash
make build      # Compile binary to bin/server.exe
make run        # Run stdio mode (for Claude Desktop testing)
make run-http   # Run HTTP mode on :3000
make test       # Run unit tests
make lint       # Check code quality
make fmt        # Format code
make check      # Run lint + tests
make help       # Show all commands
```

### Logs

All output is JSON-formatted, one object per line to stderr:

```json
{"level":"INFO","msg":"Starting MCP server","mode":"http","port":3000,"timestamp":"2025-01-15T10:30:00Z"}
{"level":"DEBUG","msg":"Tool called","tool":"run_allure_launch","project_id":1}
```

Capture with:

```bash
docker-compose logs -f allure-mcp
# or
journalctl -u allure-mcp -f
```

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

## Comparison: Why Allure MCP?

| Feature | Allure MCP | Manual API | Manual Dashboard |
|---------|-----------|-----------|-----------------|
| Launch tests from Claude | ✓ | ✗ | ✗ |
| Check status in chat | ✓ | ✗ | ✗ |
| Get reports without switching apps | ✓ | Partial | ✗ |
| Team-ready deployment | ✓ | Requires integration | N/A |
| Local + server modes | ✓ | ✗ | ✗ |
| Docker + K8s ready | ✓ | ✗ | ✗ |
| Production-grade security | ✓ | Depends | N/A |

---

## Examples

### In Claude Desktop

Ask Claude:

> "Run the smoke tests for project 1 in Allure"

Claude uses the `run_allure_launch` tool automatically.

Or:

> "Check the status of launch 12345"

Claude uses `get_launch_status` and reports results.

### Via API (Programmatic)

See [API Reference](./docs/API.md#examples) for Python, Bash, cURL examples.

---

## Community & Support

- **Issues & Feature Requests:** [GitHub Issues](https://github.com/MimoJanra/TestOpsMCP/issues)
- **Questions:** Open a GitHub Discussion
- **Security Issues:** Email mimojanra@gmail.com (do not open public issues)

---

## Related Projects

- **[Model Context Protocol Spec](https://spec.modelcontextprotocol.io/)** — MCP standard
- **[Claude Desktop Docs](https://claude.ai/docs)** — How to configure MCP servers
- **[Allure TestOps](https://qameta.io/allure-testops/)** — Test execution & reporting
- **[MCP Ecosystem](https://modelcontextprotocol.io/servers)** — Other available MCP servers

---

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit changes (`git commit -m "Add my feature"`)
4. Push to branch (`git push origin feature/my-feature`)
5. Open a Pull Request

### Development Setup

```bash
git clone https://github.com/MimoJanra/TestOpsMCP.git
cd TestOpsMCP
make check   # Run tests + linting
```

---

## License

[Apache License 2.0](LICENSE) — See [LICENSE](LICENSE) for full details.

---

## Keywords

`test-orchestration`, `allure`, `mcp`, `claude`, `ai`, `testing`, `qa`, `automation`, `go`, `golang`, `docker`, `kubernetes`

**Useful searches:**
- Allure TestOps MCP integration
- Claude AI test orchestration
- Model Context Protocol test runner
- AI-powered test automation
- Allure TestOps + Claude Desktop integration
