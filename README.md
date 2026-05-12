# 🧪 TestOps MCP — AI-Powered Test Orchestration

> **Control your entire test suite from Claude. No dashboard switching. No API calls. Just conversation.**

Integrate **Allure TestOps** with Claude using the Model Context Protocol. Launch tests, track execution, get reports—all within your AI assistant. Built for teams. Ready for production.

[![GitHub Release](https://img.shields.io/github/v/release/MimoJanra/TestOpsMCP?include_prereleases&label=Latest%20Release)](https://github.com/MimoJanra/TestOpsMCP/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://golang.org)

## Why TestOps MCP?

### ⚡ **Speed**
- Start a test run in seconds — no dashboard navigation
- Get live status updates without leaving Claude
- Automated reports delivered to your chat

### 🤖 **Intelligence**
- Claude analyzes test results and suggests fixes
- Ask natural questions: *"Why are these tests failing?"*
- Get insights without manual log parsing

### 👥 **Team-Ready**
- Deploy once, share with your whole team
- HTTP + auth for secure team access
- Works with Claude Desktop, Claude Web, and custom MCP clients

### 🏭 **Production-Grade**
- 57+ tools + full OpenAPI coverage (600+ endpoints searchable)
- Docker, Kubernetes, Systemd, ngrok support
- Enterprise security: JWT auth, CORS, TLS

---

## ✨ What You Can Do

**Launch & Monitor**
```
You: "Run smoke tests for project 1"
Claude: ✓ Launch started (ID: 12345)
        ↳ 156 tests queued
```

**Track Progress**
```
You: "What's the status?"
Claude: 📊 RUNNING — 89/156 passed
        ✓ 89 | ✗ 12 | ⚠️ 8 | ⊘ 47
```

**Deep Insights**
```
You: "Why did the API tests fail?"
Claude: The auth endpoint returned 401. 
        Likely cause: token refresh broken.
        Suggestion: Check token_expired tests first.
```

---

## 🚀 Setup Options

### **Option A: Per-User (Recommended for Teams)**
Each team member uses their own Allure API token.  
→ [Per-User Setup Guide](#per-user-setup-recommended)

### **Option B: Shared Server**
One server for the team, shared token.  
→ [Shared Server Setup Guide](#shared-server-setup)

---

## Per-User Setup (Recommended)

### 1️⃣ Download Binary for Your Platform

[**👉 Get Latest Release**](https://github.com/MimoJanra/TestOpsMCP/releases/latest)

- **Windows**: `testops-mcp-windows-amd64.exe`
- **macOS Intel**: `testops-mcp-macos-amd64`
- **macOS Apple Silicon**: `testops-mcp-macos-arm64`
- **Linux**: `testops-mcp-linux-amd64`

### 2️⃣ Get Your Allure API Token

1. Log in to **Allure TestOps**
2. Click **Settings → API tokens** (or your profile menu)
3. Create a new API token or copy an existing one
4. Keep this token safe (you'll need it in next step)

→ [Allure Docs: How to get API token](https://docs.qameta.io/allure-testops/advanced/api/)

### 3️⃣ Add to Claude Desktop Config

Open **Settings → Developer → Edit Config** in Claude Desktop.

Add this to the `mcpServers` section:

```json
{
  "mcpServers": {
    "testops": {
      "command": "C:\\Users\\YourName\\Downloads\\testops-mcp-windows-amd64.exe",
      "env": {
        "ALLURE_BASE_URL": "https://your-testops-instance.com",
        "ALLURE_TOKEN": "your-api-token-here"
      }
    }
  }
}
```

**Replace with your values:**
- `C:\\Users\\YourName\\Downloads\\testops-mcp-windows-amd64.exe` → path to your binary
- `https://your-testops-instance.com` → your Allure TestOps URL (e.g., `https://testops.mycompany.com`)
- `your-api-token-here` → your personal API token from step 2

### 4️⃣ Restart Claude

Close and reopen Claude Desktop. TestOps tools now appear in the tool dropdown! ✅

**That's it. You're ready to use it.**

---

## Alternative: Lazy Setup (⚠️ Less Secure)

**If you don't want to edit config files:**

1. Start the server without `ALLURE_TOKEN`:
```bash
./testops-mcp-linux-amd64 --http
```

2. Add to Claude Desktop config (no token):
```json
{
  "mcpServers": {
    "testops": {
      "command": "http://your-server:3000"
    }
  }
}
```

3. In Claude, use the `configure_allure_token` tool:
```
Me: "Set up my Allure token"
Claude: [uses configure_allure_token tool]
You: configure_allure_token(token="your-api-token-here")
```

**⚠️ WARNING:**
- Token will be visible in your chat history
- Anyone with chat access can see it
- Token is only stored for this session (lost after you close chat)
- **Not recommended for sensitive/shared accounts**

**Use this only if:**
- ✅ You're the only one with chat access
- ✅ You trust your chat history isn't logged
- ✅ You don't mind the token being visible in your conversation

**Recommended instead:** Always use the secure method above (config file).

---

## Shared Server Setup

If your team wants to run a **single shared server** instead of per-user setup:

### 1️⃣ Get a Shared API Token

1. In Allure TestOps, create a dedicated **service account** or use a shared admin token
2. Copy the API token
3. Keep it in a secure location (it will be in the server's `.env`)

### 2️⃣ Deploy the Server

```bash
# Download binary for your platform
# https://github.com/MimoJanra/TestOpsMCP/releases/latest

# Create .env file
cat > .env << EOF
ALLURE_BASE_URL=https://your-testops-instance.com
ALLURE_TOKEN=your-shared-api-token
PORT=3000
LOG_LEVEL=INFO
MCP_AUTH_TOKEN=secure-random-string-here
CORS_ALLOWED_ORIGIN=https://claude.ai
EOF

# Run in HTTP mode
./testops-mcp-linux-amd64 --http
```

Server starts on `http://localhost:3000`

→ **Full team deployment guide:** [Shared Server & DevOps Setup](#shared-server--devops-guide-below)

---

## Shared Server & DevOps Guide

**→ See [DEPLOYMENT.md](./docs/DEPLOYMENT.md) for:**
- 🐳 Docker & Docker Compose
- ☸️ Kubernetes setup
- 🔗 ngrok tunneling
- 🖥️ Systemd service
- 🔒 Nginx reverse proxy + HTTPS
- 📊 Monitoring & logging

```bash
# macOS/Linux
./testops-mcp-macos-amd64 --http

# Windows
testops-mcp-windows-amd64.exe --http

# Linux
./testops-mcp-linux-amd64 --http
```

Server starts on `http://localhost:3000` → [Full team setup guide →](#-team-deployment-http-mode)

---

## 🔧 Build from Source (Optional)

If you prefer to build the binary yourself:

```bash
# Clone
git clone https://github.com/MimoJanra/TestOpsMCP.git
cd TestOpsMCP

# Build
make build

# Binary created at: ./bin/server.exe (Windows)
```

Then follow **Step 2** above to add it to Claude Desktop config.

---

## 🛠️ 57+ Tools for Complete Test Automation

**Everything you need — plus full OpenAPI coverage:**

### 🔍 **Universal Tools** (Access any TestOps API endpoint)
- `configure_allure_token` — ⚠️ Optional: Set your token in chat (if not in config)
  - Only available if `ALLURE_TOKEN` not set in environment
  - Use only if you're lazy and trust your chat history
  - Token stored only for current session (not saved)
  - Example: `configure_allure_token(token="abc123xyz...")`
  - **Recommended:** Use secure config method instead (see setup above)
- `search_testops_operations` — Search for operations by intent/keyword
  - Example: `"list projects"`, `"create launch"`, `"get test results"`
  - Returns matching operations with schemas and parameters
  - Covers all 600+ TestOps API endpoints
- `execute_testops_operation` — Execute any discovered operation
  - Supports path parameters, query parameters, and request bodies
  - Handles all HTTP methods and response types
  - Uses same auth as all other tools

**Specialized Tools** (Pre-built for common operations):

### 🚀 Launch Management
- `run_allure_launch` — Start a test run
- `get_launch_status` — Real-time progress
- `get_launch_report` — Full statistics
- `list_launches` — Browse history
- `close_launch`, `reopen_launch` — Manage status
- `copy_launch`, `merge_launches` — Advanced operations

### 📊 Test Results
- `list_test_results` — Filter & find results
- `get_test_result` — Detailed analysis
- `assign_test_result` — Assign to team
- `mute_test_result` — Mark known issues
- `bulk_assign_test_results` — Batch operations
- `resolve_test_result` — Mark as fixed
- `unmute_test_result` — Re-activate checks

### ✏️ Test Cases & Scenarios
- `create_test_case` — Create new tests
- `update_test_case` — Edit all fields + steps
- `get_test_case` — Full details with execution steps
- `clone_test_case` — Duplicate with new ID
- `restore_test_case` — Recover deleted tests
- `create/update/delete_test_case_step` — Manage steps
- `run_test_case` — Execute immediately

### 👥 Collaboration
- `get_test_case_members` — See who's assigned
- `add_test_case_members` — Add team members
- `remove_test_case_members` — Unassign
- `get_test_case_external_links` — Track relations
- `add_test_case_external_link` — Link to GitHub, Jira, etc.

### 📈 Defects & Tracking
- `add_test_case_defect` — Attach bug
- `remove_test_case_defect` — Remove reference
- `get_test_case_defects` — See linked issues
- `get_launch_defects` — Launch-level tracking

### 📉 Analytics & Insights
- `get_project_stats` — Project overview
- `get_launch_trend_analytics` — Pass rate trends
- `get_launch_duration_analytics` — Execution time
- `get_test_success_rate` — Success metrics
- `get_launch_environment` — Environment details

---

## 📚 Documentation

| Guide | For | Time |
|-------|-----|------|
| [**Quick Start**](#-get-started-in-2-minutes) | First-time users | 2 min |
| [**Installation Guide**](./docs/INSTALLATION.md) | Local dev, Docker, Kubernetes | 10 min |
| [**Team Deployment**](#-team-deployment-http-mode) | Sharing with your team | 5 min |
| [**Deployment Guide**](./docs/DEPLOYMENT.md) | Production setup, reverse proxy | 15 min |
| [**API Reference**](./docs/API.md) | Tool parameters & examples | Reference |

---

## 💡 Common Use Cases

### ✅ Daily QA Workflow
*"Run regression tests for Sprint 22"*
- Claude starts the launch
- Gets real-time progress
- Reports results when done
- Suggests which tests to debug first

### 🔍 Investigate Test Failures
*"Why did the auth tests fail?"*
- Claude analyzes the failure logs
- Identifies the root cause
- Suggests fixes
- Links to related GitHub/Jira issues

### 🚀 Continuous Integration
*"What's the status of our nightly suite?"*
- Get daily reports in your messaging app
- Claude tracks trends over time
- Alerts if pass rate drops
- Suggests which tests are flaky

### 🤝 Team Collaboration
*"Assign the failing UI tests to @Sarah"*
- Claude bulk-assigns tests
- Creates/updates defect tickets
- Notifies team members
- Tracks assignment history

---

## ✅ Requirements

### For Local Claude Desktop (Easiest)
- ✅ Claude Desktop (macOS, Windows, Linux)
- ✅ Allure TestOps account with API access
- ✅ Pre-built binary (no installation needed)

### For Building from Source
- Go 1.22+ — [Download](https://golang.org/dl/)
- `make` — Usually pre-installed on macOS/Linux

### For Team Deployment
- Docker & Docker Compose — [Download](https://www.docker.com/)

---

## 🐳 Docker Image & Team Deployment

### Using Pre-built Docker Image

Pull from GitHub Container Registry:

```bash
docker pull ghcr.io/MimoJanra/TestOpsMCP:latest
```

Run in HTTP mode:

```bash
docker run \
  -e ALLURE_BASE_URL="https://your-testops.com" \
  -e ALLURE_TOKEN="your-api-token" \
  -p 3000:3000 \
  ghcr.io/MimoJanra/TestOpsMCP:latest --http
```

Server listens on `http://localhost:3000`

### Docker Compose (Recommended for Teams)

```bash
# 1. Create .env file
cp .env.example .env

# 2. Edit with your credentials
# ALLURE_BASE_URL=https://your-testops.com
# ALLURE_TOKEN=your_token

# 3. Start
docker-compose up -d

# Server now runs on :3000
```

Share with team members via Claude Desktop config:

```json
{
  "mcpServers": {
    "allure": {
      "command": "http://your-server:3000"
    }
  }
}
```

### Automatic Docker Builds

Docker images are **automatically built and published** to `ghcr.io` on every release:

- **Tags:** `latest`, `v1.2.1`, `semver`, `sha`
- **Architectures:** linux/amd64, linux/arm64
- **Auto-generated for:** Every git tag (`v*`)

---

## Configuration

All configuration is via environment variables.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ALLURE_BASE_URL` | **yes** | — | http(s) URL of your Allure TestOps instance (e.g., `https://testops.company.com`) |
| `ALLURE_TOKEN` | **no** * | — | API token for server/shared setup. If not set, each user must provide their token in Claude config |
| `REQUEST_TIMEOUT` | no | `30` | HTTP timeout for Allure calls, in seconds (1..600) |
| `PORT` | no | `3000` | Port the HTTP server listens on (stdio mode ignores this); accepts `3000` or `:3000` |
| `LOG_LEVEL` | no | `INFO` | One of `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `MCP_AUTH_TOKEN` | no | — | If set in HTTP mode, clients must send `Authorization: Bearer <token>` (recommended for security) |
| `CORS_ALLOWED_ORIGIN` | no | `*` | CORS `Access-Control-Allow-Origin` header (HTTP mode only); use `https://claude.ai` for Claude.ai integration |

**\*** `ALLURE_TOKEN`:
- **Shared server setup:** Set this, all users share one token
- **Per-user setup:** Leave empty, each user provides their token in Claude Desktop config

The server fails fast on startup if `ALLURE_BASE_URL` is missing or invalid.

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
    openapi_loader.go    # OpenAPI spec parser & operations index
    search_execute.go    # search + execute tool implementations
```

## 🚀 Deployment Options

### **🐳 Docker Compose** (Recommended for Teams)
```bash
docker-compose up -d  # Done!
```
✓ Works with multiple team members  
✓ Auto-restarts  
✓ 5-minute setup  
[Full guide →](./docs/DEPLOYMENT.md#docker-compose)

### **☁️ Kubernetes** (Enterprise Scale)
```bash
kubectl apply -f k8s-manifest.yaml
```
✓ Auto-scaling  
✓ Load balancing  
✓ Production-ready  
[Full guide →](./docs/DEPLOYMENT.md#kubernetes)

### **🔗 ngrok** (Quick Testing)
```bash
ngrok http 3000
```
✓ Instant HTTPS  
✓ Free tunneling  
✓ Share link with team  
[Full guide →](./docs/DEPLOYMENT.md#ngrok)

### **🖥️ Systemd** (Linux Servers)
```bash
sudo systemctl start allure-mcp
```
✓ Native Linux integration  
✓ Automatic startup  
[Full guide →](./docs/DEPLOYMENT.md#systemd)

📚 **Full deployment guide:** [DEPLOYMENT.md](./docs/DEPLOYMENT.md) — Includes reverse proxy, monitoring, TLS, scaling patterns

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

## 🆚 Why TestOps MCP?

### The Traditional Way ❌
```
1. Open Allure Dashboard
2. Find your project
3. Click "Run Tests"
4. Wait for email notification
5. Switch back to browser
6. Read report manually
7. Copy results to Slack
```
**15+ minutes. 4+ context switches. Tedious.**

### The TestOps MCP Way ✅
```
You: "Run smoke tests"
Claude: ✓ Started. 156 tests queued.

2 minutes later...

Claude: 📊 Results: 142 passed, 14 failed
        Root cause: API endpoint down.
        Suggested fix: Restart auth service.
```
**3 minutes. Zero context switches. Instant insights.**

### Head-to-Head Comparison

| | **TestOps MCP** | **Manual API** | **Dashboard Only** |
|---|---|---|---|
| **Launch from chat** | ✅ | ❌ | ❌ |
| **Real-time updates** | ✅ Live | ❌ Polling | ❌ Manual |
| **AI analysis** | ✅ Claude | ❌ No | ❌ No |
| **Bulk operations** | ✅ 57 tools + 600+ API | ⚠️ Limited | ❌ One-by-one |
| **Full API coverage** | ✅ Search + Execute | ❌ Hardcoded | ❌ Manual |
| **Team deployment** | ✅ Built-in | ⚠️ Fragile | ❌ Single-user |
| **Production-ready** | ✅ Yes | ⚠️ DIY | ❌ No |
| **Zero dashboard switching** | ✅ Yes | ❌ No | ❌ No |

---

## 💬 Examples

### Launch & Monitor
```
You: "Run E2E tests for project 1"

Claude: ✓ Launch created: testops-e2e-2024-05-08
        Project: Platform Tests
        Queued: 256 tests
        
You: "What's the status?"

Claude: 📊 RUNNING (42% complete)
        ✓ 107 passed
        ✗ 12 failed
        ⚠️ 5 broken
        ⊘ 132 pending
        ⏱️ Est. completion: 3 minutes
```

### Analyze Failures
```
You: "Why are the API tests failing?"

Claude: Looking at the failed tests...

        🔴 Failure Pattern Detected:
        - All failures in: api/auth/*
        - Error: 401 Unauthorized
        - Last successful run: 2 hours ago
        
        🔧 Most Likely Cause:
        API key rotation failed. Check token_refresh_service logs.
        
        💡 Suggested Actions:
        1. Restart auth service
        2. Re-run tests (3 min)
        3. Monitor token refresh logs
```

### Bulk Operations
```
You: "Mark all the UI tests as skipped for this sprint"

Claude: ✓ Found 34 UI tests in current launch
        Updating status to SKIPPED...
        
        ✓ 34 tests updated
        Launch report updated
        Team notified
```

### Team Management
```
You: "Assign all the failing tests to @QA_Team"

Claude: ✓ Found 12 failing tests
        Assigning to @QA_Team...
        
        ✓ 12 tests assigned
        Notifications sent
        Defect tickets created
```

---

## 🤝 Community & Support

**Got a question?**  
→ [Open a GitHub Issue](https://github.com/MimoJanra/TestOpsMCP/issues) or [Discussion](https://github.com/MimoJanra/TestOpsMCP/discussions)

**Found a bug?**  
→ [Report it](https://github.com/MimoJanra/TestOpsMCP/issues/new?template=bug_report.md)

**Have a feature idea?**  
→ [Share it](https://github.com/MimoJanra/TestOpsMCP/issues/new?template=feature_request.md)

**Security vulnerability?**  
→ **Don't** open a public issue. Email: `security@testopsmcp.dev`

**Want to contribute?**  
→ See [Contributing](#contributing) below

---

## Related Projects

- **[Model Context Protocol Spec](https://spec.modelcontextprotocol.io/)** — MCP standard
- **[Claude Desktop Docs](https://claude.ai/docs)** — How to configure MCP servers
- **[Allure TestOps](https://qameta.io/allure-testops/)** — Test execution & reporting
- **[MCP Ecosystem](https://modelcontextprotocol.io/servers)** — Other available MCP servers

---

## 🛠️ Contributing

**Love this project?** Help us make it better!

### Ways to Contribute
- 🐛 Fix bugs
- ✨ Add new features
- 📚 Improve documentation
- 🤔 Report issues
- 💡 Suggest improvements

### Quick Start
```bash
# Fork & clone
git clone https://github.com/YOUR_USERNAME/TestOpsMCP.git
cd TestOpsMCP

# Setup
go mod download
make build

# Run tests
make test

# Create your feature branch
git checkout -b feature/my-awesome-feature

# Code → Commit → Push → Open PR
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for full guidelines.

### Development Commands
```bash
make build      # Compile binary
make test       # Run unit tests
make lint       # Check code quality
make fmt        # Auto-format code
make check      # test + lint (recommended before PR)
make help       # See all commands
```

---

## License

[Apache License 2.0](LICENSE) — See [LICENSE](LICENSE) for full details.

---

---

## 🎯 Get Started Now

<table>
<tr>
<td width="50%">

### ⚡ First Time?
[Download the binary →](https://github.com/MimoJanra/TestOpsMCP/releases/latest)  
Takes 2 minutes to get running.

</td>
<td width="50%">

### 🚀 Ready to Deploy?
[See deployment options →](#-deployment-options)  
Docker, K8s, Systemd, and more.

</td>
</tr>
</table>

**Questions?** [Open an issue](https://github.com/MimoJanra/TestOpsMCP/issues) or [start a discussion](https://github.com/MimoJanra/TestOpsMCP/discussions)

**Like this project?** Star us on GitHub ⭐

---

## 🔍 Keywords

`test-automation` • `test-orchestration` • `allure-testops` • `mcp` • `claude-ai` • `ai-testing` • `qa-automation` • `golang` • `docker` • `kubernetes`

**Find us via:**
- "Allure TestOps + Claude Desktop"
- "AI test orchestration tool"
- "Model Context Protocol testing"
- "AI-powered QA automation"
- "Chat-based test runner"

---

<div align="center">

**Made with ❤️ by [Artem Alekseev](https://github.com/MimoJanra)**

[Apache License 2.0](LICENSE) — Free to use, modify, and distribute.

</div>
