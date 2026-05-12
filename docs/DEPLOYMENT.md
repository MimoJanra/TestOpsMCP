# Shared Server & DevOps Deployment Guide

Complete guide for deploying TestOps MCP as a **shared server** for your team.

**This is for Option 2 (Shared Server Setup).**  
For per-user setup, see [README.md](../README.md#per-user-setup-recommended)

## Table of Contents

- [Quick Start (Docker Compose)](#quick-start-docker-compose)
- [Docker Container Registry](#docker-container-registry)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Nginx Reverse Proxy + HTTPS](#nginx-reverse-proxy--https)
- [Systemd Service (Linux)](#systemd-service-linux)
- [ngrok Tunneling (Testing)](#ngrok-tunneling-testing)
- [Security Best Practices](#security-best-practices)
- [Monitoring & Logging](#monitoring--logging)
- [Troubleshooting](#troubleshooting)

## Quick Start (Docker Compose)

**Fastest way to get started.**

### Prerequisites

- Docker & Docker Compose
- One Allure TestOps API token (shared service account recommended)
- Your Allure TestOps URL

### Setup

```bash
# 1. Clone repo
git clone https://github.com/MimoJanra/TestOpsMCP.git
cd TestOpsMCP

# 2. Create .env file
cat > .env << 'EOF'
ALLURE_BASE_URL=https://your-testops.com
ALLURE_TOKEN=your-shared-token
PORT=3000
LOG_LEVEL=INFO
MCP_AUTH_TOKEN=generate-a-random-string-here
CORS_ALLOWED_ORIGIN=https://claude.ai
EOF

# 3. Start
docker-compose up -d

# 4. Check logs
docker-compose logs -f
```

Server runs on `http://localhost:3000`

### Share with Team

Add to Claude Desktop config:

```json
{
  "mcpServers": {
    "testops": {
      "command": "http://your-server.com:3000"
    }
  }
}
```

---

## Docker Deployment

### Build for Production

```bash
# Build with specific tag
docker build -t testops-mcp:1.0 .

# Push to registry
docker push myregistry.com/testops-mcp:1.0
```

### Run with Docker Compose (Production)

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  testops-mcp:
    image: ghcr.io/MimoJanra/TestOpsMCP:latest
    restart: unless-stopped
    command: ./testops-mcp --http
    ports:
      - "3000:3000"
    env_file:
      - .env.prod
    environment:
      LOG_LEVEL: INFO
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 256M
        reservations:
          cpus: '0.5'
          memory: 128M
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/sse"]
      interval: 30s
      timeout: 10s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "10"
```

Run:

```bash
docker-compose -f docker-compose.prod.yml up -d
docker-compose -f docker-compose.prod.yml logs -f
```

---

## Reverse Proxy Setup

### Nginx

Create `/etc/nginx/sites-available/allure-mcp`:

```nginx
upstream testops_mcp_backend {
    server localhost:3000;
}

server {
    listen 443 ssl http2;
    server_name testops.company.com;

    # SSL certificates
    ssl_certificate /etc/letsencrypt/live/testops.company.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/testops.company.com/privkey.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req zone=api_limit burst=20 nodelay;

    # SSE endpoint
    location /sse {
        proxy_pass http://testops_mcp_backend;
        proxy_http_version 1.1;
        
        # SSE requires these headers
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $server_name;
        
        # Pass auth header
        proxy_pass_request_headers on;
        
        # Disable buffering for SSE
        proxy_buffering off;
        
        # Timeouts for long-lived connection
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }

    # Messages endpoint
    location /messages {
        proxy_pass http://testops_mcp_backend;
        proxy_http_version 1.1;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $server_name;
        proxy_pass_request_headers on;
        
        # Timeout for message processing
        proxy_read_timeout 60s;
        proxy_connect_timeout 10s;
    }

    # Deny all other paths
    location / {
        return 404;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name testops.company.com;
    return 301 https://$server_name$request_uri;
}
```

Enable and reload:

```bash
sudo ln -s /etc/nginx/sites-available/allure-mcp /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Caddy

Create `Caddyfile`:

```caddy
allure-mcp.example.com {
    # Proxy to local backend
    reverse_proxy localhost:3000 {
        # SSE support
        header_up Upgrade websocket
        header_up Connection upgrade
        
        # Preserve client headers
        header_up X-Real-IP {remote}
        header_up X-Forwarded-For {remote}
        header_up X-Forwarded-Proto {scheme}
    }

    # Rate limiting
    rate_limit 10r/s burst=20

    # Security headers (built-in to Caddy 2+)
}
```

Run:

```bash
caddy run
```

---

## HTTPS/TLS

### Let's Encrypt with Certbot

```bash
sudo apt install certbot python3-certbot-nginx

# Generate certificate
sudo certbot certonly --nginx -d allure-mcp.example.com

# Auto-renewal
sudo systemctl enable certbot.timer
sudo systemctl start certbot.timer

# Check renewal
sudo certbot renew --dry-run
```

### Self-Signed Certificate (Testing)

```bash
openssl req -x509 -newkey rsa:4096 -nodes \
    -out cert.pem -keyout key.pem -days 365 \
    -subj "/CN=allure-mcp.local"
```

Use with Caddy:

```caddy
allure-mcp.local {
    tls cert.pem key.pem
    reverse_proxy localhost:3000
}
```

---

## Health Checks

### Docker Health Check

Already configured in `Dockerfile`:

```dockerfile
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD test -S /proc/net/unix || exit 1
```

### Kubernetes Probe

```yaml
livenessProbe:
  httpGet:
    path: /sse
    port: 3000
    scheme: HTTP
  initialDelaySeconds: 10
  periodSeconds: 30
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /messages
    port: 3000
    scheme: HTTP
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 2
```

### Manual Health Check

```bash
curl -v http://localhost:3000/sse
# Should return: event: endpoint\ndata: /messages?sessionId=...
```

---

## Scaling

### Horizontal Scaling (Multiple Replicas)

Allure MCP Server is **stateless** and can run multiple replicas behind a load balancer.

#### With Nginx Load Balancing

```nginx
upstream allure_mcp {
    server localhost:3001;
    server localhost:3002;
    server localhost:3003;
    keepalive 32;
}

server {
    listen 80;
    location / {
        proxy_pass http://allure_mcp;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
}
```

#### With Kubernetes

Scale replicas:

```bash
kubectl scale deployment allure-mcp --replicas=3
```

#### With Docker Swarm

```bash
docker swarm init
docker stack deploy -c docker-compose.yml allure
docker service scale allure_allure-mcp=3
```

### Resource Requirements

**Per instance:**
- CPU: 0.5-1 core
- Memory: 128-256 MB
- Storage: 1 GB (logs)

**For 10 concurrent users:** 2 replicas  
**For 50+ concurrent users:** 5+ replicas

---

## Monitoring

### Logging

Logs are JSON-formatted to stderr. Capture with:

```bash
docker-compose logs -f allure-mcp
```

Or with systemd:

```bash
journalctl -u allure-mcp -f -S "1 hour ago"
```

### Example Log Output

```json
{"level":"INFO","msg":"Starting MCP server","mode":"http","port":3000,"timestamp":"2025-01-15T10:30:00Z"}
{"level":"DEBUG","msg":"Tool called","tool":"run_allure_launch","project_id":1,"timestamp":"2025-01-15T10:30:15Z"}
```

### Prometheus Metrics

For production monitoring, collect logs with:

- **Filebeat** → Elasticsearch/Kibana
- **Loki** + **Grafana**
- **Datadog** / **New Relic** agents

Example Grafana alert:

```yaml
alert: AllureMCPError
expr: |
  count(rate(container_logs_errors[5m])) > 0.1
annotations:
  summary: "High error rate in Allure MCP"
```

### Uptime Monitoring

Use external monitors like:

- **Pingdom**
- **UptimeRobot**
- **New Relic Synthetics**

```bash
curl -f https://allure-mcp.example.com/sse > /dev/null 2>&1
```

---

## Backup & Recovery

### Configuration Backup

Store `.env` securely:

```bash
# Encrypt before storing
gpg --encrypt -r your-key-id .env

# Or use secret manager
vault write secret/allure-mcp \
  base_url="https://allure.example.com" \
  token="xxxxx"
```

### Database/State

Allure MCP Server is **stateless** — no database to backup. All state is in Allure TestOps.

### Disaster Recovery

To restore service:

1. Spin up new server with same configuration
2. Point to existing Allure TestOps instance
3. No data migration needed

```bash
docker-compose -f docker-compose.prod.yml up -d
```

---

## Performance Tuning

### Connection Pooling

Default Go HTTP client reuses connections. For high throughput:

```bash
# Increase file descriptors (Linux)
ulimit -n 65536

# Or in systemd service
[Service]
LimitNOFILE=65536
LimitNPROC=65536
```

### Timeouts

Adjust in `.env`:

```env
REQUEST_TIMEOUT=60  # Increase for slow Allure instances
```

### Caching

Consider adding HTTP caching headers for team deployments:

```nginx
location /sse {
    add_header Cache-Control "no-store, must-revalidate";
    proxy_cache off;
}
```

---

## Checklist

- [ ] SSL/HTTPS enabled
- [ ] Firewall rules configured
- [ ] Health checks passing
- [ ] Monitoring alerts set up
- [ ] Backup strategy documented
- [ ] Log rotation configured
- [ ] Rate limiting enabled
- [ ] MCP_AUTH_TOKEN is strong
- [ ] CORS_ALLOWED_ORIGIN restricted (not `*`)
- [ ] Secrets not in `.env` (use secret manager)
