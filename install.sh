#!/usr/bin/env bash
#
# One-shot installer for the TestOps MCP server (per-user / stdio mode).
#
# Downloads the right binary for your OS/arch from GitHub Releases, asks for your
# Allure TestOps URL + API token, and registers the server in BOTH:
#   * Claude Desktop  (claude_desktop_config.json)
#   * Claude Code      (via `claude mcp add -s user`, if the CLI is installed)
#
# Re-running the script is safe — it overwrites the existing "testops" entry.
#
# Usage:
#   Interactive:      ./install.sh
#   Non-interactive:  ALLURE_BASE_URL=https://your-testops.com ALLURE_TOKEN=xxx ./install.sh
#   Pin a version:    VERSION=v2.0.3 ./install.sh
#
set -euo pipefail

REPO="MimoJanra/TestOpsMCP"
VERSION="${VERSION:-latest}"
SERVER_NAME="${SERVER_NAME:-testops}"

c_cyan() { printf '\033[36m%s\033[0m\n' "$1"; }
c_green() { printf '\033[32m%s\033[0m\n' "$1"; }
c_yellow() { printf '\033[33m%s\033[0m\n' "$1"; }
step() { printf '\n'; c_cyan "==> $1"; }
ok() { c_green "    OK  $1"; }
warn() { c_yellow "    !   $1"; }
die() { printf '\033[31mError: %s\033[0m\n' "$1" >&2; exit 1; }

echo "TestOps MCP installer"

# --- 1. Detect platform ----------------------------------------------------
step "Detecting platform"
case "$(uname -s)" in
    Darwin) PLAT="macos" ;;
    Linux)  PLAT="linux" ;;
    *) die "Unsupported OS: $(uname -s). Use install.ps1 on Windows." ;;
esac
case "$(uname -m)" in
    x86_64|amd64)   ARCH="amd64" ;;
    arm64|aarch64)  ARCH="arm64" ;;
    *) die "Unsupported CPU arch: $(uname -m)" ;;
esac
ASSET="testops-mcp-${PLAT}-${ARCH}"
ok "${PLAT} / ${ARCH}  ->  ${ASSET}"

# --- 2. Download binary -----------------------------------------------------
step "Downloading binary (${VERSION})"
if [ "$VERSION" = "latest" ]; then
    URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"
else
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"
fi

INSTALL_DIR="$HOME/.testops-mcp"
mkdir -p "$INSTALL_DIR"
BIN="$INSTALL_DIR/testops-mcp"

if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$URL" -o "$BIN" || die "Download failed from $URL — check the version tag or download manually from https://github.com/${REPO}/releases"
elif command -v wget >/dev/null 2>&1; then
    wget -qO "$BIN" "$URL" || die "Download failed from $URL"
else
    die "Neither curl nor wget found. Install one and retry."
fi
[ -s "$BIN" ] || die "Downloaded file is empty: $BIN"
chmod +x "$BIN"
ok "saved to $BIN"

# --- 3. Collect Allure credentials -----------------------------------------
step "Allure TestOps credentials"
if [ -z "${ALLURE_BASE_URL:-}" ]; then
    read -r -p "Allure TestOps URL (e.g. https://your-testops.com): " ALLURE_BASE_URL
fi
ALLURE_BASE_URL="${ALLURE_BASE_URL%/}"
case "$ALLURE_BASE_URL" in
    http://*|https://*) ;;
    *) ALLURE_BASE_URL="https://$ALLURE_BASE_URL" ;;
esac
if [ -z "${ALLURE_TOKEN:-}" ]; then
    read -r -s -p "Allure API token (input hidden): " ALLURE_TOKEN
    echo
fi
[ -n "$ALLURE_BASE_URL" ] && [ -n "$ALLURE_TOKEN" ] || die "Both URL and token are required."
ok "URL: $ALLURE_BASE_URL"

# --- 4. Claude Desktop ------------------------------------------------------
if [ "$PLAT" = "macos" ]; then
    DESKTOP_CFG="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
else
    DESKTOP_CFG="${XDG_CONFIG_HOME:-$HOME/.config}/Claude/claude_desktop_config.json"
fi

write_desktop_config() {
    local path="$1"
    mkdir -p "$(dirname "$path")"
    [ -f "$path" ] && cp "$path" "$path.bak"

    if command -v python3 >/dev/null 2>&1; then
        CFG_PATH="$path" SRV_NAME="$SERVER_NAME" SRV_CMD="$BIN" \
        A_URL="$ALLURE_BASE_URL" A_TOK="$ALLURE_TOKEN" python3 - <<'PY'
import json, os
path = os.environ['CFG_PATH']
try:
    with open(path) as f:
        cfg = json.load(f)
    if not isinstance(cfg, dict):
        cfg = {}
except (FileNotFoundError, ValueError):
    cfg = {}
cfg.setdefault('mcpServers', {})
cfg['mcpServers'][os.environ['SRV_NAME']] = {
    'command': os.environ['SRV_CMD'],
    'env': {
        'ALLURE_BASE_URL': os.environ['A_URL'],
        'ALLURE_TOKEN': os.environ['A_TOK'],
    },
}
with open(path, 'w') as f:
    json.dump(cfg, f, indent=2)
    f.write('\n')
PY
        return 0
    fi

    if command -v jq >/dev/null 2>&1; then
        [ -f "$path" ] || echo '{}' > "$path"
        local tmp; tmp="$(mktemp)"
        jq --arg n "$SERVER_NAME" --arg c "$BIN" --arg u "$ALLURE_BASE_URL" --arg t "$ALLURE_TOKEN" \
           '.mcpServers[$n] = {command:$c, env:{ALLURE_BASE_URL:$u, ALLURE_TOKEN:$t}}' \
           "$path" > "$tmp" && mv "$tmp" "$path"
        return 0
    fi

    return 1
}

step "Configuring Claude Desktop"
if write_desktop_config "$DESKTOP_CFG"; then
    ok "updated $DESKTOP_CFG"
else
    warn "Neither python3 nor jq found — cannot safely merge JSON."
    warn "Add this to mcpServers in $DESKTOP_CFG manually:"
    cat <<EOF
      "$SERVER_NAME": {
        "command": "$BIN",
        "env": { "ALLURE_BASE_URL": "$ALLURE_BASE_URL", "ALLURE_TOKEN": "<token>" }
      }
EOF
fi

# --- 5. Claude Code (CLI) ---------------------------------------------------
step "Configuring Claude Code"
if command -v claude >/dev/null 2>&1; then
    claude mcp remove -s user "$SERVER_NAME" >/dev/null 2>&1 || true
    claude mcp add -s user "$SERVER_NAME" "$BIN" \
        -e "ALLURE_BASE_URL=$ALLURE_BASE_URL" \
        -e "ALLURE_TOKEN=$ALLURE_TOKEN"
    ok "registered '$SERVER_NAME' in Claude Code (user scope)"
else
    warn "claude CLI not found on PATH — skipped. Run this later if you use Claude Code:"
    echo "      claude mcp add -s user $SERVER_NAME \"$BIN\" -e ALLURE_BASE_URL=$ALLURE_BASE_URL -e ALLURE_TOKEN=<token>"
fi

# --- Done -------------------------------------------------------------------
c_green "\nDone."
echo "  * Claude Desktop: fully quit and reopen it."
echo "  * Then ask Claude: \"List all projects in Allure\""
