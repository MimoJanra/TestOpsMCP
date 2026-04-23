#!/bin/bash
# Example: Launch tests via Allure MCP Server
# Usage: ./launch-tests.sh <project_id> <launch_name>

set -euo pipefail

# Configuration
SERVER_URL="${SERVER_URL:-http://localhost:3000}"
AUTH_TOKEN="${MCP_AUTH_TOKEN:-}"
PROJECT_ID="${1:-1}"
LAUNCH_NAME="${2:-CI Tests}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Step 1: Open SSE stream and get session ID
log_info "Connecting to Allure MCP server at $SERVER_URL..."

# Create a temporary file for the SSE stream
SSE_TEMP=$(mktemp)
trap "rm -f $SSE_TEMP" EXIT

# Start SSE stream in background
HEADERS=""
if [ -n "$AUTH_TOKEN" ]; then
    HEADERS="-H Authorization: Bearer $AUTH_TOKEN"
fi

curl -s $HEADERS "$SERVER_URL/sse" > "$SSE_TEMP" &
SSE_PID=$!

# Give it a moment to connect
sleep 1

# Extract session ID from SSE stream
SESSION_ID=$(grep -oP 'sessionId=\K[^"]+' "$SSE_TEMP" | head -1)

if [ -z "$SESSION_ID" ]; then
    log_error "Failed to get session ID from server"
    kill $SSE_PID 2>/dev/null || true
    exit 1
fi

log_success "Connected with session ID: $SESSION_ID"

# Step 2: Launch tests
log_info "Launching tests (project_id=$PROJECT_ID, name=\"$LAUNCH_NAME\")..."

LAUNCH_RESPONSE=$(curl -s -X POST "$SERVER_URL/messages?sessionId=$SESSION_ID" \
    $HEADERS \
    -H "Content-Type: application/json" \
    -d "{
        \"jsonrpc\": \"2.0\",
        \"id\": 1,
        \"method\": \"tools/call\",
        \"params\": {
            \"name\": \"run_allure_launch\",
            \"arguments\": {
                \"project_id\": $PROJECT_ID,
                \"launch_name\": \"$LAUNCH_NAME\"
            }
        }
    }")

# Wait for response on SSE stream
sleep 2

LAUNCH_MESSAGE=$(tail "$SSE_TEMP" | grep -o '"text":"[^"]*"' | tail -1 | cut -d'"' -f4)

if [ -z "$LAUNCH_MESSAGE" ]; then
    log_error "Failed to launch tests"
    kill $SSE_PID 2>/dev/null || true
    exit 1
fi

log_success "Launch created: $LAUNCH_MESSAGE"

# Extract launch ID (format: "Launch started: ID=12345, Name='Test'")
LAUNCH_ID=$(echo "$LAUNCH_MESSAGE" | grep -oP 'ID=\K[0-9]+')

if [ -z "$LAUNCH_ID" ]; then
    log_error "Could not extract launch ID from response"
    kill $SSE_PID 2>/dev/null || true
    exit 1
fi

log_success "Launch ID: $LAUNCH_ID"

# Step 3: Poll status until complete
log_info "Waiting for tests to complete..."

MAX_WAIT=300 # 5 minutes
ELAPSED=0
POLL_INTERVAL=10

while [ $ELAPSED -lt $MAX_WAIT ]; do
    STATUS_RESPONSE=$(curl -s -X POST "$SERVER_URL/messages?sessionId=$SESSION_ID" \
        $HEADERS \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"id\": 2,
            \"method\": \"tools/call\",
            \"params\": {
                \"name\": \"get_launch_status\",
                \"arguments\": {
                    \"launch_id\": $LAUNCH_ID
                }
            }
        }")

    sleep 2

    STATUS_MESSAGE=$(tail "$SSE_TEMP" | grep -o '"text":"[^"]*"' | tail -1 | cut -d'"' -f4)

    log_info "Status: $STATUS_MESSAGE"

    # Check if COMPLETED
    if echo "$STATUS_MESSAGE" | grep -q "COMPLETED\|SUBMITTED"; then
        log_success "Tests completed!"
        break
    fi

    ELAPSED=$((ELAPSED + POLL_INTERVAL))
    sleep $POLL_INTERVAL
done

# Step 4: Get final report
log_info "Fetching final report..."

REPORT_RESPONSE=$(curl -s -X POST "$SERVER_URL/messages?sessionId=$SESSION_ID" \
    $HEADERS \
    -H "Content-Type: application/json" \
    -d "{
        \"jsonrpc\": \"2.0\",
        \"id\": 3,
        \"method\": \"tools/call\",
        \"params\": {
            \"name\": \"get_launch_report\",
            \"arguments\": {
                \"launch_id\": $LAUNCH_ID
            }
        }
    }")

sleep 2

REPORT_MESSAGE=$(tail "$SSE_TEMP" | grep -o '"text":"[^"]*"' | tail -1 | cut -d'"' -f4)

log_success "Final Report: $REPORT_MESSAGE"

# Check for failures
if echo "$REPORT_MESSAGE" | grep -qE "Failed=[1-9]|Broken=[1-9]"; then
    log_error "Some tests failed or were broken"
    kill $SSE_PID 2>/dev/null || true
    exit 1
fi

log_success "All tests passed!"
kill $SSE_PID 2>/dev/null || true
exit 0
