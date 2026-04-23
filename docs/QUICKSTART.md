# Quick Start Guide

Get Allure MCP Server running in 5 minutes.

## Choose Your Path

### 🖥️ Local Development (Claude Desktop)

Perfect for: Testing locally, development, single user

```bash
# 1. Clone (or already have the repo)
cd TestOpsMCP

# 2. Configure
cp .env.example .env
# Edit .env with your Allure credentials:
# ALLURE_BASE_URL=https://your-allure.com
# ALLURE_TOKEN=your_token_here

# 3. Build & run
make build
make run

# 4. Configure Claude Desktop
# Edit %APPDATA%\Claude\claude_desktop_config.json:
{
  "mcpServers": {
    "allure": {
      "command": "C:\\Users\\YourName\\TestOpsMCP\\bin\\server.exe",
      "env": {
        "ALLURE_BASE_URL": "https://your-allure.com",
        "ALLURE_TOKEN": "your_token_here"
      }
    }
  }
}

# 5. Restart Claude Desktop
# Tools appear in dropdown ✓
```

**Time:** ~5 minutes  
**Requirements:** Go 1.22+, Claude Desktop  
**See also:** [Installation Guide](./INSTALLATION.md#local-development)

---

### 🐳 Docker (Single Instance)

Perfect for: Testing, single server deployment, CI/CD

```bash
# 1. Build image
docker build -t allure-mcp .

# 2. Run with environment variables
docker run -d \
  -e ALLURE_BASE_URL=https://your-allure.com \
  -e ALLURE_TOKEN=your_token_here \
  -e MCP_AUTH_TOKEN=your_shared_secret \
  -p 3000:3000 \
  allure-mcp --http

# 3. Test
curl -H "Authorization: Bearer your_shared_secret" \
     http://localhost:3000/sse

# 4. Configure Claude Desktop (HTTP mode)
{
  "mcpServers": {
    "allure": {
      "url": "http://localhost:3000",
      "env": {
        "MCP_AUTH_TOKEN": "your_shared_secret"
      }
    }
  }
}
```

**Time:** ~2 minutes  
**Requirements:** Docker  
**See also:** [Deployment Guide](./DEPLOYMENT.md#docker-deployment)

---

### 🐳 Docker Compose (Team)

Perfect for: Team deployment, easy configuration, multiple services

```bash
# 1. Create .env file
cp .env.example .env
# Edit .env with your Allure credentials and team settings

# 2. Start services
docker-compose up -d

# 3. Check logs
docker-compose logs -f allure-mcp

# 4. Stop (when done)
docker-compose down

# 5. Configure Claude Desktop (team members)
{
  "mcpServers": {
    "allure": {
      "url": "http://your-server:3000",
      "env": {
        "MCP_AUTH_TOKEN": "your_shared_secret_from_.env"
      }
    }
  }
}
```

**Time:** ~3 minutes  
**Requirements:** Docker Compose  
**See also:** [Deployment Guide](./DEPLOYMENT.md#docker-deployment)

---

### ☁️ Kubernetes (Production)

Perfect for: Large teams, auto-scaling, production

```bash
# 1. Create namespace and secrets
kubectl create namespace allure-mcp

# Edit k8s-manifest.yaml with your credentials

# 2. Deploy
kubectl apply -f k8s-manifest.yaml

# 3. Check status
kubectl get all -n allure-mcp
kubectl logs -f deployment/allure-mcp -n allure-mcp

# 4. Get service URL
kubectl get svc -n allure-mcp

# 5. Configure Claude Desktop
{
  "mcpServers": {
    "allure": {
      "url": "http://your-k8s-service:3000",
      "env": {
        "MCP_AUTH_TOKEN": "your_shared_secret"
      }
    }
  }
}
```

**Time:** ~10 minutes  
**Requirements:** Kubernetes cluster  
**See also:** [Deployment Guide](./DEPLOYMENT.md#kubernetes)

---

## Using Allure MCP in Claude

Once configured, ask Claude:

### Launch Tests

> "Run the smoke tests for project 1 in Allure"

Claude automatically uses `run_allure_launch` and returns the launch ID.

### Check Status

> "What's the status of launch 12345?"

Claude uses `get_launch_status` and reports progress.

### Get Results

> "Show me the test report for launch 12345"

Claude uses `get_launch_report` and displays pass/fail statistics.

---

## Troubleshooting

### "Connection refused" in Docker

```bash
# Check if container is running
docker-compose ps

# View logs
docker-compose logs allure-mcp

# Likely cause: Invalid ALLURE_BASE_URL or ALLURE_TOKEN
# Fix: Verify .env file and restart
docker-compose restart allure-mcp
```

### "Cannot find binary" in local mode

```bash
# Ensure Go 1.22+ is installed
go version

# Rebuild
make clean
make build

# Run
make run
```

### Claude Desktop tools don't appear

1. Check config file syntax (JSON must be valid)
2. Ensure command path is correct and binary exists
3. Restart Claude Desktop completely (not just refresh)
4. Check server logs: `make run` or `docker-compose logs -f`

### "Unauthorized" errors from Allure

1. Verify `ALLURE_BASE_URL` is reachable: `curl -v https://your-allure.com`
2. Generate new API token in Allure UI
3. Update `ALLURE_TOKEN` in `.env`
4. Restart server

---

## Common Commands

### Local Development

```bash
make build      # Compile
make run        # Run stdio mode (Claude Desktop)
make run-http   # Run HTTP mode (test manually)
make test       # Run tests
make lint       # Check code quality
make clean      # Remove artifacts
```

### Docker

```bash
docker build -t allure-mcp .                    # Build image
docker run allure-mcp --http                   # Run HTTP mode
docker-compose up -d                           # Start services
docker-compose logs -f allure-mcp              # View logs
docker-compose down                            # Stop services
```

### Kubernetes

```bash
kubectl apply -f k8s-manifest.yaml              # Deploy
kubectl get all -n allure-mcp                  # Check status
kubectl logs -f deployment/allure-mcp          # View logs
kubectl delete namespace allure-mcp            # Remove
```

---

## Next Steps

- **[Installation Guide](./INSTALLATION.md)** — Detailed setup for each method
- **[Deployment Guide](./DEPLOYMENT.md)** — Production patterns and best practices
- **[API Reference](./API.md)** — Complete tool documentation
- **[Security Guide](./SECURITY.md)** — Security best practices
- **[Contributing](../CONTRIBUTING.md)** — How to contribute improvements

---

## Need Help?

- 📖 Check [docs/](.) for detailed guides
- 🐛 [Report bugs on GitHub](https://github.com/MimoJanra/TestOpsMCP/issues)
- 💬 [Start a GitHub Discussion](https://github.com/MimoJanra/TestOpsMCP/discussions)
- 📧 Email: alk@tassta.com

**Happy testing! 🚀**
