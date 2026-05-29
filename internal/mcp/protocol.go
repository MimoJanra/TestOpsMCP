package mcp

import (
	"bytes"
	"encoding/json"
)

const ProtocolVersion = "2025-03-26"

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// IsNotification reports whether the request lacks an id and therefore
// expects no response, per JSON-RPC 2.0.
func (r *JSONRPCRequest) IsNotification() bool {
	id := bytes.TrimSpace(r.ID)
	return len(id) == 0 || bytes.Equal(id, []byte("null"))
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ClientCapabilities struct {
	Elicitation *struct{} `json:"elicitation,omitempty"`
	Sampling    *struct{} `json:"sampling,omitempty"`
	Roots       *struct {
		ListChanged bool `json:"listChanged"`
	} `json:"roots,omitempty"`
}

type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

type InitializeResponse struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    struct {
		Tools struct {
			ListChanged bool `json:"listChanged"`
		} `json:"tools"`
		Resources struct {
			Subscribe   bool `json:"subscribe"`
			ListChanged bool `json:"listChanged"`
		} `json:"resources"`
	} `json:"capabilities"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema any            `json:"inputSchema"`
	Annotations map[string]any `json:"annotations,omitempty"`
	Meta        map[string]any `json:"_meta,omitempty"`
}

type ToolsListResponse struct {
	Tools []Tool `json:"tools"`
}

// Resource represents an MCP resource entry (e.g. a widget HTML page).
type MCPResource struct {
	URI      string `json:"uri"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType,omitempty"`
}

type ResourcesListResponse struct {
	Resources []MCPResource `json:"resources"`
}

type ResourcesReadRequest struct {
	URI string `json:"uri"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text"`
}

type ResourcesReadResponse struct {
	Contents []ResourceContent `json:"contents"`
}

type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolCallResponse struct {
	Content []any          `json:"content"`
	IsError bool           `json:"isError,omitempty"`
	Meta    map[string]any `json:"_meta,omitempty"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// JSON-RPC 2.0 error codes used by the server.
const (
	ErrCodeParse          = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternal       = -32603
)
