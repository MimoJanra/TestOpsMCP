package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
	"github.com/MimoJanra/TestOpsMCP/internal/session"
	"github.com/MimoJanra/TestOpsMCP/internal/tasks"
)

type HandlerFunc func(ctx context.Context, input json.RawMessage) (any, error)

// Typed wraps a typed handler into a HandlerFunc with auto-unmarshaling.
func Typed[T any](fn func(context.Context, T) (any, error)) HandlerFunc {
	return func(ctx context.Context, input json.RawMessage) (any, error) {
		var args T
		if err := json.Unmarshal(input, &args); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}
		return fn(ctx, args)
	}
}

type PromptArg struct {
	Name        string
	Description string
	Required    bool
}

type RegistryPromptMessage struct {
	Role string
	Text string
}

type RegistryPrompt struct {
	Name        string
	Description string
	Arguments   []PromptArg
	GetMessages func(args map[string]string) []RegistryPromptMessage
}

type Tool struct {
	Name        string
	Description string
	InputSchema any
	Handler     HandlerFunc
	// Annotations holds MCP tool annotations (readOnlyHint, destructiveHint, title, …).
	// If nil at registration time, autoAnnotate() fills it in during NewRegistry.
	Annotations map[string]any
	// Meta holds the _meta field for MCP tool list responses (e.g. widget resource URI).
	Meta map[string]any
}

// Resource represents an MCP resource served by this server (e.g. a widget HTML page).
type Resource struct {
	URI      string
	Name     string
	MimeType string
	// GetHTML returns the full HTML content with any dynamic content (e.g. inlined bundle) applied.
	GetHTML func() string
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
	taskStore   *tasks.Store

	// sessionTokens maps MCP session ID → user-provided API token.
	// Each SSE session (HTTP) or the fixed "stdio" session stores its own token
	// so concurrent users never overwrite each other.
	sessionTokens   map[string]string
	sessionTokensMu sync.RWMutex

	// resources holds MCP resources (e.g. widget HTML pages).
	resources   map[string]*Resource
	resourcesMu sync.RWMutex

	// prompts holds registered MCP prompt templates.
	prompts   map[string]*RegistryPrompt
	promptsMu sync.RWMutex
}

func NewRegistry(allureClient *allure.Client, logger *core.Logger) *Registry {
	r := &Registry{
		tools:         make(map[string]*Tool),
		allure:        allureClient,
		logger:        logger,
		sessionTokens: make(map[string]string),
		resources:     make(map[string]*Resource),
		prompts:       make(map[string]*RegistryPrompt),
		taskStore:     tasks.NewStore(),
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

	r.registerPrompts()
	r.registerLaunchTools()
	r.registerResultTools()
	r.registerTestCaseTools()
	r.registerTestCaseExtraTools()
	r.registerProjectTools()
	r.registerAnalyticsTools()
	r.registerBulkTools()
	r.registerRelationTools()
	r.registerWidgets()
	r.registerTaskTools()

	// Configuration tool for per-session token override.
	// Always registered so that:
	//  - users can supply a token when ALLURE_TOKEN is not set in config; and
	//  - users can override the server-wide token for their own session.
	if r.allure != nil {
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
			Annotations: map[string]any{
				"readOnlyHint":    false,
				"destructiveHint": false,
			},
			Handler: Typed(r.configureAllureToken),
		})
	}

	// Search + Execute tools for full OpenAPI coverage
	if r.opIndex != nil {
		r.register(&Tool{
			Name:        "search_testops_operations",
			Description: "Search for TestOps API operations by intent/keyword. Returns up to 10 matching operations with their IDs, paths, methods, and required parameters. Renders an interactive picker widget in Claude Desktop and claude.ai.",
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
			Meta: map[string]any{
				"ui": map[string]any{
					"resourceUri": "ui://widgets/action-picker",
				},
			},
			// autoAnnotate will classify search_ prefix as readOnly
			Handler: Typed(r.searchTestOpsOperations),
		})

		r.register(&Tool{
			Name: "execute_testops_operation",
			Description: "Execute a TestOps API operation by operation_id (from search_testops_operations results). " +
				"Handles path parameters, query parameters, and request bodies automatically. " +
				"For operations that require an array or a specific body structure, pass the value under the special key \"body\" " +
				"(e.g. {\"testCaseId\": 1, \"body\": [{\"customField\": {\"id\": 5}, \"values\": [{\"id\": 12}]}]}). " +
				"Named path/query parameters are matched by name from the spec; everything else is sent as the request body. " +
				"Renders result in an interactive widget.",
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
			Meta: map[string]any{
				"ui": map[string]any{
					"resourceUri": "ui://widgets/results-display",
				},
			},
			// Marked destructive: can execute any API operation including DELETE/PUT.
			Annotations: map[string]any{
				"readOnlyHint":    false,
				"destructiveHint": true,
			},
			Handler: Typed(r.executeTestOpsOperation),
		})
	}

	// Apply MCP tool annotations (readOnlyHint / destructiveHint) based on
	// naming conventions. Tools that set Annotations explicitly during
	// registration keep their own values; the rest get auto-classified here.
	r.mu.Lock()
	for _, tool := range r.tools {
		if tool.Annotations == nil {
			tool.Annotations = autoAnnotate(tool.Name)
		} else {
			// Back-fill any missing hint keys so the MCP response is always complete.
			if _, ok := tool.Annotations["readOnlyHint"]; !ok {
				defaults := autoAnnotate(tool.Name)
				tool.Annotations["readOnlyHint"] = defaults["readOnlyHint"]
				tool.Annotations["destructiveHint"] = defaults["destructiveHint"]
			}
		}
	}
	r.mu.Unlock()

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

type configureAllureTokenArgs struct {
	Token string `json:"token"`
}

func (r *Registry) configureAllureToken(ctx context.Context, args configureAllureTokenArgs) (any, error) {
	if args.Token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	sessID := session.IDFromContext(ctx)
	if sessID == "" {
		sessID = session.StdioID
	}

	r.sessionTokensMu.Lock()
	r.sessionTokens[sessID] = args.Token
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

// ClearSessionToken removes the stored API token for the given MCP session ID.
func (r *Registry) ClearSessionToken(sessID string) {
	r.sessionTokensMu.Lock()
	delete(r.sessionTokens, sessID)
	r.sessionTokensMu.Unlock()
}

// RegisterResource adds a resource to the registry.
func (r *Registry) RegisterResource(res *Resource) {
	r.resourcesMu.Lock()
	defer r.resourcesMu.Unlock()
	r.resources[res.URI] = res
}

// ListResources returns all registered resources.
func (r *Registry) ListResources() []*Resource {
	r.resourcesMu.RLock()
	defer r.resourcesMu.RUnlock()
	list := make([]*Resource, 0, len(r.resources))
	for _, res := range r.resources {
		list = append(list, res)
	}
	return list
}

// GetResource returns the resource with the given URI, or nil if not found.
func (r *Registry) GetResource(uri string) *Resource {
	r.resourcesMu.RLock()
	defer r.resourcesMu.RUnlock()
	return r.resources[uri]
}

// autoAnnotate returns MCP tool annotations derived from the tool's name.
//
// Naming conventions:
//   - get_* / list_* / search_* / suggest_* / validate_* → readOnly
//   - delete_* / bulk_delete_* / detach_* → destructive write
//   - everything else → non-destructive write
func autoAnnotate(name string) map[string]any {
	// Permanently destructive — deletes data that may be hard to recover.
	destructivePrefixes := []string{
		"delete_",
		"bulk_delete_",
		"detach_",
	}
	for _, p := range destructivePrefixes {
		if strings.HasPrefix(name, p) {
			return map[string]any{
				"readOnlyHint":    false,
				"destructiveHint": true,
			}
		}
	}

	// Pure reads — no side-effects on Allure data.
	readOnlyPrefixes := []string{
		"get_",
		"list_",
		"search_",
		"suggest_",
		"validate_",
	}
	for _, p := range readOnlyPrefixes {
		if strings.HasPrefix(name, p) {
			return map[string]any{
				"readOnlyHint":    true,
				"destructiveHint": false,
			}
		}
	}

	// Default: reversible write (create, update, set, add, remove, run, …).
	return map[string]any{
		"readOnlyHint":    false,
		"destructiveHint": false,
	}
}

// RegisterPrompt adds a prompt template to the registry.
func (r *Registry) RegisterPrompt(p *RegistryPrompt) {
	r.promptsMu.Lock()
	defer r.promptsMu.Unlock()
	r.prompts[p.Name] = p
}

// ListPrompts returns all registered prompt templates.
func (r *Registry) ListPrompts() []*RegistryPrompt {
	r.promptsMu.RLock()
	defer r.promptsMu.RUnlock()
	list := make([]*RegistryPrompt, 0, len(r.prompts))
	for _, p := range r.prompts {
		list = append(list, p)
	}
	return list
}

// GetPrompt returns the messages and description for the named prompt, with arguments applied.
func (r *Registry) GetPrompt(name string, args map[string]string) ([]RegistryPromptMessage, string, error) {
	r.promptsMu.RLock()
	p := r.prompts[name]
	r.promptsMu.RUnlock()
	if p == nil {
		return nil, "", fmt.Errorf("prompt not found: %s", name)
	}
	if args == nil {
		args = map[string]string{}
	}
	return p.GetMessages(args), p.Description, nil
}

// Search + Execute handlers

func (r *Registry) searchTestOpsOperations(ctx context.Context, req SearchRequest) (any, error) {
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

func (r *Registry) executeTestOpsOperation(ctx context.Context, req ExecuteRequest) (any, error) {
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
