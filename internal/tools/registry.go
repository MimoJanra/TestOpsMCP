package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
	"github.com/MimoJanra/TestOpsMCP/internal/session"
)

type HandlerFunc func(ctx context.Context, input json.RawMessage) (any, error)

type Tool struct {
	Name        string
	Description string
	InputSchema any
	Handler     HandlerFunc
}

// Registry holds the set of tools exposed by the MCP server.
// Tools are registered once during NewRegistry and must not be mutated
// afterwards; ListTools therefore returns shared pointers without copying.
type Registry struct {
	tools       map[string]*Tool
	allure      *allure.Client
	logger      *core.Logger
	mu          sync.RWMutex
	opIndex     *OperationsIndex
	openAPISpec *OpenAPISpec

	// sessionTokens maps MCP session ID → user-provided API token.
	// Each SSE session (HTTP) or the fixed "stdio" session stores its own token
	// so concurrent users never overwrite each other.
	sessionTokens   map[string]string
	sessionTokensMu sync.RWMutex
}

func NewRegistry(allureClient *allure.Client, logger *core.Logger) *Registry {
	r := &Registry{
		tools:         make(map[string]*Tool),
		allure:        allureClient,
		logger:        logger,
		sessionTokens: make(map[string]string),
	}

	// Load OpenAPI spec and build operations index
	if specPath, err := FindSpecFile(); err == nil {
		if spec, err := LoadOpenAPI(specPath); err == nil {
			r.openAPISpec = spec
			if idx, err := BuildOperationsIndex(spec); err == nil {
				r.opIndex = idx
				logger.Info("loaded OpenAPI spec", map[string]any{
					"spec_file":  specPath,
					"operations": len(idx.ListAll()),
				})
			} else {
				logger.Info("failed to build operations index", map[string]any{"error": err.Error()})
			}
		} else {
			logger.Info("failed to load OpenAPI spec", map[string]any{"error": err.Error()})
		}
	}

	// Warn if no Allure token is configured
	if allureClient != nil && !allureClient.HasToken() {
		logger.Info("ALLURE_TOKEN not configured - each user must provide their token in Claude Desktop config", nil)
	}

	// Set up callback for session token if client exists.
	// The callback receives the request context so it can look up the token
	// for the specific MCP session that originated the request.
	if allureClient != nil {
		allureClient.SetSessionTokenFunc(func(ctx context.Context) string {
			return r.getSessionToken(session.IDFromContext(ctx))
		})
	}

	r.registerLaunchTools()
	r.registerResultTools()
	r.registerTestCaseTools()
	r.registerTestCaseExtraTools()
	r.registerProjectTools()
	r.registerAnalyticsTools()
	r.registerBulkTools()
	r.registerRelationTools()

	// Configuration tool for session token (if no token in config)
	if r.allure != nil && !r.allure.HasToken() {
		r.register(&Tool{
			Name:        "configure_allure_token",
			Description: "⚠️ SECURITY WARNING: Configure your Allure API token for this chat session. Only use if you trust this channel. Token will NOT be saved after session ends. For production, use ALLURE_TOKEN environment variable instead.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"token": map[string]any{
						"type":        "string",
						"description": "Your Allure TestOps API token",
					},
				},
				"required": []string{"token"},
			},
			Handler: r.configureAllureToken,
		})
	}

	// Search + Execute tools for full OpenAPI coverage
	if r.opIndex != nil {
		r.register(&Tool{
			Name:        "search_testops_operations",
			Description: "Search for TestOps API operations by intent/keyword. Returns up to 10 matching operations with their IDs, paths, methods, and required parameters.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"intent": map[string]any{
						"type":        "string",
						"description": "What you want to do (e.g., 'create project', 'list launches', 'get test results')",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Max results to return (1-100, default 10)",
						"default":     10,
					},
				},
				"required": []string{"intent"},
			},
			Handler: r.searchTestOpsOperations,
		})

		r.register(&Tool{
			Name: "execute_testops_operation",
			Description: "Execute a TestOps API operation by operation_id (from search_testops_operations results). " +
				"Handles path parameters, query parameters, and request bodies automatically. " +
				"For operations that require an array or a specific body structure, pass the value under the special key \"body\" " +
				"(e.g. {\"testCaseId\": 1, \"body\": [{\"customField\": {\"id\": 5}, \"values\": [{\"id\": 12}]}]}). " +
				"Named path/query parameters are matched by name from the spec; everything else is sent as the request body.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"operation_id": map[string]any{
						"type":        "string",
						"description": "The operation_id from search_testops_operations results",
					},
					"parameters": map[string]any{
						"type": "object",
						"description": "Parameters for the operation. Named path/query params are matched automatically. " +
							"Use the special key \"body\" to pass an array or exact body object.",
					},
				},
				"required": []string{"operation_id"},
			},
			Handler: r.executeTestOpsOperation,
		})
	}

	return r
}

func (r *Registry) register(tool *Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = tool
}

func (r *Registry) GetTool(name string) *Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tools[name]
}

func (r *Registry) ListTools() []*Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Token configuration handler

func (r *Registry) configureAllureToken(ctx context.Context, input json.RawMessage) (any, error) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.Token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	sessID := session.IDFromContext(ctx)
	if sessID == "" {
		sessID = session.StdioID
	}

	r.sessionTokensMu.Lock()
	r.sessionTokens[sessID] = req.Token
	r.sessionTokensMu.Unlock()

	r.logger.Info("session token configured", map[string]any{
		"session": sessID,
		"warning": "token set in chat session - not saved after session ends",
	})

	return map[string]any{
		"status":  "configured",
		"warning": "⚠️ Token stored only for this session. It will be lost when you close the conversation. For persistent setup, use ALLURE_TOKEN environment variable.",
	}, nil
}

// SetSessionToken stores an Allure API token for the given MCP session ID.
// Called by the MCP server when a client passes X-Allure-Token on the SSE connection.
func (r *Registry) SetSessionToken(sessID, token string) {
	r.sessionTokensMu.Lock()
	r.sessionTokens[sessID] = token
	r.sessionTokensMu.Unlock()
}

// getSessionToken returns the API token stored for the given MCP session ID, or "" if none.
func (r *Registry) getSessionToken(sessID string) string {
	r.sessionTokensMu.RLock()
	defer r.sessionTokensMu.RUnlock()
	return r.sessionTokens[sessID]
}

// Search + Execute handlers

func (r *Registry) searchTestOpsOperations(ctx context.Context, input json.RawMessage) (any, error) {
	var req SearchRequest
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.Intent == "" {
		return nil, fmt.Errorf("intent is required")
	}

	if r.opIndex == nil {
		return nil, fmt.Errorf("operations index not available")
	}

	ops := r.opIndex.Search(req.Intent)
	results := buildSearchResults(ops, req.Limit)

	r.logger.Info("search testops operations", map[string]any{
		"intent":  req.Intent,
		"results": len(results),
	})

	return map[string]any{
		"intent":  req.Intent,
		"results": results,
		"total":   len(r.opIndex.ListAll()),
	}, nil
}

func (r *Registry) executeTestOpsOperation(ctx context.Context, input json.RawMessage) (any, error) {
	var req ExecuteRequest
	if err := json.Unmarshal(input, &req); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if req.OperationID == "" {
		return nil, fmt.Errorf("operation_id is required")
	}

	if r.opIndex == nil {
		return nil, fmt.Errorf("operations index not available")
	}

	op := r.opIndex.Get(req.OperationID)
	if op == nil {
		return nil, fmt.Errorf("operation not found: %s", req.OperationID)
	}

	r.logger.Info("execute testops operation", map[string]any{
		"operation_id": req.OperationID,
		"path":         op.Path,
		"method":       op.Method,
	})

	result, err := r.executeOperation(ctx, op, req.Parameters)
	if err != nil {
		r.logger.Error("execute testops operation", err, map[string]any{
			"operation_id": req.OperationID,
		})
		return nil, fmt.Errorf("execute operation: %w", err)
	}

	return result, nil
}
