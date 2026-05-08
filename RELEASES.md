# Release Downloads & Setup

All pre-built binaries are provided for easy setup without requiring Go installation.

## 📥 Available Binaries

| Platform | Download | Architecture |
|----------|----------|---------------|
| **Windows** | `testops-windows-amd64.exe` | x86-64 (64-bit) |
| **macOS Intel** | `testops-darwin-amd64` | x86-64 (Intel) |
| **macOS Apple Silicon** | `testops-darwin-arm64` | ARM64 (M1/M2/M3) |
| **Linux** | `testops-linux-amd64` | x86-64 (64-bit) |

## 🚀 Setup Instructions

### Windows

1. Download `testops-windows-amd64.exe`
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
   .\testops-windows-amd64.exe --http
   ```

### macOS (Intel & Apple Silicon)

1. Download appropriate binary:
   - **Intel (x86-64)**: `testops-darwin-amd64`
   - **Apple Silicon (M1/M2/M3)**: `testops-darwin-arm64`

2. Create setup directory:
   ```bash
   mkdir ~/testops
   cd ~/testops
   mv ~/Downloads/testops-darwin-* .
   chmod +x testops-darwin-*
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
   ./testops-darwin-amd64 --http
   
   # Apple Silicon
   ./testops-darwin-arm64 --http
   ```

### Linux

1. Download `testops-linux-amd64`
2. Setup directory:
   ```bash
   mkdir ~/testops
   cd ~/testops
   mv ~/Downloads/testops-linux-amd64 .
   chmod +x testops-linux-amd64
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
   ./testops-linux-amd64 --http
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
./testops-darwin-amd64
```
Used for Claude Desktop integration via MCP configuration.

**HTTP Mode** (Team/Server):
```bash
./testops-darwin-amd64 --http
```
Runs on `http://localhost:3000` for team deployments.

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
docker run -d \
  -e ALLURE_BASE_URL=https://your-testops.com \
  -e ALLURE_TOKEN=your-token \
  -p 3000:3000 \
  mimjanra/testops-mcp:latest
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

Over **55 MCP tools** for Allure TestOps integration:

- Launch management (create, run, close, reopen, copy, merge)
- Test case operations (create, update, delete, clone, restore)
- Test result handling (assign, mute, resolve, unmute)
- Defect management (link, unlink, view)
- Team collaboration (members, external links)
- Bulk operations (mass clone, bulk updates)
- Analytics and reporting

See [README.md](README.md) for complete tool documentation.

## 📞 Support

For issues or questions:
1. Check Allure TestOps connection settings
2. Verify API token hasn't expired
3. Review logs in console output
4. Open an issue on GitHub with error details
