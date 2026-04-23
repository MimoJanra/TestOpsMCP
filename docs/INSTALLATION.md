# Installation Guide

Complete guide to install and configure Allure MCP Server for different environments.

## Table of Contents

- [Local Development](#local-development)
- [Docker Setup](#docker-setup)
- [Team Deployment](#team-deployment)
- [Troubleshooting](#troubleshooting)

## Local Development

### Prerequisites

- **Go 1.22+** — [Download from golang.org](https://golang.org/dl/)
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

Visit `http://localhost:3000/sse` to verify it's running.

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

---

## Docker Setup

### Prerequisites

- Docker 20.10+ or Docker Desktop
- Docker Compose 1.29+ (if using `docker-compose.yml`)

### Option 1: Docker CLI

Build the image:

```bash
docker build -t allure-mcp .
```

Run the container:

```bash
docker run \
  -e ALLURE_BASE_URL=https://your-allure.com \
  -e ALLURE_TOKEN=your_token_here \
  -e MCP_AUTH_TOKEN=your_shared_secret \
  -e LOG_LEVEL=INFO \
  -p 3000:3000 \
  allure-mcp
```

### Option 2: Docker Compose (Recommended)

Create `.env` file (copy from `.env.example`):

```bash
cp .env.example .env
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

### Custom Image Tags

For registry / team deployments:

```bash
docker build -t registry.example.com/allure-mcp:1.0 .
docker push registry.example.com/allure-mcp:1.0
```

Then reference in `docker-compose.yml`:

```yaml
services:
  allure-mcp:
    image: registry.example.com/allure-mcp:1.0
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
        image: allure-mcp:latest
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
        - name: MCP_AUTH_TOKEN
          valueFrom:
            secretKeyRef:
              name: mcp-secrets
              key: auth-token
        livenessProbe:
          httpGet:
            path: /sse
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
ExecStart=/opt/allure-mcp/bin/server --http
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

**Solution:** Ensure Go 1.22+ is installed:

```bash
go version
# Should show go version go1.22 or later
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
- [Review Security Notes](../README.md#security-notes) in main README
