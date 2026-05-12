package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Operation represents a single OpenAPI operation
type Operation struct {
	Path        string                 `json:"path"`
	Method      string                 `json:"method"`
	OperationID string                 `json:"operation_id"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Parameters  []OperationParameter   `json:"parameters"`
	RequestBody *OperationRequestBody  `json:"request_body"`
	Responses   map[string]interface{} `json:"responses"`
	Tags        []string               `json:"tags"`
}

// OperationParameter represents a parameter in an operation
type OperationParameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Schema      interface{} `json:"schema"`
}

// OperationRequestBody represents the request body schema
type OperationRequestBody struct {
	Required bool        `json:"required"`
	Content  interface{} `json:"content"`
}

// OpenAPISpec represents parsed OpenAPI specification
type OpenAPISpec struct {
	Paths      map[string]map[string]interface{} `json:"paths"`
	Components map[string]interface{}             `json:"components"`
}

// OperationsIndex holds searchable operation index
type OperationsIndex struct {
	operations map[string]*Operation // key: operation_id
	byTag      map[string][]*Operation
}

// LoadOpenAPI loads and parses the testops.json OpenAPI spec
func LoadOpenAPI(specPath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("read spec file: %w", err)
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}

	return &spec, nil
}

// BuildOperationsIndex creates a searchable index from OpenAPI spec
func BuildOperationsIndex(spec *OpenAPISpec) (*OperationsIndex, error) {
	idx := &OperationsIndex{
		operations: make(map[string]*Operation),
		byTag:      make(map[string][]*Operation),
	}

	if spec.Paths == nil {
		return idx, nil
	}

	for path, methods := range spec.Paths {
		if methods == nil {
			continue
		}

		// Methods are stored as: GET, POST, PUT, PATCH, DELETE, etc.
		for method, methodSpec := range methods {
			if methodSpec == nil {
				continue
			}

			op := parseOperation(path, method, methodSpec)
			if op == nil {
				continue
			}

			idx.operations[op.OperationID] = op

			// Index by tags for easier searching
			for _, tag := range op.Tags {
				idx.byTag[tag] = append(idx.byTag[tag], op)
			}
		}
	}

	return idx, nil
}

// Search finds operations matching the query string
func (idx *OperationsIndex) Search(query string) []*Operation {
	query = strings.ToLower(query)
	var results []*Operation

	for _, op := range idx.operations {
		if matchesQuery(op, query) {
			results = append(results, op)
		}
	}

	// Sort by relevance (exact matches first)
	sort.Slice(results, func(i, j int) bool {
		iScore := scoreMatch(results[i], query)
		jScore := scoreMatch(results[j], query)
		return iScore > jScore
	})

	return results
}

// Get retrieves a specific operation by ID
func (idx *OperationsIndex) Get(operationID string) *Operation {
	return idx.operations[operationID]
}

// ListAll returns all operations
func (idx *OperationsIndex) ListAll() []*Operation {
	ops := make([]*Operation, 0, len(idx.operations))
	for _, op := range idx.operations {
		ops = append(ops, op)
	}
	sort.Slice(ops, func(i, j int) bool {
		return ops[i].OperationID < ops[j].OperationID
	})
	return ops
}

// Helper functions

func parseOperation(path string, method string, methodSpec interface{}) *Operation {
	specMap, ok := methodSpec.(map[string]interface{})
	if !ok {
		return nil
	}

	operationID, _ := specMap["operationId"].(string)
	if operationID == "" {
		return nil
	}

	op := &Operation{
		Path:       path,
		Method:     strings.ToUpper(method),
		OperationID: operationID,
		Summary:    getStringValue(specMap, "summary"),
		Description: getStringValue(specMap, "description"),
		Tags:       getStringArray(specMap, "tags"),
		Responses:  make(map[string]interface{}),
	}

	// Parse parameters
	if params, ok := specMap["parameters"].([]interface{}); ok {
		for _, p := range params {
			if param := parseParameter(p); param != nil {
				op.Parameters = append(op.Parameters, *param)
			}
		}
	}

	// Parse request body
	if rb, ok := specMap["requestBody"]; ok {
		op.RequestBody = parseRequestBody(rb)
	}

	// Parse responses
	if responses, ok := specMap["responses"].(map[string]interface{}); ok {
		op.Responses = responses
	}

	return op
}

func parseParameter(paramData interface{}) *OperationParameter {
	paramMap, ok := paramData.(map[string]interface{})
	if !ok {
		return nil
	}

	param := &OperationParameter{
		Name:        getStringValue(paramMap, "name"),
		In:          getStringValue(paramMap, "in"),
		Description: getStringValue(paramMap, "description"),
	}

	if required, ok := paramMap["required"].(bool); ok {
		param.Required = required
	}

	if schema, ok := paramMap["schema"]; ok {
		param.Schema = schema
	}

	if param.Name == "" {
		return nil
	}

	return param
}

func parseRequestBody(rbData interface{}) *OperationRequestBody {
	rbMap, ok := rbData.(map[string]interface{})
	if !ok {
		return nil
	}

	rb := &OperationRequestBody{
		Content: make(map[string]interface{}),
	}

	if required, ok := rbMap["required"].(bool); ok {
		rb.Required = required
	}

	if content, ok := rbMap["content"]; ok {
		rb.Content = content
	}

	return rb
}

func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringArray(m map[string]interface{}, key string) []string {
	if arr, ok := m[key].([]interface{}); ok {
		var result []string
		for _, v := range arr {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

func matchesQuery(op *Operation, query string) bool {
	if strings.Contains(strings.ToLower(op.OperationID), query) {
		return true
	}
	if strings.Contains(strings.ToLower(op.Summary), query) {
		return true
	}
	if strings.Contains(strings.ToLower(op.Description), query) {
		return true
	}
	if strings.Contains(strings.ToLower(op.Path), query) {
		return true
	}
	for _, tag := range op.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func scoreMatch(op *Operation, query string) int {
	score := 0
	if strings.Contains(strings.ToLower(op.OperationID), query) {
		score += 100
	}
	if strings.HasPrefix(strings.ToLower(op.Summary), query) {
		score += 50
	}
	if strings.Contains(strings.ToLower(op.Summary), query) {
		score += 25
	}
	if strings.Contains(strings.ToLower(op.Path), query) {
		score += 10
	}
	return score
}

// FindSpecFile looks for testops.json in spec folder and common locations
func FindSpecFile() (string, error) {
	locations := []string{
		"spec/testops.json",
		"./spec/testops.json",
		"testops.json",
		"./testops.json",
		"../testops.json",
		"../../testops.json",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			abs, err := filepath.Abs(loc)
			if err == nil {
				return abs, nil
			}
			return loc, nil
		}
	}

	// Try to find from current working directory
	cwd, err := os.Getwd()
	if err == nil {
		path := filepath.Join(cwd, "spec", "testops.json")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("testops.json not found (looked in spec/ and common locations)")
}
