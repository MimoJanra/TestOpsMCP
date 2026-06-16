<#
.SYNOPSIS
    One-shot installer for the TestOps MCP server (per-user / stdio mode).

.DESCRIPTION
    Downloads the right binary for the current OS/arch from GitHub Releases into a
    PERMANENT per-user folder (%LOCALAPPDATA%\TestOpsMCP, not Downloads/Temp which can
    be auto-cleaned), asks the user for the Allure TestOps URL + API token with detailed
    in-console guidance, and registers the server in BOTH:
      * Claude Desktop  (claude_desktop_config.json)
      * Claude Code      (via `claude mcp add -s user`, if the CLI is installed)

    Re-running is safe: it overwrites the existing "testops" entry. Config files are
    backed up to *.bak before being modified.

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File .\install.ps1

.EXAMPLE
    powershell -ExecutionPolicy Bypass -File .\install.ps1 -AllureBaseUrl https://allure.acme.com -AllureToken xxxx
#>
[CmdletBinding()]
param(
    [string]$AllureBaseUrl = $env:ALLURE_BASE_URL,
    [string]$AllureToken   = $env:ALLURE_TOKEN,
    [string]$Version       = "latest",
    [string]$ServerName    = "testops",
    [switch]$NoClaudeCode,
    [switch]$NoClaudeDesktop
)

$ErrorActionPreference = "Stop"
$Repo = "MimoJanra/TestOpsMCP"

# Allow HTTPS downloads on older Windows that still default to TLS 1.0/1.1.
try { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 } catch {}

function Write-Line($msg)  { Write-Host $msg }
function Write-Head($msg)  { Write-Host ""; Write-Host $msg -ForegroundColor Cyan }
function Write-Ok($msg)    { Write-Host "   [OK] $msg" -ForegroundColor Green }
function Write-Bad($msg)   { Write-Host "   [!] $msg" -ForegroundColor Red }
function Write-Hint($msg)  { Write-Host $msg -ForegroundColor DarkGray }

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

Write-Host "============================================================" -ForegroundColor White
Write-Host "   TestOps for Claude - Installer" -ForegroundColor White
Write-Host "============================================================" -ForegroundColor White
Write-Line ""
Write-Line "This wizard will do everything for you:"
Write-Line "  - download the program,"
Write-Line "  - ask you for 2 things (your Allure URL and an access token),"
Write-Line "  - configure Claude automatically."
Write-Line "You do not have to edit any files by hand. Just follow the prompts."

# --- Step 1: platform -------------------------------------------------------
Write-Head "[Step 1 of 4] Detecting your system..."
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64' -or $env:PROCESSOR_ARCHITEW6432 -eq 'ARM64') { 'arm64' } else { 'amd64' }
$asset = "testops-mcp-windows-$arch.exe"
Write-Ok "Windows ($arch)"

# --- Step 2: download -------------------------------------------------------
Write-Head "[Step 2 of 4] Downloading the program from GitHub..."
if ($Version -eq "latest") {
    $url = "https://github.com/$Repo/releases/latest/download/$asset"
} else {
    $url = "https://github.com/$Repo/releases/download/$Version/$asset"
}
# Permanent per-user folder - survives clearing Downloads / Temp.
$installDir = Join-Path $env:LOCALAPPDATA "TestOpsMCP"
if (-not (Test-Path $installDir)) { New-Item -ItemType Directory -Force -Path $installDir | Out-Null }
$binPath = Join-Path $installDir "testops-mcp.exe"

try {
    Invoke-WebRequest -Uri $url -OutFile $binPath -UseBasicParsing
} catch {
    Write-Bad "Could not download the program."
    Write-Line ""
    Write-Line "Possible reasons:"
    Write-Line "  - no internet, or a corporate proxy/VPN is blocking GitHub;"
    Write-Line "  - GitHub is temporarily unavailable."
    Write-Line ""
    Write-Line "What to do: check your internet connection and run the installer again."
    Write-Line "If it still fails, send a screenshot of this window to whoever gave you the installer."
    Write-Hint "Technical details: $url"
    Write-Hint $_.Exception.Message
    exit 1
}
if (-not (Test-Path $binPath) -or (Get-Item $binPath).Length -eq 0) {
    Write-Bad "The downloaded file is empty. Please run the installer again."
    exit 1
}
Write-Ok "Done. Saved to a permanent folder (will NOT be deleted when you clear Downloads):"
Write-Hint "       $binPath"

# --- Step 3: Allure credentials --------------------------------------------
Write-Head "[Step 3 of 4] Enter your 2 details from Allure TestOps"
Write-Line "------------------------------------------------------------"

# (1) URL
if (-not $AllureBaseUrl) {
    Write-Line ""
    Write-Line "(1) ALLURE URL"
    Write-Line "    This is the link you use to open Allure in your browser."
    Write-Line "    Example:  https://allure.yourcompany.com"
    Write-Line "    Just the site address (no /login, no trailing slash)."
    Write-Line ""
    $AllureBaseUrl = Read-Host "    Enter your Allure URL and press Enter"
}
$AllureBaseUrl = $AllureBaseUrl.Trim().TrimEnd('/')
if ($AllureBaseUrl -notmatch '^https?://') { $AllureBaseUrl = "https://$AllureBaseUrl" }

# (2) Token
if (-not $AllureToken) {
    Write-Line ""
    Write-Line "(2) ALLURE API TOKEN"
    Write-Line "    Where to get it (takes about 30 seconds):"
    Write-Line "      1. Open Allure in your browser and sign in."
    Write-Line "      2. Click your avatar / name in the TOP-RIGHT corner."
    Write-Line "      3. Go to 'Settings' -> 'API tokens'"
    Write-Line "         (it may be called 'Tokens' or 'Integration tokens')."
    Write-Line "      4. Click 'Create' / 'Generate' / '+ Create token'."
    Write-Line "      5. Give it a name, e.g. 'Claude', and confirm."
    Write-Line "      6. Copy the token NOW - it is shown only ONCE!"
    Write-Line "      7. Paste it below (right-click in this window = paste)."
    Write-Line ""
    Write-Hint "    NOTE: the token will NOT show on screen while you paste - that is normal"
    Write-Hint "    (it is hidden for security). Just paste it and press Enter."
    Write-Line ""
    $sec = Read-Host "    Paste your token and press Enter" -AsSecureString
    $AllureToken = [Runtime.InteropServices.Marshal]::PtrToStringAuto(
        [Runtime.InteropServices.Marshal]::SecureStringToBSTR($sec))
}
$AllureToken = $AllureToken.Trim()

Write-Line "------------------------------------------------------------"
if (-not $AllureBaseUrl -or -not $AllureToken) {
    Write-Bad "Both the URL and the token are required. Please run the installer again."
    exit 1
}
Write-Ok "URL accepted: $AllureBaseUrl"
Write-Ok ("Token accepted (length: {0} characters)" -f $AllureToken.Length)

$envVars = @{ ALLURE_BASE_URL = $AllureBaseUrl; ALLURE_TOKEN = $AllureToken }

# --- Step 4: configure Claude ----------------------------------------------
Write-Head "[Step 4 of 4] Configuring Claude..."
if (-not $NoClaudeDesktop) {
    $desktopCfg = Join-Path $env:APPDATA "Claude\claude_desktop_config.json"
    Update-DesktopConfig -Path $desktopCfg -Name $ServerName -Command $binPath -EnvVars $envVars
    Write-Ok "Claude Desktop configured."
    Write-Hint "       (file: $desktopCfg)"
}
if (-not $NoClaudeCode) {
    $claude = Get-Command claude -ErrorAction SilentlyContinue
    if ($claude) {
        & claude mcp remove -s user $ServerName *> $null
        & claude mcp add -s user $ServerName $binPath `
            -e "ALLURE_BASE_URL=$AllureBaseUrl" `
            -e "ALLURE_TOKEN=$AllureToken"
        if ($LASTEXITCODE -eq 0) { Write-Ok "Claude Code configured too." }
        else { Write-Bad "Claude Code: command returned code $LASTEXITCODE (not critical)." }
    } else {
        Write-Ok "Claude Code not found - skipping (that's fine if you don't use it)."
    }
}

# --- Done -------------------------------------------------------------------
Write-Line ""
Write-Host "============================================================" -ForegroundColor Green
Write-Host "   DONE! Just 2 simple steps left:" -ForegroundColor Green
Write-Host "============================================================" -ForegroundColor Green
Write-Line ""
Write-Line "  1) RESTART Claude Desktop completely:"
Write-Line "       - find the Claude icon in the system tray (bottom-right, near the clock);"
Write-Line "       - right-click it -> 'Quit';"
Write-Line "       - open Claude again."
Write-Line "     IMPORTANT: closing the window with the X is NOT enough."
Write-Line ""
Write-Line "  2) TEST it: send this message to Claude:"
Write-Host "       List all projects in Allure" -ForegroundColor White
Write-Line ""
Write-Line "If something went wrong, send a screenshot of this window to"
Write-Line "whoever gave you the installer."
