package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// SearchRequest represents a search query for operations
type SearchRequest struct {
	Intent string `json:"intent"`
	Limit  int    `json:"limit"`
}

// SearchResult represents a single search result
type SearchResult struct {
	OperationID string   `json:"operation_id"`
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Parameters  []struct {
		Name        string `json:"name"`
		In          string `json:"in"`
		Description string `json:"description"`
		Required    bool   `json:"required"`
	} `json:"parameters"`
}

// ExecuteRequest represents an execution request
type ExecuteRequest struct {
	OperationID string      `json:"operation_id"`
	Parameters  interface{} `json:"parameters"`
}

// buildSearchResults converts Operation slice to SearchResult slice
func buildSearchResults(ops []*Operation, limit int) []SearchResult {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if limit > len(ops) {
		limit = len(ops)
	}

	results := make([]SearchResult, limit)
	for i := 0; i < limit; i++ {
		op := ops[i]
		result := SearchResult{
			OperationID: op.OperationID,
			Path:        op.Path,
			Method:      op.Method,
			Summary:     op.Summary,
			Description: op.Description,
			Tags:        op.Tags,
			Parameters: make([]struct {
				Name        string `json:"name"`
				In          string `json:"in"`
				Description string `json:"description"`
				Required    bool   `json:"required"`
			}, len(op.Parameters)),
		}

		for j, p := range op.Parameters {
			result.Parameters[j] = struct {
				Name        string `json:"name"`
				In          string `json:"in"`
				Description string `json:"description"`
				Required    bool   `json:"required"`
			}{
				Name:        p.Name,
				In:          p.In,
				Description: p.Description,
				Required:    p.Required,
			}
		}

		results[i] = result
	}

	return results
}

// destructiveHTTPMethods are the methods executeOperation refuses to run
// without an explicit confirmation, since op.Method comes straight from the
// OpenAPI spec and can target any of 600+ operations, including deletes and
// overwrites, with no other server-side gate in front of it.
var destructiveHTTPMethods = map[string]bool{
	http.MethodDelete: true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
}

// executeOperation executes a discovered operation against TestOps API
func (r *Registry) executeOperation(ctx context.Context, op *Operation, params interface{}) (any, error) {
	var paramMap map[string]interface{}
	if params != nil {
		if pm, ok := params.(map[string]interface{}); ok {
			paramMap = pm
		}
	}

	if destructiveHTTPMethods[op.Method] {
		confirmed, _ := paramMap["confirm"].(bool)
		if !confirmed {
			return nil, fmt.Errorf(
				"operation %q uses %s and is destructive; re-call execute_testops_operation with parameters.confirm=true to proceed",
				op.OperationID, op.Method,
			)
		}
	}

	// Build URL with path parameters and query parameters
	reqURL := r.allure.GetBaseURL() + op.Path
	var pathParams map[string]interface{}
	var queryParams url.Values
	var bodyData interface{}

	if paramMap != nil {
		pathParams = make(map[string]interface{})
		queryParams = make(url.Values)

		// Special key "body" allows passing any value (object OR array) as the request body directly.
		if explicitBody, hasBody := paramMap["body"]; hasBody {
			bodyData = explicitBody
		}

		// unknownParams collects parameters not found in the spec's named parameters list.
		// They are used as the request body only if no explicit "body" key was provided.
		unknownParams := make(map[string]interface{})

		for name, value := range paramMap {
			if name == "body" || name == "confirm" {
				// "body" already handled above; "confirm" is the destructive-op ack, not API data.
				continue
			}
			var found bool
			for _, p := range op.Parameters {
				if p.Name == name {
					found = true
					switch p.In {
					case "path":
						pathParams[name] = value
					case "query":
						// Arrays repeat the parameter (?id=1&id=2), the form Spring expects.
						if arr, ok := value.([]interface{}); ok {
							for _, v := range arr {
								queryParams.Add(name, formatParamValue(v))
							}
						} else {
							queryParams.Set(name, formatParamValue(value))
						}
					case "body":
						bodyData = value
					}
					break
				}
			}
			if !found {
				unknownParams[name] = value
			}
		}

		// If no explicit body was provided but the operation has a requestBody schema,
		// use only the unrecognised parameters as the body (not path/query params too).
		if bodyData == nil && op.RequestBody != nil && len(unknownParams) > 0 {
			bodyData = unknownParams
		}
	}

	// Replace path parameters (URL-encode values to prevent path injection).
	for key, value := range pathParams {
		reqURL = strings.ReplaceAll(reqURL, "{"+key+"}", url.PathEscape(formatParamValue(value)))
	}

	// Add query parameters
	if len(queryParams) > 0 {
		reqURL += "?" + queryParams.Encode()
	}

	// Create HTTP request
	var body io.Reader
	if bodyData != nil {
		bodyBytes, err := json.Marshal(bodyData)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, op.Method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	if err := r.allure.SetAuthHeader(ctx, req); err != nil {
		return nil, fmt.Errorf("set auth header: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := r.allure.GetHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Check status code
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response.
	// When Content-Type is absent we still attempt JSON parsing so that
	// a body is never silently discarded (some proxies strip the header).
	var result interface{}
	if len(respBody) > 0 {
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" || strings.Contains(contentType, "application/json") {
			if err := json.Unmarshal(respBody, &result); err != nil {
				// JSON parsing failed – return raw body so callers can inspect it.
				result = map[string]interface{}{
					"raw_response": string(respBody),
					"status_code":  resp.StatusCode,
				}
			}
		} else if isTextContentType(contentType) {
			result = map[string]interface{}{
				"raw_response": string(respBody),
				"status_code":  resp.StatusCode,
				"content_type": contentType,
			}
		} else {
			// Binary (e.g. attachment content): a Go string of arbitrary bytes
			// is mangled into U+FFFD by JSON encoding, so return base64.
			result = map[string]interface{}{
				"content_base64": base64.StdEncoding.EncodeToString(respBody),
				"size":           len(respBody),
				"status_code":    resp.StatusCode,
				"content_type":   contentType,
			}
		}
	} else {
		result = map[string]interface{}{
			"status_code": resp.StatusCode,
		}
	}

	return result, nil
}

// formatParamValue renders a JSON-decoded value for a URL. JSON numbers decode
// to float64, and %v prints large ones in exponent form (1.78e+12), which the
// API rejects — timestamps and ids must stay plain integers.
func formatParamValue(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if t == math.Trunc(t) && math.Abs(t) < 1e18 {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}

func isTextContentType(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.HasPrefix(ct, "text/") || strings.Contains(ct, "json") ||
		strings.Contains(ct, "xml") || strings.Contains(ct, "csv") || strings.Contains(ct, "javascript")
}
