package mcp

import (
	"encoding/json"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type InitializeRequest struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    struct{} `json:"capabilities"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

type InitializeResponse struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    struct {
		Tools struct{} `json:"tools"`
	} `json:"capabilities"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolsListResponse struct {
	Tools []Tool `json:"tools"`
}

type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolCallResponse struct {
	Content []interface{} `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
