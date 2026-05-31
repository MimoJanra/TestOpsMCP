# 🧪 TestOps MCP — AI-Powered Test Orchestration

> **Control your entire test suite from Claude. No dashboard switching. No API calls. Just conversation.**

Integrate **Allure TestOps** with Claude using the Model Context Protocol. Launch tests, track execution, get reports—all within your AI assistant. Built for teams. Ready for production.

[![GitHub Release](https://img.shields.io/github/v/release/MimoJanra/TestOpsMCP?include_prereleases&label=Latest%20Release)](https://github.com/MimoJanra/TestOpsMCP/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go)](https://golang.org)
[![MCP Badge](https://lobehub.com/badge/mcp/mimojanra-testopsmcp?style=plastic)](https://lobehub.com/mcp/mimojanra-testopsmcp)## Why TestOps MCP?

### ⚡ Speed
- Start a test run in seconds — no dashboard navigation
- Get live status updates without leaving Claude
- Automated reports delivered to your chat

### 🤖 Intelligence
- Claude analyzes test results and suggests fixes
- Ask natural questions: *"Why are these tests failing?"*
- Get insights without manual log parsing

### 👥 Team-Ready
- Deploy once, share with your whole team
- HTTP + auth for secure team access
- Works with Claude Desktop, Claude Web, and custom MCP clients

### 🏭 Production-Grade
- **104 tools** — complete Allure TestOps test case API coverage + full OpenAPI fallback (600+ endpoints)
- **MCP Prompts** — built-in templates (`analyze-test-failures`, `launch-report-summary`) for one-click workflows
- **MCP Resources** — attach `allure://docs/quickstart` as context; widget resources for visual dashboards
- **AI analysis** — `analyze_launch_failures` asks Claude to find root causes via MCP sampling
- **Async tasks** — long-running operations return immediately with `task_id`; track with `get_task_status`
- **Confirmation dialogs** — destructive operations (delete) ask for confirmation via MCP elicitation
- Docker, Kubernetes, Systemd, ngrok support
- Enterprise security: JWT auth, CORS, TLS
- MCP protocol 2025-11-25 with version negotiation, pagination, subscriptions, completion

---

## 🎯 Quick Start (2 minutes)

Your team lead has already deployed TestOps MCP. Here's how to connect:

### 1. Open Claude Desktop Settings
Click **Settings** (bottom left) → **Developer** tab

### 2. Click "Edit Config"
Add this to the `mcpServers` section:

**Windows:**
```json
{
  "mcpServers": {
    "testops": {
      "command": "cmd",
      "args": [
        "/c", "npx", "-y", "mcp-remote",
        "http://your-server.com:3000/mcp",
        "--header", "Authorization:Bearer your-mcp-auth-token",
        "--header", "X-Allure-Token:your-personal-allure-token"
      ]
    }
  }
}
```

**macOS / Linux:**
```json
{
  "mcpServers": {
    "testops": {
      "command": "npx",
      "args": [
        "-y", "mcp-remote",
        "http://your-server.com:3000/mcp",
        "--header", "Authorization:Bearer your-mcp-auth-token",
        "--header", "X-Allure-Token:your-personal-allure-token"
      ]
    }
  }
}
```

- **`Authorization`** *(optional)* — your personal bearer token from `MCP_AUTH_TOKENS` on the server. Omit the `--header "Authorization:..."` arg if the server has no auth configured.
- **`X-Allure-Token`** — your personal Allure API token. All actions in Allure will be done under your account.

### 3. Restart Claude Desktop
Close and reopen Claude.

### 4. Try it!
Ask Claude: `"List all projects in Allure"`

You should see your TestOps projects! ✅

**Setting up a server for your team?** → See [Installation & Deployment](#-installation-options)

---

## ✨ What You Can Do

```
You: "Run smoke tests for project 1"
Claude: ✓ Launch started (ID: 12345)
        ↳ 156 tests queued

You: "What's the status?"
Claude: 📊 RUNNING — 89/156 passed
        ✓ 89 | ✗ 12 | ⚠️ 8 | ⊘ 47

You: "Why did the API tests fail?"
Claude: The auth endpoint returned 401. 
        Likely cause: token refresh broken.
        Suggestion: Check token_expired tests first.
```

---

## 📦 Installation Options

### **Option A: Per-User (Recommended)**
Each team member uses their own API token.
→ [Per-User Setup](#per-user-setup)

### **Option B: Shared Server**
One server for the team with a shared token.
→ [Shared Server Setup](#shared-server-setup)

---

## Per-User Setup

**Best for:** Teams where each person has their own Allure account

1. Download the binary for your OS from [Releases](https://github.com/MimoJanra/TestOpsMCP/releases/latest)

2. Get your Allure API token:
   - Log into Allure TestOps (your-testops.com)
   - Click your **profile icon** (top right) or go to **Settings**
   - Select **API tokens** or **Integration tokens**
   - Click **+ Create token** (or **Generate**)
   - Give it a name like `Claude MCP`
   - Copy the token (it won't be shown again!)
   - Store it somewhere safe

3. Add to Claude Desktop config (**Settings → Developer → Edit Config**):

```json
{
  "mcpServers": {
    "testops": {
      "command": "C:\\Users\\YourName\\Downloads\\testops-mcp-windows-amd64.exe",
      "env": {
        "ALLURE_BASE_URL": "https://your-testops.com",
        "ALLURE_TOKEN": "your-token-here",
        "REQUEST_TIMEOUT": "30",
        "LOG_LEVEL": "INFO"
      }
    }
  }
}
```

4. Restart Claude Desktop and test:
```
"List all projects in Allure"
```

---

## Shared Server Setup

**Best for:** Teams using a single shared Allure account or service account

### Quick Start with Docker Compose (Recommended)

**1. Clone repo and create .env:**
```bash
# Clone repo
git clone https://github.com/MimoJanra/TestOpsMCP.git
cd TestOpsMCP

# Create .env file (ALLURE_TOKEN is optional - each user provides their own)
cat > .env << EOF
ALLURE_BASE_URL=https://your-testops.com
PORT=3000
LOG_LEVEL=INFO
MCP_AUTH_TOKEN=generate-random-string-here
EOF
```

**Optional:** If you want a fallback shared token for the server:
```bash
# Add to .env (optional)
ALLURE_TOKEN=your-shared-token
```

**2. Start with Docker Compose:**
```bash
# Start server
docker-compose up -d

# Check logs
docker-compose logs -f
```

Server runs on `http://localhost:3000`

**3. Team members connect with their own tokens**

Each team member adds to their Claude Desktop config (**Settings → Developer → Edit Config**):

**Windows:**
```json
{
  "mcpServers": {
    "testops": {
      "command": "cmd",
      "args": [
        "/c", "npx", "-y", "mcp-remote",
        "http://your-server.com:3000/mcp",
        "--header", "Authorization:Bearer your-mcp-auth-token",
        "--header", "X-Allure-Token:your-personal-allure-token"
      ]
    }
  }
}
```

**macOS / Linux:**
```json
{
  "mcpServers": {
    "testops": {
      "command": "npx",
      "args": [
        "-y", "mcp-remote",
        "http://your-server.com:3000/mcp",
        "--header", "Authorization:Bearer your-mcp-auth-token",
        "--header", "X-Allure-Token:your-personal-allure-token"
      ]
    }
  }
}
```

- **`Authorization`** — each user's named bearer token from `MCP_AUTH_TOKENS`. Omit the arg entirely if no auth is configured.
- **`X-Allure-Token`** — each person's own Allure API token. Actions in Allure are performed under that account.

Each person needs to:
1. Get their personal Allure API token (Settings → API tokens in Allure TestOps)
2. Add it to their Claude Desktop config as `X-Allure-Token`
3. Restart Claude Desktop

All actions will be done from their personal Allure account! ✅

### Binary Setup (Alternative)

If you prefer running without Docker:

```bash
# Get binary for your OS
./testops-mcp-linux-amd64 --http
```

---

**For production deployment** (Kubernetes, Systemd, reverse proxy, HTTPS, monitoring):  
→ See [DEPLOYMENT.md](./docs/DEPLOYMENT.md)

**Docker image available:** `ghcr.io/MimoJanra/TestOpsMCP:latest`

---

## 🛠️ 104 Tools

### Universal Tools
- **`search_testops_operations`** — Search for any API operation by keyword
  - Example: `"list projects"`, `"create launch"`, `"get test results"`
  - Covers all 600+ TestOps endpoints
- **`execute_testops_operation`** — Execute any discovered operation (supports `body` key for array payloads)
- **`configure_allure_token`** — Set token in chat (if not in config)

### Launch Management
`run_allure_launch`\* • `get_launch_status` • `get_launch_report` • `list_launches` • `get_launch_details` • `close_launch` • `reopen_launch` • `copy_launch`\* • `merge_launches`\* • `get_launch_environment` • `update_launch_environment`

> \* async — returns `task_id` immediately; track with `get_task_status`

### Test Results
`list_test_results` • `get_test_result` • `assign_test_result` • `mute_test_result` • `resolve_test_result` • `unmute_test_result` • `bulk_assign_test_results` • `bulk_mute_test_results` • `bulk_unmute_test_results` • `bulk_resolve_test_results`

### Test Cases — Core
`list_test_cases` • `get_test_case` • `create_test_case` • `update_test_case` • `delete_test_case` • `clone_test_case` • `restore_test_case` • `run_test_case` • `search_test_cases` • `suggest_test_cases` • `validate_test_case_query`

### Test Cases — Steps & Scenario
`create_test_case_step` • `update_test_case_step` • `delete_test_case_step` • `move_test_case_step` • `copy_test_case_step` • `delete_test_case_scenario` • `get_test_case_scenario_from_run` • `detach_test_case_automation`

### Test Cases — Metadata
`get_test_case_tags` • `set_test_case_tags` • `get_test_case_issues` • `set_test_case_issues` • `get_test_case_examples` • `set_test_case_examples` • `get_test_case_custom_fields` • `update_test_case_custom_fields` • `get_test_case_keys` • `set_test_case_keys` • `get_test_case_relations` • `set_test_case_relations` • `get_test_case_workflow`

### Test Cases — Members & Links
`get_test_case_members` • `add_test_case_members` • `remove_test_case_members` • `get_test_case_external_links` • `add_test_case_external_link` • `delete_test_case_external_link`

### Test Cases — Defects
`add_test_case_defect` • `remove_test_case_defect` • `get_test_case_defects` • `get_launch_defects`

### Test Cases — Versions & Attachments
`list_test_case_versions` • `create_test_case_version` • `restore_test_case_version` • `get_test_case_version_data` • `delete_test_case_version` • `get_test_case_attachments` • `delete_test_case_attachment`

### Test Cases — Audit & State
`get_test_case_audit` • `get_test_case_history` • `list_deleted_test_cases` • `list_muted_test_cases`

### Bulk Test Case Operations
`bulk_set_test_case_status` • `bulk_add_test_case_tags` • `bulk_remove_test_case_tags` • `bulk_clone_test_cases` • `bulk_mute_test_cases` • `bulk_delete_test_cases` • `bulk_move_test_cases` • `bulk_set_test_case_layer` • `bulk_add_test_case_members` • `bulk_remove_test_case_members` • `bulk_add_test_case_custom_fields` • `bulk_remove_test_case_custom_fields` • `bulk_add_test_case_external_links` • `bulk_add_test_case_issues` • `bulk_remove_test_case_issues` • `bulk_run_test_cases_new_launch` • `bulk_run_test_cases_existing_launch` • `bulk_create_test_plan`

### Projects & Analytics
`list_projects` • `get_project` • `get_project_stats` • `get_launch_trend_analytics` • `get_launch_duration_analytics` • `get_test_success_rate`

### Async Task Management
`get_task_status` • `list_running_tasks` • `cancel_task`

### AI Analysis
`analyze_launch_failures` — fetches failed tests, asks Claude to identify root causes and suggest fixes (requires client sampling support)

**→ Full reference:** [API.md](./docs/API.md)

---

## 💡 Common Use Cases

### Daily QA Workflow
```
"Run regression tests for Sprint 22"
→ Claude starts launch, tracks progress, reports results
→ Suggests which tests to debug first
```

### Investigate Failures
```
"Why did the auth tests fail?"
→ Claude calls analyze_launch_failures
→ Fetches failures, asks Claude to identify root causes
→ Returns grouped analysis with suggested fixes
```

### Continuous Integration
```
"What's the status of our nightly suite?"
→ Get daily reports, track trends
→ Alerts if pass rate drops
```

### Team Collaboration
```
"Assign the failing UI tests to @Sarah"
→ Claude bulk-assigns, creates defect tickets
→ Notifies team members
```

---

## 🐳 Deployment

### Local Development (Claude Desktop)
Just run the binary — no special setup needed.

### Team Shared Server
Use Docker Compose for easy team deployment:

```bash
docker-compose up -d
```

Or deploy with:
- **Docker** — Container image at `ghcr.io/MimoJanra/TestOpsMCP`
- **Kubernetes** — Enterprise-scale with auto-scaling
- **Systemd** — Native Linux integration
- **ngrok** — Quick HTTPS tunneling for testing

**→ Full guide:** [DEPLOYMENT.md](./docs/DEPLOYMENT.md)

---

## ⚙️ Configuration

| Variable | Required | Default | Notes |
|----------|----------|---------|-------|
| `ALLURE_BASE_URL` | ✅ Yes | — | Your Allure TestOps URL |
| `ALLURE_TOKEN` | ❓ Optional | — | Required for shared servers, optional for per-user |
| `REQUEST_TIMEOUT` | No | 30s | HTTP timeout (1–600 seconds) |
| `PORT` | No | 3000 | Server port (HTTP mode only) |
| `LOG_LEVEL` | No | INFO | DEBUG/INFO/WARN/ERROR |
| `MCP_AUTH_TOKENS` | No | — | Named user tokens: `alice:tok1,bob:tok2` (preferred for teams) |
| `MCP_AUTH_TOKEN` | No | — | Single legacy bearer token; treated as user `"default"` |
| `CORS_ALLOWED_ORIGIN` | No | * | CORS origin (set to `https://claude.ai` for Claude.ai) |
| `AUDIT_LOG_PATH` | No | `audit` | Directory for daily audit JSONL files |
| `AUDIT_RETENTION_DAYS` | No | `30` | Days to keep audit files |

---

## 🔒 Security

### Local Development (Stdio Mode)
- No network exposure
- Safe for personal use
- Default mode for Claude Desktop

### Team Shared Server (HTTP Mode)
**Always set `MCP_AUTH_TOKEN` in production:**
- Use HTTPS (or ngrok) to encrypt credentials
- Set concrete CORS origin (not `*`)
- Place behind reverse proxy (nginx, Caddy)
- Never commit `.env` to git
- Rotate credentials regularly

**Example production setup** (nginx + HTTPS):
→ See [Security notes in DEPLOYMENT.md](./docs/DEPLOYMENT.md#security)

---

## 📚 Documentation

| Guide | For | Time |
|-------|-----|------|
| **[Quick Start](#quick-start-5-minutes)** | First-time users | 2 min |
| **[Installation Options](#installation-options)** | Per-user or shared setup | 5 min |
| **[DEPLOYMENT.md](./docs/DEPLOYMENT.md)** | Production, Docker, K8s, reverse proxy | 15 min |
| **[API.md](./docs/API.md)** | Tool reference & examples | Reference |

---

## 🔄 Comparison: TestOps MCP vs Alternatives

| Feature | **TestOps MCP** | Dashboard | Manual API |
|---------|---|---|---|
| Launch from chat | ✅ | ❌ | ❌ |
| Real-time updates | ✅ | ⚠️ Manual | ⚠️ Polling |
| AI analysis | ✅ Claude | ❌ | ❌ |
| Full API coverage | ✅ 600+ endpoints | ⚠️ Limited | ✅ |
| Team deployment | ✅ Built-in | ❌ | ⚠️ DIY |
| Production-ready | ✅ | ⚠️ | ⚠️ |
| **Time to results** | **3 min** | 15+ min | 10+ min |
| **Context switches** | **0** | 4+ | 2+ |

---

## 🔧 Build from Source

```bash
# Clone
git clone https://github.com/MimoJanra/TestOpsMCP.git
cd TestOpsMCP

# Build
make build

# Binary at: ./bin/server.exe (Windows) or ./bin/server (Unix)
```

**Requirements:**
- Go 1.26+ — [Download](https://golang.org/dl/)
- `make` — Usually pre-installed on macOS/Linux

**Development commands:**
```bash
make build      # Compile
make run        # Run in stdio mode
make run-http   # Run in HTTP mode on :3000
make test       # Run tests
make lint       # Check quality
make check      # test + lint
```

---

## 🤝 Contributing

**Love this project?** Help us make it better!

1. Fork the repo
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make changes and test: `make check`
4. Push and open a PR

**Ways to contribute:**
- 🐛 Fix bugs
- ✨ Add features
- 📚 Improve docs
- 🤔 Report issues
- 💡 Suggest improvements

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

---

## Support & Community

**Questions?** [Open an issue](https://github.com/MimoJanra/TestOpsMCP/issues) or [discussion](https://github.com/MimoJanra/TestOpsMCP/discussions)

**Found a bug?** [Report it](https://github.com/MimoJanra/TestOpsMCP/issues/new?template=bug_report.md)

**Have an idea?** [Share it](https://github.com/MimoJanra/TestOpsMCP/issues/new?template=feature_request.md)

**Security issue?** Email: `security@testopsmcp.dev` (don't open public issue)

---

## Related Projects

- [Model Context Protocol Spec](https://spec.modelcontextprotocol.io/)
- [Claude Desktop Docs](https://claude.ai/docs)
- [Allure TestOps](https://qameta.io/allure-testops/)
- [MCP Ecosystem](https://modelcontextprotocol.io/servers)

---

## License

[Apache License 2.0](LICENSE)

---

## Get Started

<table>
<tr>
<td width="50%">

### ⚡ First Time?
[Download binary →](https://github.com/MimoJanra/TestOpsMCP/releases/latest)
2 minutes to running.

</td>
<td width="50%">

### 🚀 Ready to Deploy?
[See DEPLOYMENT.md →](./docs/DEPLOYMENT.md)
Docker, K8s, and more.

</td>
</tr>
</table>

---

<div align="center">

**Made with ❤️ by [Artem Alekseev](https://github.com/MimoJanra)**

Apache License 2.0 — Free to use, modify, and distribute.

</div>
