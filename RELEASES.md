# Release Downloads & Setup

All pre-built binaries are provided for easy setup without requiring Go installation.

## 🆕 What's New in v2.0.0

- **MCP protocol 2025-11-25** with version negotiation and stable cursor-based pagination
- **114 tools** (+ `find_project`, `remove_test_cases_from_launch`, `get_task_status`, `list_running_tasks`, `cancel_task`, `analyze_launch_failures`)
- **Async task system** — `run_allure_launch`, `copy_launch`, `merge_launches`, bulk runs return `task_id` immediately; tasks auto-expire after 1 hour
- **AI failure analysis** — `analyze_launch_failures` asks Claude to identify root causes via MCP sampling
- **Confirmation dialogs** — `delete_test_case` / `bulk_delete_test_cases` require user confirmation via elicitation (requires HTTP transport; stdio returns an error)
- **Resource subscriptions** — subscribe to `ui://widgets/launch-dashboard?launch_id=N` for live push notifications on status change
- **Argument completion** — `completion/complete` returns live `project_id` and `launch_id` suggestions (30 s cache, no API spam)
- **Panic recovery** — both sync handlers and async goroutines are covered; panics mark the task Failed instead of crashing the server
- **Standard elicitation/sampling** — uses proper JSON-RPC request/response pattern; compatible with Claude Desktop and all spec-compliant clients
- **slog internals** — logger now backed by `log/slog` with JSON output

## 📥 Available Binaries

| Platform | Download | Architecture |
|----------|----------|---------------|
| **Windows 64-bit** | `testops-mcp-windows-amd64.exe` | x86-64 |
| **Windows ARM** | `testops-mcp-windows-arm64.exe` | ARM64 |
| **macOS Intel** | `testops-mcp-macos-amd64` | x86-64 (Intel) |
| **macOS Apple Silicon** | `testops-mcp-macos-arm64` | ARM64 (M1/M2/M3) |
| **Linux 64-bit** | `testops-mcp-linux-amd64` | x86-64 |
| **Linux ARM** | `testops-mcp-linux-arm64` | ARM64 |

## 🚀 Setup Instructions

### Windows

1. Download `testops-mcp-windows-amd64.exe` (or `arm64` for ARM devices)
2. Create a folder (e.g., `C:\testops`)
3. Move the binary there
4. Create `.env` file in the same folder:
   ```
   ALLURE_BASE_URL=https://your-testops.com
   ALLURE_TOKEN=your-token-here
   ```
5. Open PowerShell and run:
   ```powershell
   cd C:\testops
   .\testops-mcp-windows-amd64.exe --http
   ```

### macOS (Intel & Apple Silicon)

1. Download appropriate binary:
   - **Intel (x86-64)**: `testops-mcp-macos-amd64`
   - **Apple Silicon (M1/M2/M3)**: `testops-mcp-macos-arm64`

2. Create setup directory:
   ```bash
   mkdir ~/testops
   cd ~/testops
   mv ~/Downloads/testops-mcp-macos-* .
   chmod +x testops-mcp-macos-*
   ```

3. Create `.env` file:
   ```bash
   cat > .env << EOF
   ALLURE_BASE_URL=https://your-testops.com
   ALLURE_TOKEN=your-token-here
   EOF
   ```

4. Run:
   ```bash
   # Intel
   ./testops-mcp-macos-amd64 --http
   
   # Apple Silicon
   ./testops-mcp-macos-arm64 --http
   ```

### Linux

1. Download `testops-mcp-linux-amd64` (or `arm64` for ARM servers)
2. Setup directory:
   ```bash
   mkdir ~/testops
   cd ~/testops
   mv ~/Downloads/testops-mcp-linux-amd64 .
   chmod +x testops-mcp-linux-amd64
   ```

3. Create `.env`:
   ```bash
   cat > .env << EOF
   ALLURE_BASE_URL=https://your-testops.com
   ALLURE_TOKEN=your-token-here
   EOF
   ```

4. Run:
   ```bash
   ./testops-mcp-linux-amd64 --http
   ```

## 🔧 Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ALLURE_BASE_URL` | Yes | Base URL of your Allure TestOps instance (e.g., https://testops.example.com) |
| `ALLURE_TOKEN` | Yes | API token from Allure TestOps (Settings → Integrations → API Token) |

### Running Modes

**Stdio Mode** (Claude Desktop):
```bash
./testops-mcp-macos-amd64
```
Used for Claude Desktop integration via MCP configuration.

**HTTP Mode** (Team/Server):
```bash
./testops-mcp-macos-amd64 --http
```
Runs on `http://localhost:3000` for team deployments.

## 🔧 Additional Configuration (v2.0+)

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_AUTH_TOKENS` | — | Named user tokens: `alice:tok1,bob:tok2` |
| `MCP_AUTH_TOKEN` | — | Single legacy token (user `"default"`) |
| `AUDIT_LOG_PATH` | `audit` | Directory for daily audit JSONL files |
| `AUDIT_RETENTION_DAYS` | `30` | Days to keep audit files |

## 📋 Getting Your API Token

1. Log in to Allure TestOps
2. Go to **Settings** (⚙️ icon)
3. Navigate to **Integrations**
4. Find or create an API token
5. Copy the token
6. Set `ALLURE_TOKEN=your-token-here`

## ✅ Verify Installation

Test the connection:
```bash
curl http://localhost:3000/health
```

Expected response: Connection successful message

## 🐳 Docker Deployment (Optional)

For team/server deployments:

```bash
docker pull ghcr.io/MimoJanra/TestOpsMCP:latest

docker run -d \
  -e ALLURE_BASE_URL=https://your-testops.com \
  -e ALLURE_TOKEN=your-token \
  -p 3000:3000 \
  ghcr.io/MimoJanra/TestOpsMCP:latest --http
```

## 🆘 Troubleshooting

**"Permission denied" on macOS/Linux:**
```bash
chmod +x testops-darwin-amd64
# or
chmod +x testops-linux-amd64
```

**"Cannot execute binary" on macOS:**
May need to allow app in Security settings:
```bash
xattr -d com.apple.quarantine ./testops-darwin-amd64
```

**Connection refused:**
- Verify `ALLURE_BASE_URL` is correct
- Check `ALLURE_TOKEN` is valid
- Ensure Allure TestOps instance is accessible

## 📚 Available Tools

Over **114 MCP tools** for Allure TestOps integration:

- Launch management (create, run\*, close, reopen, copy\*, merge\*) — \*async
- Test case operations (create, update, delete†, clone, restore, versions, audit) — †with confirmation
- Test result handling (assign, mute, resolve, unmute, bulk operations)
- Custom fields, tags, issues, members, external links
- Relations and integration keys (Jira, Azure DevOps, etc.)
- Bulk operations (mass clone\*, bulk updates, bulk delete†, bulk run\*) — \*async, †with confirmation
- Analytics and reporting
- Async task management (`get_task_status`, `list_running_tasks`, `cancel_task`)
- AI failure analysis (`analyze_launch_failures` via MCP sampling)
- Interactive widgets (Launch Dashboard with live subscriptions, Action Picker, Results Display)
- Full OpenAPI coverage via search + execute pattern

See [README.md](README.md) for complete tool documentation.

## 📞 Support

For issues or questions:
1. Check Allure TestOps connection settings
2. Verify API token hasn't expired
3. Review logs in console output
4. Open an issue on GitHub with error details
