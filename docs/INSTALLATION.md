# Installation Guide

Complete guide to install and configure Allure MCP Server for different environments.

## Table of Contents

- [Local Development](#local-development)
- [Docker Setup](#docker-setup)
- [Team Deployment](#team-deployment)
- [Troubleshooting](#troubleshooting)

## Local Development

### Prerequisites

- **Go 1.26+** — [Download from golang.org](https://golang.org/dl/)
- **Make** — Usually pre-installed on Unix-like systems
- Access to **Allure TestOps instance** with API token
- **Claude Desktop** (for testing MCP integration)

### Step 1: Get Your Allure API Token

1. Log in to Allure TestOps
2. Go to **Settings** > **Integrations** > **API tokens**
3. Create a new token (or use an existing one)
4. Copy the token value

### Step 2: Clone and Build

```bash
git clone https://github.com/MimoJanra/TestOpsMCP.git
cd TestOpsMCP
make build
```

The binary will be at `bin/server.exe` (or `bin/server` on Unix).

### Step 3: Create Environment File

Copy `.env.example` to `.env` and fill in your credentials:

```bash
cp .env.example .env
```

Edit `.env`:

```env
ALLURE_BASE_URL=https://your-allure.com
ALLURE_TOKEN=your_token_here
LOG_LEVEL=DEBUG  # Optional, for development
```

**⚠️ Never commit `.env` to git** — it contains secrets!

### Step 4: Run Locally

**Stdio mode** (for Claude Desktop):

```bash
make run
```

Or directly:

```bash
./bin/server.exe
```

Then restart Claude Desktop to pick up the MCP server in the dropdown.

**HTTP mode** (for testing):

```bash
make run-http
```

Visit `http://localhost:3000/health` to verify it's running (returns `ok`).

### Step 5: Configure Claude Desktop

Edit your Claude Desktop config:

**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

**Mac:** `~/Library/Application Support/Claude/claude_desktop_config.json`

**Linux:** `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "allure": {
      "command": "C:\\Users\\YourUsername\\GolandProjects\\TestOpsMCP\\bin\\server.exe",
      "env": {
        "ALLURE_BASE_URL": "https://your-allure.com",
        "ALLURE_TOKEN": "your_token_here"
      }
    }
  }
}
```

Restart Claude Desktop. The Allure tools should now appear in the tool dropdown.

### Step 6 (Alternative): Configure Claude Code CLI

If you use **Claude Code** in the terminal or IDE extensions instead of Claude Desktop, register the server with `claude mcp add`:

**Stdio mode** — runs the binary directly (no HTTP server needed):

```bash
# macOS / Linux
claude mcp add -s user testops ./bin/server \
  -e ALLURE_BASE_URL=https://your-allure.com \
  -e ALLURE_TOKEN=your_token_here

# Windows (PowerShell)
claude mcp add -s user testops .\bin\server.exe `
  -e ALLURE_BASE_URL=https://your-allure.com `
  -e ALLURE_TOKEN=your_token_here
```

**HTTP mode** — connect to a running server (`make run-http` or Docker):

```bash
claude mcp add -s user --transport http testops http://localhost:3000/mcp \
  --header "Authorization:Bearer your-mcp-auth-token" \
  --header "X-Allure-Token:your-personal-allure-token"
```

**Scope flags explained:**

| Flag | Where it's saved | Shared with team? |
|------|-----------------|-------------------|
| `-s user` | `~/.claude.json` | No |
| `-s local` | `.claude/settings.local.json` | No (gitignored) |
| `-s project` | `.mcp.json` in repo root | Yes (committed) |

Use `-s user` for personal setups. Use `-s project` to share the config via git (don't commit secrets — use `-e` env vars or a `.env` file).

Verify registration:

```bash
claude mcp list
# testops: /path/to/server.exe [user]
```

Remove if needed:

```bash
claude mcp remove testops
```

---

## Docker Setup

### Prerequisites

- Docker 20.10+ or Docker Desktop
- Docker Compose 1.29+ (if using `docker-compose.yml`)

### Option 1: Pre-built Image from GitHub Container Registry (Recommended)

No need to clone the repo — pull the ready-to-use image:

```bash
docker pull ghcr.io/MimoJanra/TestOpsMCP:latest
```

Run the container:

```bash
docker run \
  -e ALLURE_BASE_URL=https://your-allure.com \
  -e ALLURE_TOKEN=your_token_here \
  -e MCP_AUTH_TOKENS=alice:token-for-alice,bob:token-for-bob \
  -e LOG_LEVEL=INFO \
  -p 3000:3000 \
  ghcr.io/MimoJanra/TestOpsMCP:latest --http
```

Or pin a specific version:

```bash
docker run ... ghcr.io/MimoJanra/TestOpsMCP:v2.0.3 --http
```

### Option 2: Docker Compose (Recommended for Teams)

Create a `docker-compose.yml` (no need to clone the repo):

```yaml
services:
  testops-mcp:
    image: ghcr.io/MimoJanra/TestOpsMCP:latest
    restart: unless-stopped
    command: ./testops-mcp --http
    ports:
      - "3000:3000"
    env_file:
      - .env
```

Create `.env`:

```env
ALLURE_BASE_URL=https://your-allure.com
ALLURE_TOKEN=your_token_here
MCP_AUTH_TOKENS=alice:token-for-alice,bob:token-for-bob
LOG_LEVEL=INFO
```

Start the service:

```bash
docker-compose up -d
```

**View logs:**

```bash
docker-compose logs -f
```

**Stop the service:**

```bash
docker-compose down
```

### Option 3: Build from Source

Only needed if you want to modify the code:

```bash
git clone https://github.com/MimoJanra/TestOpsMCP.git
cd TestOpsMCP
docker build -t testops-mcp:local .
docker run -e ALLURE_BASE_URL=https://your-allure.com -e ALLURE_TOKEN=your_token_here \
  -p 3000:3000 testops-mcp:local --http
```

---

## Team Deployment

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: allure-mcp
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: allure-mcp
  template:
    metadata:
      labels:
        app: allure-mcp
    spec:
      containers:
      - name: allure-mcp
        image: ghcr.io/MimoJanra/TestOpsMCP:latest
        ports:
        - containerPort: 3000
        env:
        - name: ALLURE_BASE_URL
          valueFrom:
            secretKeyRef:
              name: allure-secrets
              key: base-url
        - name: ALLURE_TOKEN
          valueFrom:
            secretKeyRef:
              name: allure-secrets
              key: token
        - name: MCP_AUTH_TOKENS
          valueFrom:
            secretKeyRef:
              name: mcp-secrets
              key: auth-tokens
        livenessProbe:
          httpGet:
            path: /health
            port: 3000
          initialDelaySeconds: 10
          periodSeconds: 30
---
apiVersion: v1
kind: Service
metadata:
  name: allure-mcp
  namespace: default
spec:
  selector:
    app: allure-mcp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: LoadBalancer
```

### Systemd Service (Linux)

Create `/etc/systemd/system/allure-mcp.service`:

```ini
[Unit]
Description=Allure MCP Server
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=allure-mcp
Group=allure-mcp
WorkingDirectory=/opt/allure-mcp
ExecStart=/opt/allure-mcp/testops-mcp --http
Restart=on-failure
RestartSec=10

# Environment variables
EnvironmentFile=/opt/allure-mcp/.env

# Resource limits
MemoryLimit=256M
CPUQuota=100%

# Logging to journalctl
StandardOutput=journal
StandardError=journal
SyslogIdentifier=allure-mcp

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable allure-mcp
sudo systemctl start allure-mcp
sudo journalctl -u allure-mcp -f
```

---

## Troubleshooting

### "Connection refused" error

**Problem:** Server can't connect to Allure TestOps

**Solutions:**

1. Check `ALLURE_BASE_URL` is correct (with `https://` or `http://`)
2. Verify Allure is reachable: `curl https://your-allure.com`
3. Check firewall / proxy settings
4. Verify `ALLURE_TOKEN` is valid

### "Unauthorized" errors

**Problem:** 401/403 responses from Allure

**Solution:** Generate a new API token from Allure UI and update `ALLURE_TOKEN`

### Docker container exits immediately

**Problem:** Container crashes on startup

**Solution:** Check logs:

```bash
docker-compose logs allure-mcp
```

Common causes:
- Missing `ALLURE_BASE_URL` or `ALLURE_TOKEN`
- Port 3000 already in use: `docker-compose.yml` has `ports: ["3000:3000"]`

### High memory usage

**Problem:** Container using excessive memory

**Solution:** Add memory limit in `docker-compose.yml`:

```yaml
deploy:
  resources:
    limits:
      memory: 256M
```

### Cannot find compiled binary

**Problem:** `make run` fails with "file not found"

**Solution:** Ensure Go 1.26+ is installed:

```bash
go version
# Should show go version go1.26 or later
```

Then rebuild:

```bash
make clean
make build
```

---

## Next Steps

- [See Deployment Guide](./DEPLOYMENT.md) for production setup
- [Check API Reference](./API.md) for tool usage
- [Review Security Notes](../README.md#-security) in main README
