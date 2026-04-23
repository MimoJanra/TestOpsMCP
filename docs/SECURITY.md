# Security Guide

Security considerations and best practices for Allure MCP Server.

## Table of Contents

- [Authentication](#authentication)
- [Network Security](#network-security)
- [Secret Management](#secret-management)
- [Compliance](#compliance)
- [Security Checklist](#security-checklist)

## Authentication

### Stdio Mode (Local Development)

**Security Model:** Subprocess with inherited privileges

- No auth: runs as your user, no network exposure
- Suitable only for **local development** and Claude Desktop
- No credentials transmitted over network
- **⚠️ Never use stdio mode with untrusted code**

### HTTP Mode (Team/Server)

**Always set `MCP_AUTH_TOKEN` in production.**

```bash
# Generate a strong random token
openssl rand -base64 32

# Export to .env
MCP_AUTH_TOKEN=your_generated_token_here
```

Clients must include the token:

```bash
curl -H "Authorization: Bearer $MCP_AUTH_TOKEN" http://localhost:3000/sse
```

The token is checked on:
- `GET /sse` — SSE stream endpoint
- `POST /messages` — Message submission endpoint

**Note:** Token is case-sensitive and checked with Bearer scheme.

---

## Network Security

### Use HTTPS in Production

**Never expose HTTP over the internet.** Always use HTTPS.

#### Option 1: Reverse Proxy with TLS

Nginx example:

```nginx
server {
    listen 443 ssl http2;
    server_name allure-mcp.example.com;

    ssl_certificate /etc/letsencrypt/live/allure-mcp.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/allure-mcp.example.com/privkey.pem;

    # TLS hardening
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### Option 2: Caddy (Auto HTTPS)

```caddy
allure-mcp.example.com {
    reverse_proxy localhost:3000
}
```

Caddy automatically obtains and renews HTTPS certificates.

#### Option 3: Kubernetes with TLS

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: allure-mcp
spec:
  tls:
  - hosts:
    - allure-mcp.example.com
    secretName: allure-mcp-tls
  rules:
  - host: allure-mcp.example.com
    http:
      paths:
      - path: /
        backend:
          service:
            name: allure-mcp
            port:
              number: 3000
```

### Restrict CORS

**Default `CORS_ALLOWED_ORIGIN=*` allows any site to call your server.**

Set to specific domains:

```bash
# Production
CORS_ALLOWED_ORIGIN=https://claude.ai

# Internal team only
CORS_ALLOWED_ORIGIN=https://allure-mcp.internal.example.com
```

In Nginx:

```nginx
location / {
    add_header 'Access-Control-Allow-Origin' 'https://claude.ai' always;
    add_header 'Access-Control-Allow-Methods' 'GET,POST,OPTIONS' always;
    add_header 'Access-Control-Allow-Headers' 'Content-Type,Authorization' always;
}
```

### Rate Limiting

Nginx:

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

location / {
    limit_req zone=api_limit burst=20 nodelay;
    proxy_pass http://localhost:3000;
}
```

Caddy:

```caddy
allure-mcp.example.com {
    rate_limit 10r/s burst=20
    reverse_proxy localhost:3000
}
```

### Firewall Rules

Allow only trusted IPs:

```bash
# Nginx
location / {
    allow 10.0.0.0/8;     # Internal network
    allow 203.0.113.0/24;  # Team office
    deny all;
}
```

Or at firewall level:

```bash
# iptables
sudo iptables -A INPUT -p tcp --dport 3000 -s 10.0.0.0/8 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 3000 -j DROP
```

---

## Secret Management

### Never Commit `.env`

`.gitignore` already contains:

```
.env
.env.local
.env.*.secret
```

Verify before committing:

```bash
git status
# Should NOT show .env

git diff --cached | grep -i "token\|password\|secret"
# Should return nothing
```

### Store Secrets Securely

**Development:**
- Use `.env` (local only)
- Never commit to git
- Limit file permissions: `chmod 600 .env`

**Team/Production:**
- Use secret manager: Vault, AWS Secrets Manager, Azure Key Vault
- Example with Vault:

```bash
vault write secret/allure-mcp \
  base_url="https://allure.example.com" \
  token="$(openssl rand -base64 32)"
```

- Read in startup script:

```bash
export ALLURE_BASE_URL=$(vault kv get -field=base_url secret/allure-mcp)
export ALLURE_TOKEN=$(vault kv get -field=token secret/allure-mcp)
./bin/server --http
```

**Kubernetes:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: allure-credentials
type: Opaque
stringData:
  ALLURE_BASE_URL: https://allure.example.com
  ALLURE_TOKEN: your_token_here
  MCP_AUTH_TOKEN: your_mcp_secret
```

Mount in pod:

```yaml
env:
- name: ALLURE_TOKEN
  valueFrom:
    secretKeyRef:
      name: allure-credentials
      key: ALLURE_TOKEN
```

### Rotate Credentials

**Allure API Token:**

1. Generate new token in Allure UI
2. Update `ALLURE_TOKEN` in secret manager
3. Restart server with new token
4. Revoke old token in Allure UI

**MCP Auth Token:**

1. Generate new token: `openssl rand -base64 32`
2. Update `MCP_AUTH_TOKEN` in `.env` or secret manager
3. Restart server
4. Notify team of new token
5. Clients must update their config

---

## Compliance

### Data Privacy

The server:
- ✓ Does not store user data
- ✓ Does not log API tokens
- ✓ Does not persist launch history
- ✓ Proxies requests only to Allure TestOps

All state is in Allure TestOps. This server is stateless.

**⚠️ Logs may contain launch names/project IDs** — handle accordingly if sensitive.

### Audit Logging

Enable structured logging for audit trails:

```bash
LOG_LEVEL=INFO
# Server logs all tool calls and errors
```

Example log:

```json
{"level":"INFO","msg":"Tool called","tool":"run_allure_launch","project_id":1,"timestamp":"2025-01-15T10:30:00Z","clientIP":"10.0.0.1"}
```

Send logs to centralized system:

- **ELK Stack** (Elasticsearch + Kibana + Logstash)
- **Datadog** / **New Relic** / **Splunk**
- **Loki** + **Grafana**

### Compliance Standards

**SOC2:**
- ✓ Encrypted secrets (use secret manager)
- ✓ Access logs with audit trail
- ✓ Rate limiting prevents abuse
- ✓ TLS/HTTPS enforced

**GDPR:**
- ✓ No personal data stored
- ✓ Stateless design
- ✓ Can be deleted without data migration

**HIPAA (for healthcare):**
- ✓ Encrypt at rest (secret manager)
- ✓ Encrypt in transit (HTTPS)
- ✓ Access logs (centralized monitoring)
- ✓ Run in VPC/private network

---

## Security Checklist

### Before Production Deployment

- [ ] **Authentication**
  - [ ] `MCP_AUTH_TOKEN` is set to strong random value
  - [ ] Token is stored in secret manager (Vault, AWS Secrets Manager, etc.)
  - [ ] Token rotated monthly

- [ ] **Network**
  - [ ] HTTPS/TLS enabled
  - [ ] Reverse proxy (Nginx, Caddy) configured
  - [ ] Firewall rules restrict access
  - [ ] CORS_ALLOWED_ORIGIN is specific (not `*`)
  - [ ] Rate limiting enabled

- [ ] **Secrets**
  - [ ] `.env` not committed to git
  - [ ] Allure token stored in secret manager
  - [ ] MCP auth token stored in secret manager
  - [ ] File permissions: `chmod 600` for `.env`

- [ ] **Monitoring**
  - [ ] Structured JSON logging enabled
  - [ ] Logs sent to centralized system
  - [ ] Alerting configured for errors/failures
  - [ ] Access logs reviewed regularly

- [ ] **Infrastructure**
  - [ ] Running as non-root user (UID 1000)
  - [ ] Read-only filesystem where possible
  - [ ] Resource limits set (CPU, memory)
  - [ ] Health checks configured
  - [ ] Backups documented (if applicable)

- [ ] **Updates**
  - [ ] Go version is 1.22+
  - [ ] Dependencies up-to-date (`go mod tidy`)
  - [ ] Security advisories checked (`go list -u -m all`)
  - [ ] Update process documented

- [ ] **Documentation**
  - [ ] Security policy documented
  - [ ] Incident response plan in place
  - [ ] Credential rotation schedule documented
  - [ ] Disaster recovery plan documented

### Before Each Deployment

- [ ] No secrets in code or commit history
- [ ] All tests pass (`make check`)
- [ ] Linting passes (`make lint`)
- [ ] Build is reproducible (`make clean && make build`)
- [ ] CHANGELOG updated
- [ ] Security fixes highlighted

---

## Incident Response

### If API Token Leaks

1. **Immediately** revoke token in Allure UI
2. Generate new token
3. Update secret manager with new token
4. Restart all server instances
5. Review logs for unauthorized access
6. Document timeline and root cause

### If MCP Auth Token Leaks

1. Generate new token: `openssl rand -base64 32`
2. Update `.env` or secret manager
3. Restart server
4. Notify all team members
5. Force clients to update config
6. Review logs for unauthorized access

### If Server Compromised

1. **Stop the server** immediately
2. Revoke both Allure token and MCP auth token
3. Review all logs for unauthorized tool calls
4. Redeploy clean version
5. Update all credentials
6. Notify stakeholders
7. Document incident

---

## Reporting Security Issues

**Do NOT open public GitHub issues for security vulnerabilities.**

Email security concerns to: mimojanra@gmail.com

Include:
- Description of vulnerability
- Steps to reproduce
- Potential impact
- Proposed fix (if any)

We will:
1. Acknowledge receipt within 24 hours
2. Assess severity
3. Prepare patch (if applicable)
4. Coordinate disclosure with you
5. Release patched version
6. Credit you (with permission)

---

## Additional Resources

- [OWASP Top 10](https://owasp.org/Top10/) — Common vulnerabilities
- [Go Security Checklist](https://owasp.org/www-community/attacks/Go_Code_Quality_Tools) — Go-specific security
- [MCP Security Spec](https://spec.modelcontextprotocol.io/) — MCP protocol security
- [TLS Best Practices](https://wiki.mozilla.org/Security/Server_Side_TLS) — Mozilla TLS recommendations

---

**Keep security updated. Rotate credentials. Monitor logs. 🔐**
