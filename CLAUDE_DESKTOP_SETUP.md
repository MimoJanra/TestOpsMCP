# Claude Desktop Setup Guide

This guide shows how to connect Claude Desktop to the Allure MCP Server.

## Prerequisites

- Claude Desktop installed ([download](https://claude.ai/download))
- Allure MCP Server built: `go build -o bin/server ./cmd/server`
- Allure TestOps credentials (base URL and API token)

## Local Setup (stdio mode)

### Step 1: Create `.env` file

In the server directory, copy and fill in your credentials:

```bash
cp .env.example .env
```

Edit `.env`:

```env
ALLURE_BASE_URL=https://tassta.testops.cloud
ALLURE_TOKEN=your-allure-api-token-here
REQUEST_TIMEOUT=30
LOG_LEVEL=INFO
```

**Do not commit `.env` to git.**

### Step 2: Find your config file location

**macOS:**
```bash
~/Library/Application Support/Claude/claude_desktop_config.json
```

**Windows:**
```
%APPDATA%\Claude\claude_desktop_config.json
```

**Linux:**
```bash
~/.config/Claude/claude_desktop_config.json
```

### Step 3: Add MCP server to config

Open `claude_desktop_config.json` and add (or update) the `mcpServers` section:

```json
{
  "mcpServers": {
    "allure": {
      "command": "/absolute/path/to/bin/server",
      "env": {
        "ALLURE_BASE_URL": "https://tassta.testops.cloud",
        "ALLURE_TOKEN": "your-allure-api-token-here",
        "LOG_LEVEL": "INFO"
      }
    }
  }
}
```

**Important notes:**
- Use **absolute path** to the server binary (not relative)
- On Windows: use `C:\\Users\\...\\bin\\server` or `C:/Users/.../bin/server`
- On macOS/Linux: use `/Users/...` or `~/path/to/bin/server`
- The `.env` file in the server directory is ignored when env vars are passed in config

### Step 4: Restart Claude Desktop

Close and reopen Claude Desktop completely. You should see:
- A "MCP" indicator in the chat UI
- A dropdown showing available tools

### Step 5: Use the tools

In a chat, ask Claude to use the Allure tools:

```
Create a test launch in Allure for project 1 with name "Smoke Tests"
```

Claude will:
1. Call `run_allure_launch` with `project_id=1, launch_name="Smoke Tests"`
2. Show you the result (launch ID, status)

You can also ask for:
- Launch status: _"What's the status of launch 12345?"_
- Test report: _"Get the report for launch 12345"_

## Shared Server Setup (HTTP mode)

If your team has a **shared MCP server** running on a server, follow these steps instead.

### Step 1: Get the server URL and auth token from your admin

Your admin should provide:
- Server URL (e.g., `http://mcp-server.company.com:3000` or `https://mcp.company.com`)
- Shared auth token (if required)

### Step 2: Add to config

```json
{
  "mcpServers": {
    "allure": {
      "url": "http://mcp-server.company.com:3000",
      "env": {
        "MCP_AUTH_TOKEN": "shared-auth-token-from-admin"
      }
    }
  }
}
```

**Notes:**
- Use `"url"` instead of `"command"` for HTTP mode
- Include `MCP_AUTH_TOKEN` in env if your admin set one
- If the server uses HTTPS with self-signed cert, you may need to add: `"skipCertificateVerification": true`

### Step 3: Restart Claude Desktop

Same as local setup — close and reopen Claude Desktop.

## Troubleshooting

### Tools not appearing in Claude Desktop

**Check the MCP indicator:**
- If red/disconnected: server process crashed or failed to start
- If yellow/loading: server is initializing
- If green: connected

**View logs:**

**macOS/Linux:**
```bash
# Follow Claude Desktop logs
tail -f ~/Library/Logs/Claude/mcp-server.log
```

**Windows:**
```powershell
# Check Event Viewer or look for logs in:
dir "$env:APPDATA\Claude\logs\"
```

**Manual test (stdio mode):**

```bash
# Start the server manually
export ALLURE_BASE_URL=https://...
export ALLURE_TOKEN=...
./bin/server

# In another terminal, send a test request
cat <<'EOF' | ./bin/server
{"jsonrpc":"2.0","id":1,"method":"tools/list"}
EOF
```

You should get back a JSON response listing the three tools.

### "Method not allowed" or "Unknown session"

This means Claude Desktop tried to use HTTP instead of stdio. Check:
1. Config has `"command"` (not `"url"`) for local setup
2. Path to server binary is correct and executable
3. File is not corrupted: `file /path/to/bin/server`

### Auth token errors (shared server)

If you get `"unauthorized"` responses:
1. Confirm `MCP_AUTH_TOKEN` matches what admin provided
2. Ask admin to check server logs: `journalctl -u allure-mcp -f`
3. Check server URL is correct and accessible

### Server crashes immediately

Check the `.env` or config env vars:
- `ALLURE_BASE_URL` must be a valid URL (starts with `http://` or `https://`)
- `ALLURE_TOKEN` must not be empty
- `REQUEST_TIMEOUT` must be a number between 1 and 600

### High latency / slow responses

Check:
- Network latency to Allure TestOps server
- Set `REQUEST_TIMEOUT` higher (default 30s) if Allure is slow
- Check server CPU/memory: `top` (macOS/Linux) or Task Manager (Windows)

## Example: Step-by-step chat flow

```
User: Create a test launch called "API Tests" in project 2

Claude: I'll create that launch for you.
[calls run_allure_launch with project_id=2, launch_name="API Tests"]

Claude: I've created the launch. Here's the result:
{
  "launch_id": 54321,
  "status": "started"
}

User: Get me the status of that launch

Claude: Let me check the status of launch 54321.
[calls get_launch_status with launch_id=54321]

Claude: The launch is currently RUNNING.

User: Now get the report

Claude: I'll fetch the report for launch 54321.
[calls get_launch_report with launch_id=54321]

Claude: Here's the test execution report:
{
  "total": 42,
  "passed": 38,
  "failed": 3,
  "broken": 1
}
```

## Advanced: Custom env per developer

Each developer on your team can have different local configs:

```json
{
  "mcpServers": {
    "allure": {
      "command": "/Users/alice/src/TestOpsMCP/bin/server",
      "env": {
        "ALLURE_BASE_URL": "https://allure.company.com",
        "ALLURE_TOKEN": "alice-personal-token",
        "LOG_LEVEL": "DEBUG"
      }
    }
  }
}
```

This way each person's token is only in their local config, never in git.

## Support

- **Server logs:** Check `LOG_LEVEL=DEBUG` to see request/response details
- **Server source:** [github.com/MimoJanra/TestOpsMCP](https://github.com/MimoJanra/TestOpsMCP)
- **MCP spec:** [spec.modelcontextprotocol.io](https://spec.modelcontextprotocol.io/)
