package mcp

import (
	"bytes"
	"encoding/json"
)

const ProtocolVersion = "2025-11-25"

var supportedVersions = []string{"2025-11-25", "2025-06-18", "2025-03-26", "2024-11-05"}

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

type ServerCapabilities struct {
	Tools struct {
		ListChanged bool `json:"listChanged"`
	} `json:"tools"`
	Resources struct {
		Subscribe   bool `json:"subscribe"`
		ListChanged bool `json:"listChanged"`
	} `json:"resources"`
	Prompts struct {
		ListChanged bool `json:"listChanged"`
	} `json:"prompts"`
	Logging     *struct{} `json:"logging,omitempty"`
	Elicitation *struct{} `json:"elicitation,omitempty"`
}

type InitializeResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

// PaginatedRequest and PaginatedResponse support cursor-based pagination
// in tools/list, resources/list, and prompts/list.
type PaginatedRequest struct {
	Cursor string `json:"cursor,omitempty"`
}

type PaginatedResponse struct {
	NextCursor string `json:"nextCursor,omitempty"`
}

type ToolsListRequest struct {
	PaginatedRequest
}

type ToolsListResponse struct {
	Tools []Tool `json:"tools"`
	PaginatedResponse
}

type ResourcesListRequest struct {
	PaginatedRequest
}

type PromptsListRequest struct {
	PaginatedRequest
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema any            `json:"inputSchema"`
	Annotations map[string]any `json:"annotations,omitempty"`
	Meta        map[string]any `json:"_meta,omitempty"`
}

// Resource represents an MCP resource entry (e.g. a widget HTML page).
type MCPResource struct {
	URI      string `json:"uri"`
	Name     string `json:"name"`
	MimeType string `json:"mimeType,omitempty"`
}

type ResourcesListResponse struct {
	Resources []MCPResource `json:"resources"`
	PaginatedResponse
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

type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type PromptsListResponse struct {
	Prompts []Prompt `json:"prompts"`
	PaginatedResponse
}

type PromptGetRequest struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments,omitempty"`
}

type PromptMessage struct {
	Role    string      `json:"role"`
	Content TextContent `json:"content"`
}

type PromptGetResponse struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
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

// Elicitation types — server asks client/user to confirm or fill a form.
type ElicitRequest struct {
	Message         string          `json:"message"`
	RequestedSchema json.RawMessage `json:"requestedSchema,omitempty"`
}

type ElicitResult struct {
	Action  string          `json:"action"`  // "accept" | "reject" | "cancel"
	Content json.RawMessage `json:"content,omitempty"`
}

// Completion types for argument autocompletion (completion/complete).
type CompleteRequest struct {
	Ref      CompleteRef      `json:"ref"`
	Argument CompleteArgument `json:"argument"`
}

type CompleteRef struct {
	Type string `json:"type"` // "ref/prompt" | "ref/resource"
	Name string `json:"name"`
}

type CompleteArgument struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CompleteResult struct {
	Completion CompleteCompletion `json:"completion"`
}

type CompleteCompletion struct {
	Values  []string `json:"values"`
	HasMore bool     `json:"hasMore,omitempty"`
}

// JSON-RPC 2.0 error codes used by the server.
const (
	ErrCodeParse          = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternal       = -32603
)
