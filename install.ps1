<#
.SYNOPSIS
    One-shot installer for the TestOps MCP server (per-user / stdio mode).

.DESCRIPTION
    Downloads the right binary for your OS/arch from GitHub Releases, asks for your
    Allure TestOps URL + API token, and registers the server in BOTH:
      * Claude Desktop  (claude_desktop_config.json)
      * Claude Code      (via `claude mcp add -s user`, if the CLI is installed)

    Re-running the script is safe — it overwrites the existing "testops" entry.

.EXAMPLE
    # Interactive (asks for URL + token):
    powershell -ExecutionPolicy Bypass -File .\install.ps1

.EXAMPLE
    # Non-interactive:
    powershell -ExecutionPolicy Bypass -File .\install.ps1 `
        -AllureBaseUrl https://your-testops.com -AllureToken xxxxxxxx
#>
[CmdletBinding()]
param(
    [string]$AllureBaseUrl = $env:ALLURE_BASE_URL,
    [string]$AllureToken   = $env:ALLURE_TOKEN,
    [string]$Version       = "latest",       # e.g. v2.0.3 ; "latest" = newest release
    [string]$ServerName    = "testops",
    [switch]$NoClaudeCode,
    [switch]$NoClaudeDesktop
)

$ErrorActionPreference = "Stop"
$Repo = "MimoJanra/TestOpsMCP"

# TestOps MCP supports older Windows that default to TLS 1.0/1.1.
try { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 } catch {}

function Write-Step($msg) { Write-Host "`n==> $msg" -ForegroundColor Cyan }
function Write-Ok($msg)   { Write-Host "    OK  $msg" -ForegroundColor Green }
function Write-Warn2($msg) { Write-Host "    !   $msg" -ForegroundColor Yellow }

function Update-DesktopConfig {
    param([string]$Path, [string]$Name, [string]$Command, [hashtable]$EnvVars)

    $dir = Split-Path -Parent $Path
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }

    if (Test-Path $Path) {
        Copy-Item $Path "$Path.bak" -Force
        $raw = Get-Content -Raw -Path $Path
        if ([string]::IsNullOrWhiteSpace($raw)) {
            $config = [pscustomobject]@{}
        } else {
            $config = $raw | ConvertFrom-Json
        }
    } else {
        $config = [pscustomobject]@{}
    }

    if (-not ($config.PSObject.Properties.Name -contains 'mcpServers')) {
        $config | Add-Member -NotePropertyName 'mcpServers' -NotePropertyValue ([pscustomobject]@{}) -Force
    }

    $envObj = [pscustomobject]@{}
    foreach ($k in $EnvVars.Keys) {
        $envObj | Add-Member -NotePropertyName $k -NotePropertyValue $EnvVars[$k] -Force
    }

    $serverEntry = [pscustomobject]@{ command = $Command; env = $envObj }
    $config.mcpServers | Add-Member -NotePropertyName $Name -NotePropertyValue $serverEntry -Force

    $json = $config | ConvertTo-Json -Depth 100
    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($Path, $json, $utf8NoBom)
}

Write-Host "TestOps MCP installer" -ForegroundColor White

# --- 1. Detect platform ---------------------------------------------------
Write-Step "Detecting platform"
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64' -or $env:PROCESSOR_ARCHITEW6432 -eq 'ARM64') { 'arm64' } else { 'amd64' }
$asset = "testops-mcp-windows-$arch.exe"
Write-Ok "windows / $arch  ->  $asset"

# --- 2. Download binary ----------------------------------------------------
Write-Step "Downloading binary ($Version)"
if ($Version -eq "latest") {
    $url = "https://github.com/$Repo/releases/latest/download/$asset"
} else {
    $url = "https://github.com/$Repo/releases/download/$Version/$asset"
}

$installDir = Join-Path $env:LOCALAPPDATA "TestOpsMCP"
if (-not (Test-Path $installDir)) { New-Item -ItemType Directory -Force -Path $installDir | Out-Null }
$binPath = Join-Path $installDir "testops-mcp.exe"

try {
    Invoke-WebRequest -Uri $url -OutFile $binPath -UseBasicParsing
} catch {
    throw "Download failed from $url`n$($_.Exception.Message)`nCheck the version tag or download manually from https://github.com/$Repo/releases"
}
if (-not (Test-Path $binPath) -or (Get-Item $binPath).Length -eq 0) {
    throw "Downloaded file is empty: $binPath"
}
Write-Ok "saved to $binPath"

# --- 3. Collect Allure credentials ----------------------------------------
Write-Step "Allure TestOps credentials"
if (-not $AllureBaseUrl) {
    $AllureBaseUrl = Read-Host "Allure TestOps URL (e.g. https://your-testops.com)"
}
$AllureBaseUrl = $AllureBaseUrl.Trim().TrimEnd('/')
if ($AllureBaseUrl -notmatch '^https?://') { $AllureBaseUrl = "https://$AllureBaseUrl" }

if (-not $AllureToken) {
    $sec = Read-Host "Allure API token (input hidden)" -AsSecureString
    $AllureToken = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
        [Runtime.InteropServices.Marshal]::SecureStringToBSTR($sec))
}
$AllureToken = $AllureToken.Trim()
if (-not $AllureBaseUrl -or -not $AllureToken) { throw "Both URL and token are required." }
Write-Ok "URL: $AllureBaseUrl"

$envVars = @{ ALLURE_BASE_URL = $AllureBaseUrl; ALLURE_TOKEN = $AllureToken }

# --- 4. Claude Desktop -----------------------------------------------------
if (-not $NoClaudeDesktop) {
    Write-Step "Configuring Claude Desktop"
    $desktopCfg = Join-Path $env:APPDATA "Claude\claude_desktop_config.json"
    Update-DesktopConfig -Path $desktopCfg -Name $ServerName -Command $binPath -EnvVars $envVars
    Write-Ok "updated $desktopCfg"
}

# --- 5. Claude Code (CLI) --------------------------------------------------
if (-not $NoClaudeCode) {
    Write-Step "Configuring Claude Code"
    $claude = Get-Command claude -ErrorAction SilentlyContinue
    if ($claude) {
        & claude mcp remove -s user $ServerName *> $null
        & claude mcp add -s user $ServerName $binPath `
            -e "ALLURE_BASE_URL=$AllureBaseUrl" `
            -e "ALLURE_TOKEN=$AllureToken"
        if ($LASTEXITCODE -eq 0) { Write-Ok "registered '$ServerName' in Claude Code (user scope)" }
        else { Write-Warn2 "claude CLI returned exit code $LASTEXITCODE" }
    } else {
        Write-Warn2 "claude CLI not found on PATH — skipped. Run this later if you use Claude Code:"
        Write-Host "      claude mcp add -s user $ServerName `"$binPath`" -e ALLURE_BASE_URL=$AllureBaseUrl -e ALLURE_TOKEN=<token>"
    }
}

# --- Done ------------------------------------------------------------------
Write-Host "`nDone." -ForegroundColor Green
Write-Host "  * Claude Desktop: fully restart it (quit from the tray, not just close the window)."
Write-Host "  * Then ask Claude: `"List all projects in Allure`""
