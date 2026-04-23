# Allure MCP Server

A Model Context Protocol (MCP) server that integrates with Allure TestOps to manage test launches and generate test reports.

## Overview

This server implements the MCP specification and provides tools for:
- Creating test launches in Allure TestOps
- Retrieving launch status
- Fetching test execution reports and statistics

It communicates via stdin/stdout using the JSON-RPC 2.0 protocol, making it compatible with any MCP client.

## Requirements

- Go 1.22+
- Access to an Allure TestOps instance
- Valid Allure API token

## Installation

### Build from source

```bash
go build -o bin/server ./cmd/server
```

### Dependencies

```bash
go mod tidy
```

## Configuration

Set the following environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `ALLURE_BASE_URL` | Base URL of your Allure TestOps instance (e.g., `https://allure.example.com`) | Yes |
| `ALLURE_TOKEN` | Allure API token for authentication | Yes |
| `REQUEST_TIMEOUT` | HTTP request timeout in seconds (default: 30) | No |

### Example

```bash
export ALLURE_BASE_URL=https://allure.example.com
export ALLURE_TOKEN=your_api_token
export REQUEST_TIMEOUT=30
```

## Usage

Run the server:

```bash
./bin/server
```

The server listens on stdin and outputs responses to stdout following the MCP protocol.

### Integration

Connect your MCP client to this server. The client will communicate via JSON-RPC 2.0 requests formatted as newline-delimited JSON.

## Available Tools

### `run_allure_launch`

Start a new test launch in Allure TestOps.

**Parameters:**
- `project_id` (integer, required): Allure project ID
- `launch_name` (string, required): Name for the test launch

**Response:**
```json
{
  "launch_id": 12345,
  "status": "started"
}
```

### `get_launch_status`

Retrieve the current status of a test launch.

**Parameters:**
- `launch_id` (integer, required): Allure launch ID

**Response:**
```json
{
  "status": "RUNNING"
}
```

**Possible statuses:** `CREATED`, `RUNNING`, `STOPPED`, `CLOSED`

### `get_launch_report`

Get test execution statistics for a launch.

**Parameters:**
- `launch_id` (integer, required): Allure launch ID

**Response:**
```json
{
  "total": 100,
  "passed": 85,
  "failed": 10,
  "broken": 5
}
```

## API Endpoints

The server communicates with Allure TestOps using the Report Service API:

- `POST /api/rs/launch` — Create a launch
- `GET /api/rs/launch/{id}` — Get launch details
- `GET /api/rs/launch/{id}/statistic` — Get launch statistics

See [Allure TestOps API documentation](https://docs.qameta.io/allure-testops/advanced/api/) for more details.

## Architecture

```
cmd/
  server/
    main.go              # Entry point, initialization

internal/
  adapters/
    allure/
      client.go          # Allure API client
      models.go          # Request/response structures
  config/
    config.go            # Configuration loading
  core/
    logger.go            # Structured logging
  mcp/
    server.go            # MCP protocol server
    protocol.go          # JSON-RPC types
  tools/
    registry.go          # Tool registration and handlers
```

## Logging

All logs are written to stderr in JSON format for structured logging. stdout is reserved for MCP protocol messages.

## Protocol Details

The server implements MCP 2024-11-05 and follows the [Model Context Protocol specification](https://spec.modelcontextprotocol.io/).

### Initialization

Clients must call `initialize` before using tools:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {},
    "clientInfo": {
      "name": "client-name",
      "version": "1.0.0"
    }
  }
}
```

### Tool Execution

To execute a tool, send a `tools/call` request:

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "run_allure_launch",
    "arguments": {
      "project_id": 1,
      "launch_name": "Smoke Tests"
    }
  }
}
```

## Error Handling

Errors during tool execution are returned with `isError: true` in the response content:

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Tool execution failed: invalid project_id"
      }
    ],
    "isError": true
  }
}
```

## Development

### Running tests

```bash
go test ./...
```

### Building

```bash
go build -o bin/server ./cmd/server
```

### Code structure

- Uses `sync.RWMutex` for thread-safe tool registry access
- Implements graceful shutdown via signal handling (SIGINT, SIGTERM)
- Context propagation throughout the call stack for timeout support
- Structured logging with correlation metadata

## License

See LICENSE file for details.

## References

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Allure TestOps API Documentation](https://docs.qameta.io/allure-testops/advanced/api/)
- [Allure TestOps User Guide](https://docs.qameta.io/allure-testops/)
