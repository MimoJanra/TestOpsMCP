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
	Components map[string]interface{}            `json:"components"`
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
	if op, ok := idx.operations[strings.TrimSpace(query)]; ok {
		return []*Operation{op} // an exact operation id
	}
	terms, verbs := parseSearchQuery(query)
	type scored struct {
		op    *Operation
		score int
	}
	var hits []scored
	for _, op := range idx.operations {
		if sc := scoreOperation(op, terms, verbs); sc > 0 {
			hits = append(hits, scored{op, sc})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].op.OperationID < hits[j].op.OperationID
	})
	results := make([]*Operation, len(hits))
	for i, h := range hits {
		results[i] = h.op
	}
	return results
}

// verbMethods maps intent verbs to the HTTP methods they usually mean.
var verbMethods = map[string][]string{
	"create": {"POST"}, "add": {"POST", "PUT"}, "new": {"POST"}, "run": {"POST"}, "upload": {"POST"},
	"get": {"GET"}, "list": {"GET"}, "find": {"GET"}, "search": {"GET", "POST"}, "read": {"GET"}, "show": {"GET"}, "download": {"GET"},
	"update": {"PATCH", "PUT"}, "edit": {"PATCH", "PUT"}, "change": {"PATCH", "PUT"}, "rename": {"PATCH", "PUT"}, "set": {"PATCH", "PUT", "POST"},
	"delete": {"DELETE"}, "remove": {"DELETE", "POST"},
}

var searchStopWords = map[string]bool{"a": true, "an": true, "the": true, "to": true, "of": true, "for": true, "in": true, "on": true, "by": true, "from": true, "with": true, "and": true, "or": true, "all": true}

// parseSearchQuery splits an intent like "create test case step" into
// content terms and verb-implied HTTP methods. The old search matched the
// whole phrase as one substring, so multi-word intents found nothing.
func parseSearchQuery(query string) (terms []string, methods map[string]bool) {
	methods = map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	}) {
		if searchStopWords[w] {
			continue
		}
		if ms, ok := verbMethods[w]; ok {
			for _, m := range ms {
				methods[m] = true
			}
			// Verbs still count as terms when they appear in ids/summaries
			// (e.g. "run", "search"), but aren't required to match.
			continue
		}
		terms = append(terms, singular(w))
	}
	if len(terms) == 0 { // query of only verbs/stop words: search the raw words
		for _, w := range strings.Fields(strings.ToLower(query)) {
			terms = append(terms, singular(w))
		}
	}
	return terms, methods
}

func singular(w string) string {
	switch {
	case len(w) > 4 && strings.HasSuffix(w, "ies"):
		return w[:len(w)-3] + "y"
	case len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss"):
		return w[:len(w)-1]
	}
	return w
}

// scoreOperation returns 0 unless every term appears somewhere in the
// operation; otherwise a relevance score weighted by where the terms match.
func scoreOperation(op *Operation, terms []string, methods map[string]bool) int {
	id := strings.ToLower(op.OperationID)
	summary := strings.ToLower(op.Summary)
	desc := strings.ToLower(op.Description)
	path := strings.ToLower(op.Path)
	compactPath := strings.NewReplacer("-", "", "_", "").Replace(path)
	tags := strings.ToLower(strings.Join(op.Tags, " "))
	segments := map[string]bool{}
	for _, seg := range strings.Split(compactPath, "/") {
		segments[singular(seg)] = true
	}
	score := 0
	for _, t := range terms {
		ts := 0
		if strings.Contains(path, t) || strings.Contains(compactPath, t) {
			ts += 10
		}
		if segments[t] { // "step" matches /step, not just /sharedstep
			ts += 8
		}
		if strings.Contains(summary, t) {
			ts += 8
		}
		if strings.Contains(id, t) {
			ts += 6
		}
		if strings.Contains(tags, t) {
			ts += 5
		}
		if strings.Contains(desc, t) {
			ts += 2
		}
		if ts == 0 {
			return 0
		}
		score += ts
	}
	if len(methods) > 0 && methods[strings.ToUpper(op.Method)] {
		score += 12
	}
	// Prefer v2 endpoints (project rule: v1 counterparts are often broken).
	if strings.Contains(path, "/v2/") {
		score += 3
	}
	// Shorter paths are usually the primary resource endpoint.
	score -= strings.Count(path, "/")
	return score
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
		Path:        path,
		Method:      strings.ToUpper(method),
		OperationID: operationID,
		Summary:     getStringValue(specMap, "summary"),
		Description: getStringValue(specMap, "description"),
		Tags:        getStringArray(specMap, "tags"),
		Responses:   make(map[string]interface{}),
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
