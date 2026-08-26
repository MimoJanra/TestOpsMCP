package allure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

type jwtEntry struct {
	jwt       string
	expiresAt time.Time
}

type Client struct {
	baseURL    string
	userToken  string
	httpClient *http.Client

	mu              sync.Mutex
	jwtCache        map[string]jwtEntry // keyed by API token — each user's JWT is stored separately
	getSessionToken func(ctx context.Context) string
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		userToken: token,
		jwtCache:  make(map[string]jwtEntry),
		httpClient: &http.Client{
			Timeout: timeout,
			Jar:     jar,
		},
	}
}

func (c *Client) SetSessionTokenFunc(fn func(ctx context.Context) string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.getSessionToken = fn
}

// resolveAPIToken returns the API token to use for this request: the per-session
// token (if set by the user via configure_allure_token) or the server-wide token.
// The function-pointer field is read under the mutex to avoid a data race with
// SetSessionTokenFunc, which writes the field under the same mutex.
func (c *Client) resolveAPIToken(ctx context.Context) string {
	c.mu.Lock()
	fn := c.getSessionToken
	c.mu.Unlock()
	if fn != nil {
		if t := fn(ctx); t != "" {
			return t
		}
	}
	return c.userToken
}

func (c *Client) getJWTToken(ctx context.Context) (string, error) {
	apiToken := c.resolveAPIToken(ctx)
	if apiToken == "" {
		return "", fmt.Errorf("no token configured - set ALLURE_TOKEN env var or use configure_allure_token tool")
	}

	// Fast path: check cache under the lock.
	c.mu.Lock()
	if entry, ok := c.jwtCache[apiToken]; ok && time.Now().Before(entry.expiresAt) {
		c.mu.Unlock()
		return entry.jwt, nil
	}
	c.mu.Unlock()

	// Cache miss – fetch a new JWT outside the lock so we don't block other
	// goroutines. Two concurrent goroutines may both reach here for the same
	// apiToken; that results in a harmless double-fetch. After the fetch we
	// re-check the cache before writing so a race winner's entry is not
	// needlessly overwritten with a stale result.
	jwt, expiresIn, err := c.fetchJWT(ctx, apiToken)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	if expiresIn <= 0 {
		expiresAt = time.Now().Add(1 * time.Hour)
	}

	c.mu.Lock()
	// Re-check: another goroutine may have already populated the cache.
	if existing, ok := c.jwtCache[apiToken]; ok && time.Now().Before(existing.expiresAt) {
		c.mu.Unlock()
		return existing.jwt, nil
	}
	c.jwtCache[apiToken] = jwtEntry{jwt: jwt, expiresAt: expiresAt}
	c.mu.Unlock()

	return jwt, nil
}

func (c *Client) fetchJWT(ctx context.Context, apiToken string) (string, int, error) {
	values := url.Values{}
	values.Set("grant_type", "apitoken")
	values.Set("scope", "openid")
	values.Set("token", apiToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/uaa/oauth/token"), strings.NewReader(values.Encode()))
	if err != nil {
		return "", 0, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Expect", "")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, errFromResponse(resp)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("decode response: %w", err)
	}

	return result.AccessToken, result.ExpiresIn, nil
}

func (c *Client) CreateLaunch(ctx context.Context, projectID int64, launchName string) (*LaunchResponse, error) {
	var result LaunchResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/rs/launch", LaunchCreateRequest{
		Name:      launchName,
		ProjectID: projectID,
	}, &result, []int{http.StatusOK, http.StatusCreated}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetLaunchStatus(ctx context.Context, launchID int64) (interface{}, error) {
	resp, err := c.doRaw(ctx, http.MethodGet, fmt.Sprintf("/api/rs/launch/%d", launchID), nil, []int{http.StatusOK}...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result LaunchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Status, nil
}

func (c *Client) GetLaunchStatistics(ctx context.Context, launchID int64) (*StatisticsResponse, error) {
	resp, err := c.doRaw(ctx, http.MethodGet, fmt.Sprintf("/api/rs/launch/%d/statistic", launchID), nil, []int{http.StatusOK}...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// The API returns a JSON array of {status, count} items.
	var items []StatisticItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var result StatisticsResponse
	for _, item := range items {
		switch strings.ToLower(item.Status) {
		case "passed":
			result.Passed += item.Count
		case "failed":
			result.Failed += item.Count
		case "broken":
			result.Broken += item.Count
		case "skipped":
			result.Skipped += item.Count
		default:
			result.Unknown += item.Count
		}
		result.Total += item.Count
	}

	return &result, nil
}

func (c *Client) CloseLaunch(ctx context.Context, launchID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/launch/%d/close", launchID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) ReopenLaunch(ctx context.Context, launchID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/launch/%d/reopen", launchID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) ListLaunches(ctx context.Context, projectID int64, page, size int) (*LaunchListResponse, error) {
	url := fmt.Sprintf("/api/launch?projectId=%d&page=%d&size=%d", projectID, page, size)
	var result LaunchListResponse
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetLaunchDetails(ctx context.Context, launchID int64) (*LaunchDetailsResponse, error) {
	var result LaunchDetailsResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/launch/%d", launchID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ListTestResults(ctx context.Context, launchID int64, status string, page, size int) (*TestResultListResponse, error) {
	url := fmt.Sprintf("/api/testresult?launchId=%d&page=%d&size=%d", launchID, page, size)
	if status != "" {
		url += fmt.Sprintf("&status=%s", status)
	}

	var result TestResultListResponse
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetTestResult(ctx context.Context, testResultID int64) (*TestResultDetailsResponse, error) {
	var result TestResultDetailsResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testresult/%d", testResultID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) AssignTestResult(ctx context.Context, testResultID int64, username string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testresult/%d/assign", testResultID), AssignTestResultRequest{Username: username}, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}...)
}

func (c *Client) MuteTestResult(ctx context.Context, testResultID int64, reason string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testresult/%d/mute", testResultID), MuteTestResultRequest{Reason: reason}, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}...)
}

func (c *Client) ListTestCases(ctx context.Context, projectID int64, page, size int) (*TestCaseListResponse, error) {
	url := fmt.Sprintf("/api/testcase?projectId=%d&page=%d&size=%d", projectID, page, size)
	var result TestCaseListResponse
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetTestCase(ctx context.Context, testCaseID int64) (*TestCaseDetailsResponse, error) {
	var result TestCaseDetailsResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetTestCaseOverview(ctx context.Context, testCaseID int64) (map[string]any, error) {
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/overview", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) RunTestCase(ctx context.Context, testCaseID, launchID int64) error {
	return c.doRequest(ctx, http.MethodPost, "/api/testcase/bulk/run/existing", RunTestCaseRequest{
		TestCaseIds: []int64{testCaseID},
		LaunchId:    launchID,
	}, []int{http.StatusOK, http.StatusAccepted}...)
}

func (c *Client) ListProjects(ctx context.Context, page, size int) (*ProjectListResponse, error) {
	url := fmt.Sprintf("/api/project?page=%d&size=%d", page, size)
	var result ProjectListResponse
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProject(ctx context.Context, projectID int64) (*ProjectDetailsResponse, error) {
	var result ProjectDetailsResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/project/%d", projectID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetProjectStats(ctx context.Context, projectID int64) (*ProjectStatsResponse, error) {
	var result ProjectStatsResponse
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/project/%d/stats", projectID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetLaunchTrendAnalytics(ctx context.Context, projectID int64) ([]TrendData, error) {
	var result []TrendData
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/analytic/%d/statistic_trend", projectID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetLaunchDurationAnalytics(ctx context.Context, projectID int64) (interface{}, error) {
	var result interface{}
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/analytic/%d/launch_duration_histogram", projectID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetTestSuccessRateAnalytics(ctx context.Context, projectID int64) (interface{}, error) {
	var result interface{}
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/analytic/%d/tc_success_rate", projectID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) CreateTestCase(ctx context.Context, projectID int64, name, description string) (*TestCaseDetailsResponse, error) {
	var result TestCaseDetailsResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/testcase", CreateTestCaseRequest{
		Name:        name,
		ProjectID:   projectID,
		Description: description,
	}, &result, []int{http.StatusOK, http.StatusCreated}...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateTestCase(ctx context.Context, testCaseID int64, req UpdateTestCaseRequest) error {
	req.ID = testCaseID
	return c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/testcase/%d", testCaseID), req, []int{http.StatusOK, http.StatusNoContent}...)
}

// GetTestCaseCustomFields returns all custom field values for a test case.
func (c *Client) GetTestCaseCustomFields(ctx context.Context, testCaseID int64) ([]CustomFieldWithValuesDto, error) {
	var result []CustomFieldWithValuesDto
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/cfv", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateTestCaseCustomFields updates custom field values for a test case via PATCH /api/testcase/{id}/cfv.
func (c *Client) UpdateTestCaseCustomFields(ctx context.Context, testCaseID int64, fields []CustomFieldWithValuesDto) error {
	return c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/testcase/%d/cfv", testCaseID), fields, []int{http.StatusOK, http.StatusNoContent}...)
}

// ListCustomFieldValues lists the valid values defined for a custom field within a project
// (e.g. the allowed Priority or Section options), so callers can pick a real value ID
// instead of guessing one.
func (c *Client) ListCustomFieldValues(ctx context.Context, projectID, customFieldID int64, query string, page, size int) (map[string]any, error) {
	u := fmt.Sprintf("/api/project/%d/cfv?customFieldId=%d&page=%d&size=%d", projectID, customFieldID, page, size)
	if query != "" {
		u += "&query=" + url.QueryEscape(query)
	}
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// ─── Tags ────────────────────────────────────────────────────────────────────

// GetTestCaseTags returns the tags of a test case.
func (c *Client) GetTestCaseTags(ctx context.Context, testCaseID int64) ([]TestTagDto, error) {
	var result []TestTagDto
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/tag", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// SetTestCaseTags replaces all tags on a test case.
func (c *Client) SetTestCaseTags(ctx context.Context, testCaseID int64, tags []TestTagDto) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/tag", testCaseID), tags, []int{http.StatusOK, http.StatusNoContent}...)
}

// CreateTestTag creates a new project-wide tag via POST /api/tag and returns it (with its new ID).
func (c *Client) CreateTestTag(ctx context.Context, name string) (TestTagDto, error) {
	var result TestTagDto
	if err := c.doJSON(ctx, http.MethodPost, "/api/tag", TestTagDto{Name: name}, &result, []int{http.StatusOK}...); err != nil {
		return TestTagDto{}, err
	}
	return result, nil
}

// ─── Issues ───────────────────────────────────────────────────────────────────

// GetTestCaseIssues returns the issues linked to a test case.
func (c *Client) GetTestCaseIssues(ctx context.Context, testCaseID int64) ([]IssueDto, error) {
	var result []IssueDto
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/issue", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// SetTestCaseIssues replaces all issues linked to a test case.
func (c *Client) SetTestCaseIssues(ctx context.Context, testCaseID int64, issues []IssueDto) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/issue", testCaseID), issues, []int{http.StatusOK, http.StatusNoContent}...)
}

// ─── Examples (parametrized) ──────────────────────────────────────────────────

// GetTestCaseExamples returns the parametrized examples of a test case.
func (c *Client) GetTestCaseExamples(ctx context.Context, testCaseID int64) ([][]TestCaseExampleParam, error) {
	var result [][]TestCaseExampleParam
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/example", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// SetTestCaseExamples replaces the parametrized examples of a test case.
func (c *Client) SetTestCaseExamples(ctx context.Context, testCaseID int64, examples [][]TestCaseExampleParam) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/example", testCaseID), examples, []int{http.StatusOK, http.StatusNoContent}...)
}

// ─── Versions ─────────────────────────────────────────────────────────────────

// ListTestCaseVersions returns all versions of a test case.
func (c *Client) ListTestCaseVersions(ctx context.Context, testCaseID int64) ([]TestCaseVersionDto, error) {
	var result []TestCaseVersionDto
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/version", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateTestCaseVersion creates a named version snapshot of a test case.
func (c *Client) CreateTestCaseVersion(ctx context.Context, testCaseID int64, req TestCaseVersionCreateRequest) (*TestCaseVersionDto, error) {
	var result TestCaseVersionDto
	if err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/version", testCaseID), req, &result, []int{http.StatusOK, http.StatusCreated}...); err != nil {
		return nil, err
	}
	return &result, nil
}

// RestoreTestCaseVersion restores a test case to a specific version.
func (c *Client) RestoreTestCaseVersion(ctx context.Context, versionID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/version/%d/restore", versionID), nil, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}...)
}

// ─── Attachments ─────────────────────────────────────────────────────────────

// GetTestCaseAttachments returns the attachments of a test case.
func (c *Client) GetTestCaseAttachments(ctx context.Context, testCaseID int64, page, size int) (*TestCaseAttachmentListResponse, error) {
	u := fmt.Sprintf("/api/testcase/attachment?testCaseId=%d&page=%d&size=%d", testCaseID, page, size)
	var result TestCaseAttachmentListResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteTestCaseAttachment deletes a test case attachment by ID.
func (c *Client) DeleteTestCaseAttachment(ctx context.Context, attachmentID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/testcase/attachment/%d", attachmentID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

// ─── Search & filtered lists ─────────────────────────────────────────────────

// SearchTestCases finds test cases by AQL query.
func (c *Client) SearchTestCases(ctx context.Context, projectID int64, rql string, page, size int) (*TestCaseListResponse, error) {
	u := fmt.Sprintf("/api/testcase/__search?projectId=%d&rql=%s&page=%d&size=%d",
		projectID, url.QueryEscape(rql), page, size)
	var result TestCaseListResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListDeletedTestCases returns deleted test cases for a project.
func (c *Client) ListDeletedTestCases(ctx context.Context, projectID int64, page, size int) (*TestCaseListResponse, error) {
	u := fmt.Sprintf("/api/testcase/deleted?projectId=%d&page=%d&size=%d", projectID, page, size)
	var result TestCaseListResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

// ─── Scenario ─────────────────────────────────────────────────────────────────

// DeleteTestCaseScenario removes the entire scenario from a test case.
func (c *Client) DeleteTestCaseScenario(ctx context.Context, testCaseID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/testcase/%d/scenario", testCaseID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

// GetTestCaseSteps returns the normalized step list for a test case.
func (c *Client) GetTestCaseSteps(ctx context.Context, testCaseID int64) (map[string]any, error) {
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/step", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// MoveTestCaseStep moves a scenario step to a new position.
func (c *Client) MoveTestCaseStep(ctx context.Context, stepID int64, pos StepPositionDto) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/step/%d/move", stepID), pos, []int{http.StatusOK, http.StatusNoContent}...)
}

// CopyTestCaseStep copies a scenario step to a new position.
func (c *Client) CopyTestCaseStep(ctx context.Context, stepID int64, pos StepPositionDto) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/step/%d/copy", stepID), pos, []int{http.StatusOK, http.StatusNoContent}...)
}

// ─── Relations ────────────────────────────────────────────────────────────────

// GetTestCaseRelations returns test-case-to-test-case relations.
func (c *Client) GetTestCaseRelations(ctx context.Context, testCaseID int64) ([]RelationDto, error) {
	var result []RelationDto
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/relation", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// SetTestCaseRelations replaces all test-case-to-test-case relations.
func (c *Client) SetTestCaseRelations(ctx context.Context, testCaseID int64, relations []RelationDto) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/relation", testCaseID), relations, []int{http.StatusOK, http.StatusNoContent}...)
}

// ─── Muted test cases ─────────────────────────────────────────────────────────

// ListMutedTestCases returns muted test cases for a project.
func (c *Client) ListMutedTestCases(ctx context.Context, projectID int64, page, size int) (*TestCaseListResponse, error) {
	u := fmt.Sprintf("/api/testcase/muted?projectId=%d&page=%d&size=%d", projectID, page, size)
	var result TestCaseListResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

// ─── Audit log ────────────────────────────────────────────────────────────────

// GetTestCaseAudit returns audit log entries for a test case.
func (c *Client) GetTestCaseAudit(ctx context.Context, testCaseID int64, page, size int) (*TestCaseAuditListResponse, error) {
	u := fmt.Sprintf("/api/testcase/audit?testCaseId=%d&page=%d&size=%d", testCaseID, page, size)
	var result TestCaseAuditListResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return &result, nil
}

// ─── Query validation & suggest ───────────────────────────────────────────────

// ValidateTestCaseQuery validates an AQL query without running it.
func (c *Client) ValidateTestCaseQuery(ctx context.Context, projectID int64, rql string) (map[string]any, error) {
	u := fmt.Sprintf("/api/testcase/query/validate?projectId=%d&rql=%s", projectID, url.QueryEscape(rql))
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// SuggestTestCases returns test case suggestions for a query string.
func (c *Client) SuggestTestCases(ctx context.Context, projectID int64, query string, page, size int) (map[string]any, error) {
	u := fmt.Sprintf("/api/testcase/suggest?projectId=%d&query=%s&page=%d&size=%d",
		projectID, url.QueryEscape(query), page, size)
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// ─── Workflow ─────────────────────────────────────────────────────────────────

// GetTestCaseWorkflow returns the workflow for a test case.
func (c *Client) GetTestCaseWorkflow(ctx context.Context, testCaseID int64) (map[string]any, error) {
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/workflow", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// ─── Test keys ────────────────────────────────────────────────────────────────

// GetTestCaseKeys returns the integration test keys for a test case.
func (c *Client) GetTestCaseKeys(ctx context.Context, testCaseID int64) ([]TestKeyDto, error) {
	var result []TestKeyDto
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/testkey", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// SetTestCaseKeys replaces all test keys for a test case.
func (c *Client) SetTestCaseKeys(ctx context.Context, testCaseID int64, keys []TestKeyDto) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/testkey", testCaseID), keys, []int{http.StatusOK, http.StatusNoContent}...)
}

// ─── Scenario from run ────────────────────────────────────────────────────────

// GetTestCaseScenarioFromRun returns the scenario from the last test run for a test case.
func (c *Client) GetTestCaseScenarioFromRun(ctx context.Context, testCaseID int64) (map[string]any, error) {
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/scenariofromrun", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// ─── Automation ───────────────────────────────────────────────────────────────

// DetachTestCaseAutomation detaches automation from a test case.
func (c *Client) DetachTestCaseAutomation(ctx context.Context, testCaseID int64, statusID, workflowID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/detachautomation", testCaseID), map[string]any{
		"statusId":   statusID,
		"workflowId": workflowID,
	}, []int{http.StatusOK, http.StatusNoContent}...)
}

// ─── Version extras ───────────────────────────────────────────────────────────

// GetTestCaseVersionData returns the test case overview data for a specific version.
func (c *Client) GetTestCaseVersionData(ctx context.Context, versionID int64) (map[string]any, error) {
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/version/%d/data", versionID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteTestCaseVersion deletes a specific version of a test case.
func (c *Client) DeleteTestCaseVersion(ctx context.Context, versionID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/testcase/version/%d", versionID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

// ─── Bulk operations (new) ────────────────────────────────────────────────────

func (c *Client) bulkPost(ctx context.Context, path string, body interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, body, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent, http.StatusCreated}...)
}

// BulkAddTestCaseCustomFields adds custom field values to multiple test cases.
//
// Uses the v2 endpoint (/api/v2/test-case/bulk/cfv/add): the v1 endpoint
// (/api/testcase/bulk/cfv/add), despite matching its documented request schema,
// returns 500 on this API's backend regardless of which custom field is targeted.
// v2 uses a flat per-value shape instead of v1's field-with-nested-values shape,
// so each (field, value) pair is expanded into its own row.
func (c *Client) BulkAddTestCaseCustomFields(ctx context.Context, projectID int64, testCaseIDs []int64, cfv []CustomFieldWithValuesDto) error {
	rows := make([]CustomFieldValueWithCfV2Dto, 0, len(cfv))
	for _, f := range cfv {
		for _, v := range f.Values {
			rows = append(rows, CustomFieldValueWithCfV2Dto{ID: v.ID, CustomField: f.CustomField})
		}
	}
	return c.bulkPost(ctx, "/api/v2/test-case/bulk/cfv/add", TestCaseCfvBulkAddDto{
		Selection: TestCaseSelectionDtoV2{ProjectID: projectID, TestCasesInclude: testCaseIDs},
		Cfv:       rows,
	})
}

// BulkRemoveTestCaseCustomFields removes custom field values from multiple test cases.
func (c *Client) BulkRemoveTestCaseCustomFields(ctx context.Context, projectID int64, testCaseIDs []int64, cfIDs []int64) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/cfv/remove", BulkCfvRemoveDto{
		Selection: TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
		IDs:       cfIDs,
	})
}

// BulkAddTestCaseExternalLinks adds external links to multiple test cases.
func (c *Client) BulkAddTestCaseExternalLinks(ctx context.Context, projectID int64, testCaseIDs []int64, links []ExternalLinkDto) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/externallink/add", BulkExternalLinkAddDto{
		Selection: TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
		Links:     links,
	})
}

// BulkAddTestCaseIssues adds issues to multiple test cases.
func (c *Client) BulkAddTestCaseIssues(ctx context.Context, projectID int64, testCaseIDs []int64, issues []IssueDto) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/issue/add", BulkIssueAddDto{
		Selection: TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
		Issues:    issues,
	})
}

// BulkRemoveTestCaseIssues removes issues from multiple test cases.
func (c *Client) BulkRemoveTestCaseIssues(ctx context.Context, projectID int64, testCaseIDs []int64, issueIDs []int64) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/issue/remove", BulkIssueRemoveDto{
		Selection: TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
		IDs:       issueIDs,
	})
}

// BulkSetTestCaseLayer sets the test layer for multiple test cases.
func (c *Client) BulkSetTestCaseLayer(ctx context.Context, projectID int64, testCaseIDs []int64, layerID int64) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/layer/set", BulkLayerSetDto{
		Selection: TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
		LayerID:   layerID,
	})
}

// BulkMoveTestCases moves multiple test cases to another project.
func (c *Client) BulkMoveTestCases(ctx context.Context, projectID int64, testCaseIDs []int64, toProjectID int64) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/move", BulkMoveDto{
		Selection:   TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
		ToProjectID: toProjectID,
	})
}

// BulkDeleteTestCases permanently deletes multiple test cases.
func (c *Client) BulkDeleteTestCases(ctx context.Context, projectID int64, testCaseIDs []int64) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/remove", BulkDeleteDto{
		Selection: TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
	})
}

// BulkRunTestCasesNewLaunch runs multiple test cases in a new launch.
func (c *Client) BulkRunTestCasesNewLaunch(ctx context.Context, projectID int64, testCaseIDs []int64, launchName string, assignees []string) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/run/new", BulkRunNewLaunchDto{
		Selection:  TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
		LaunchName: launchName,
		Assignees:  assignees,
	})
}

// BulkRunTestCasesExistingLaunch runs multiple test cases in an existing launch.
func (c *Client) BulkRunTestCasesExistingLaunch(ctx context.Context, projectID int64, testCaseIDs []int64, launchID int64, assignees []string) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/run/existing", BulkRunExistingLaunchDto{
		Selection: TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
		LaunchID:  launchID,
		Assignees: assignees,
	})
}

// BulkCreateTestPlan creates a test plan from multiple test cases.
func (c *Client) BulkCreateTestPlan(ctx context.Context, projectID int64, testCaseIDs []int64, testPlanName string) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/testplan/create", BulkCreateTestPlanDto{
		Selection:    TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
		TestPlanName: testPlanName,
	})
}

// BulkMuteTestCases mutes multiple test cases.
func (c *Client) BulkMuteTestCases(ctx context.Context, projectID int64, testCaseIDs []int64) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/mute/add", BulkMuteDto{
		Selection: TestCaseTreeSelectionDto{ProjectID: projectID, LeafsInclude: testCaseIDs},
	})
}

// ─────────────────────────────────────────────────────────────────────────────

func (c *Client) DeleteTestCase(ctx context.Context, testCaseID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/testcase/%d", testCaseID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) CreateTestCaseStep(ctx context.Context, req ScenarioStepCreateRequest, afterID int64) (int64, error) {
	url := "/api/testcase/step"
	if afterID > 0 {
		url += fmt.Sprintf("?afterId=%d", afterID)
	}

	resp, err := c.doRaw(ctx, http.MethodPost, url, req, []int{http.StatusOK}...)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result ScenarioStepCreatedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return result.CreatedStepID, nil
}

// UpdateTestCaseStep patches a scenario step.
//
// withExpectedResult controls the API's expected-result bookkeeping for this
// call and must be chosen carefully — it is not simply "true when setting
// expectedResult":
//   - If false on an edit to a step that already has an expected result, the
//     API silently detaches expectedResultId and deletes the expected-result
//     child node, wiping it.
//   - If true on an edit to a step that has NO expected result yet — even a
//     body-only edit that never mentions ExpectedResult — the API spawns a new,
//     empty "Expected Result" placeholder child node that did not previously
//     exist and is not cleaned up (confirmed live: github.com/MimoJanra/TestOpsMCP/issues/16).
//
// So: pass true when this call sets ExpectedResult, or when the step is known
// (via a prior read) to already have an expectedResultId to preserve; pass
// false otherwise. Callers should look up the step's current state rather than
// defaulting to true "to be safe".
func (c *Client) UpdateTestCaseStep(ctx context.Context, stepID int64, req ScenarioStepPatchRequest, withExpectedResult bool) error {
	url := fmt.Sprintf("/api/testcase/step/%d", stepID)
	if withExpectedResult {
		url += "?withExpectedResult=true"
	}
	return c.doRequest(ctx, http.MethodPatch, url, req, []int{http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) DeleteTestCaseStep(ctx context.Context, stepID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/testcase/step/%d", stepID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) BulkSetTestCaseStatus(ctx context.Context, projectID, statusID, workflowID int64, testCaseIDs []int64) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testcase/bulk/status/set", TestCaseBulkStatusDto{
		Selection:  selection,
		StatusID:   statusID,
		WorkflowID: workflowID,
	}, []int{http.StatusNoContent, http.StatusOK}...)
}

func (c *Client) BulkAddTestCaseTags(ctx context.Context, projectID int64, testCaseIDs []int64, tags []TestTagDto) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testcase/bulk/tag/add", TestCaseBulkTagDto{
		Selection: selection,
		Tags:      tags,
	}, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}...)
}

func (c *Client) BulkRemoveTestCaseTags(ctx context.Context, projectID int64, testCaseIDs []int64, tags []TestTagDto) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testcase/bulk/tag/remove", TestCaseBulkTagDto{
		Selection: selection,
		Tags:      tags,
	}, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}...)
}

func (c *Client) BulkAddTestCaseMembers(ctx context.Context, projectID int64, testCaseIDs []int64, members []MemberDto) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testcase/bulk/member/add", TestCaseBulkMemberDto{
		Selection: selection,
		Members:   members,
	}, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}...)
}

func (c *Client) BulkRemoveTestCaseMembers(ctx context.Context, projectID int64, testCaseIDs []int64, members []MemberDto) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testcase/bulk/member/remove", TestCaseBulkMemberDto{
		Selection: selection,
		Members:   members,
	}, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}...)
}

func (c *Client) BulkAssignTestResults(ctx context.Context, launchID int64, testResultIDs []int64, assignees []string) error {
	selection := TestResultTreeSelectionDto{
		LaunchID:     launchID,
		LeafsInclude: testResultIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testresult/bulk/assign", TestResultBulkAssignDto{
		Selection: selection,
		Assignees: assignees,
	}, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}...)
}

// DeleteTestResult permanently deletes a single test result by ID, removing it
// (and the test case it represents) from the launch it belongs to.
func (c *Client) DeleteTestResult(ctx context.Context, testResultID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/testresult/%d", testResultID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

// BulkHideTestResults hides the given test results in a launch
// (POST /api/testresult/bulk/hide). Hidden results stay in the launch data but
// are excluded from the report; unlike DeleteTestResult this is non-destructive.
func (c *Client) BulkHideTestResults(ctx context.Context, launchID int64, testResultIDs []int64) error {
	return c.doRequest(ctx, http.MethodPost, "/api/testresult/bulk/hide", TestResultBulkDto{
		Selection: TestResultTreeSelectionDto{
			LaunchID:     launchID,
			LeafsInclude: testResultIDs,
		},
	}, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}...)
}

func (c *Client) BulkMuteTestResults(ctx context.Context, launchID int64, testResultIDs []int64, reason string) error {
	selection := TestResultTreeSelectionDto{
		LaunchID:     launchID,
		LeafsInclude: testResultIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testresult/bulk/mute", TestResultBulkMuteDto{
		Selection: selection,
		Reason:    reason,
	}, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}...)
}

func (c *Client) BulkUnmuteTestResults(ctx context.Context, launchID int64, testResultIDs []int64) error {
	selection := TestResultTreeSelectionDto{
		LaunchID:     launchID,
		LeafsInclude: testResultIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testresult/bulk/unmute", map[string]interface{}{
		"selection": selection,
	}, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}...)
}

func (c *Client) BulkResolveTestResults(ctx context.Context, launchID int64, testResultIDs []int64, status string) error {
	selection := TestResultTreeSelectionDto{
		LaunchID:     launchID,
		LeafsInclude: testResultIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testresult/bulk/resolve", TestResultBulkResolveDto{
		Selection: selection,
		Status:    status,
	}, []int{http.StatusNoContent, http.StatusOK, http.StatusAccepted}...)
}

func (c *Client) AddTestCasesToLaunch(ctx context.Context, launchID int64, projectID int64, testCaseIDs []int64, assignees []string) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/launch/%d/testcase/add", launchID), LaunchTestCasesAddDto{
		Selection: selection,
		Assignees: assignees,
	}, []int{http.StatusAccepted, http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) AddTestPlanToLaunch(ctx context.Context, launchID int64, testPlanID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/launch/%d/testplan/add", launchID), LaunchTestPlanAddDto{
		TestPlanID: testPlanID,
	}, []int{http.StatusAccepted, http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) url(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.baseURL + path
}

func (c *Client) setAuthHeader(ctx context.Context, req *http.Request) error {
	jwt, err := c.getJWTToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)

	if c.httpClient.Jar != nil {
		cookies := c.httpClient.Jar.Cookies(req.URL)
		for _, cookie := range cookies {
			if cookie.Name == "XSRF-TOKEN" {
				req.Header.Set("X-XSRF-TOKEN", cookie.Value)
				break
			}
		}
	}
	return nil
}

// doRaw performs the request/auth/send/status-check steps shared by every Allure
// API call and returns the response for the caller to read and close. reqBody,
// when non-nil, is marshaled as the JSON request body. okStatuses lists the
// response codes treated as success (defaults to 200 OK); any other status is
// converted to an *APIError via errFromResponse, and the response body is
// closed before returning.
func (c *Client) doRaw(ctx context.Context, method, path string, reqBody any, okStatuses ...int) (*http.Response, error) {
	var reader io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewBuffer(data)
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, c.url(path), reader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	if reqBody != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}

	if len(okStatuses) == 0 {
		okStatuses = []int{http.StatusOK}
	}
	ok := false
	for _, s := range okStatuses {
		if resp.StatusCode == s {
			ok = true
			break
		}
	}
	if !ok {
		defer resp.Body.Close()
		return nil, errFromResponse(resp)
	}
	return resp, nil
}

// doJSON performs doRaw and, when out is non-nil, decodes the JSON response body into it.
func (c *Client) doJSON(ctx context.Context, method, path string, reqBody, out any, okStatuses ...int) error {
	resp, err := c.doRaw(ctx, method, path, reqBody, okStatuses...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// doRequest is doJSON without a response body to decode.
func (c *Client) doRequest(ctx context.Context, method, path string, reqBody any, okStatuses ...int) error {
	return c.doJSON(ctx, method, path, reqBody, nil, okStatuses...)
}

func (c *Client) CloneTestCase(ctx context.Context, testCaseID int64) (int64, error) {
	resp, err := c.doRaw(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/clone", testCaseID), nil, []int{http.StatusOK, http.StatusCreated}...)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return result.ID, nil
}

// CopyLaunch starts an async copy of a launch. Per the API spec, this endpoint
// responds 202 Accepted with no body — the new launch's ID is not returned
// here; find it afterwards via ListLaunches.
func (c *Client) CopyLaunch(ctx context.Context, launchID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/launch/%d/copy", launchID), map[string]any{}, []int{http.StatusOK, http.StatusAccepted, http.StatusCreated, http.StatusNoContent}...)
}

func (c *Client) ResolveTestResult(ctx context.Context, testResultID int64, status string) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testresult/%d/resolve", testResultID), map[string]any{"status": status}, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}...)
}

func (c *Client) UnmuteTestResult(ctx context.Context, testResultID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testresult/%d/unmute", testResultID), nil, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}...)
}

func (c *Client) GetLaunchEnvironment(ctx context.Context, launchID int64) (map[string]any, error) {
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/launch/%d/env", launchID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetTestCaseHistory(ctx context.Context, testCaseID int64, page, size int) (map[string]any, error) {
	url := fmt.Sprintf("/api/testcase/%d/history?page=%d&size=%d", testCaseID, page, size)
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetLaunchDefects(ctx context.Context, launchID int64, page, size int) (map[string]any, error) {
	url := fmt.Sprintf("/api/launch/%d/defect?page=%d&size=%d", launchID, page, size)
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetTestCaseDefects(ctx context.Context, testCaseID int64, page, size int) (map[string]any, error) {
	url := fmt.Sprintf("/api/testcase/%d/defect?page=%d&size=%d", testCaseID, page, size)
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, url, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) MergeLaunches(ctx context.Context, launchIDs []int64, launchName string) (int64, error) {
	resp, err := c.doRaw(ctx, http.MethodPost, "/api/launch/merge", map[string]any{
		"launchIds": launchIDs,
		"name":      launchName,
	}, []int{http.StatusOK, http.StatusCreated}...)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return result.ID, nil
}

func (c *Client) AddTestCaseDefect(ctx context.Context, testCaseID, defectID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/defect/%d", testCaseID, defectID), nil, []int{http.StatusOK, http.StatusCreated, http.StatusNoContent}...)
}

func (c *Client) RemoveTestCaseDefect(ctx context.Context, testCaseID, defectID int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/testcase/%d/defect/%d", testCaseID, defectID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) GetTestCaseMembers(ctx context.Context, testCaseID int64) ([]MemberDto, error) {
	var result []MemberDto
	if err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/members", testCaseID), nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) AddTestCaseMembers(ctx context.Context, testCaseID int64, members []MemberDto) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/members", testCaseID), members, []int{http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) RemoveTestCaseMembers(ctx context.Context, projectID, testCaseID int64, memberIDs []int64) error {
	return c.bulkPost(ctx, "/api/testcase/bulk/member/remove", map[string]any{
		"ids": memberIDs,
		"selection": TestCaseTreeSelectionDto{
			ProjectID:    projectID,
			LeafsInclude: []int64{testCaseID},
		},
	})
}

// GetTestCaseExternalLinks returns the external URL links attached to a test case.
// The spec exposes these only via GET /api/testcase/{id}/overview (field "links").
// The /relation endpoint returns test-case-to-test-case relations (different schema).
func (c *Client) GetTestCaseExternalLinks(ctx context.Context, testCaseID int64) ([]ExternalLinkDto, error) {
	resp, err := c.doRaw(ctx, http.MethodGet, fmt.Sprintf("/api/testcase/%d/overview", testCaseID), nil, []int{http.StatusOK}...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode only the "links" field to avoid pulling the full overview into memory.
	var overview struct {
		Links []ExternalLinkDto `json:"links"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return overview.Links, nil
}

// AddTestCaseExternalLink appends an external URL link to a test case.
// The spec provides no per-item POST endpoint; the only way to mutate links is
// PATCH /api/testcase/{id} with the complete desired "links" array (replace-all).
// We therefore fetch the current links first and patch with the appended list.
func (c *Client) AddTestCaseExternalLink(ctx context.Context, testCaseID int64, link ExternalLinkDto) error {
	current, err := c.GetTestCaseExternalLinks(ctx, testCaseID)
	if err != nil {
		return fmt.Errorf("get current links: %w", err)
	}
	updated := append(current, link)
	return c.UpdateTestCase(ctx, testCaseID, UpdateTestCaseRequest{Links: updated})
}

// DeleteTestCaseExternalLink removes the external link with the given URL from a test case.
// The spec has no DELETE endpoint for external links; removal is achieved by fetching
// the current list and PATCHing with the matching entry omitted.
func (c *Client) DeleteTestCaseExternalLink(ctx context.Context, testCaseID int64, linkURL string) error {
	current, err := c.GetTestCaseExternalLinks(ctx, testCaseID)
	if err != nil {
		return fmt.Errorf("get current links: %w", err)
	}

	remaining := make([]ExternalLinkDto, 0, len(current))
	found := false
	for _, l := range current {
		if l.URL == linkURL {
			found = true
			continue
		}
		remaining = append(remaining, l)
	}
	if !found {
		return fmt.Errorf("external link not found: %s", linkURL)
	}

	return c.UpdateTestCase(ctx, testCaseID, UpdateTestCaseRequest{Links: remaining})
}

func (c *Client) RestoreTestCase(ctx context.Context, testCaseID int64) error {
	return c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/testcase/%d/restore", testCaseID), nil, []int{http.StatusOK, http.StatusNoContent}...)
}

func (c *Client) BulkCloneTestCases(ctx context.Context, projectID int64, testCaseIDs []int64) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	return c.doRequest(ctx, http.MethodPost, "/api/testcase/bulk/clone", map[string]any{
		"selection": selection,
	}, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}...)
}

// Public methods for OpenAPI execution

// GetBaseURL returns the base URL of the TestOps API
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

// SetAuthHeader sets the authorization header for a request
func (c *Client) SetAuthHeader(ctx context.Context, req *http.Request) error {
	return c.setAuthHeader(ctx, req)
}

// GetHTTPClient returns the underlying HTTP client
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

// HasToken checks if the client has a configured token
func (c *Client) HasToken() bool {
	return c.userToken != ""
}

// ---------------------------------------------------------------------------
// Test Case Tree
// ---------------------------------------------------------------------------

// BrowseTestCaseTree returns folders (groups) and test cases (leaves) at the
// given tree path within a project. Pass an empty path to start at the root.
func (c *Client) BrowseTestCaseTree(ctx context.Context, projectID int64, path []int64, page, size int) (map[string]any, error) {
	q := fmt.Sprintf("/api/testcasetree/leaf?projectId=%d&page=%d&size=%d", projectID, page, size)
	for _, p := range path {
		q += fmt.Sprintf("&path=%d", p)
	}
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, q, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTestCaseTreeGroups returns the sub-folders (groups) at the given tree path.
func (c *Client) GetTestCaseTreeGroups(ctx context.Context, projectID int64, path []int64, page, size int) (map[string]any, error) {
	q := fmt.Sprintf("/api/testcasetree/group?projectId=%d&page=%d&size=%d", projectID, page, size)
	for _, p := range path {
		q += fmt.Sprintf("&path=%d", p)
	}
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodGet, q, nil, &result, []int{http.StatusOK}...); err != nil {
		return nil, err
	}
	return result, nil
}

// MoveTestCasesToFolder moves the given test cases to the destination tree path
// (drag-and-drop reordering / folder assignment in the tree).
func (c *Client) MoveTestCasesToFolder(ctx context.Context, projectID int64, testCaseIDs []int64, destPath []int64) error {
	return c.doRequest(ctx, http.MethodPost, "/api/testcase/bulk/draganddrop", map[string]any{
		"path": destPath,
		"selection": map[string]any{
			"projectId":    projectID,
			"leafsInclude": testCaseIDs,
		},
	}, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}...)
}

// CreateTestCaseFolder creates a new folder (group) at the given tree path.
func (c *Client) CreateTestCaseFolder(ctx context.Context, projectID int64, parentPath []int64, name string) (map[string]any, error) {
	q := fmt.Sprintf("/api/testcasetree/group?projectId=%d", projectID)
	for _, p := range parentPath {
		q += fmt.Sprintf("&path=%d", p)
	}
	var result map[string]any
	if err := c.doJSON(ctx, http.MethodPost, q, map[string]any{"name": name}, &result, []int{http.StatusOK, http.StatusCreated}...); err != nil {
		return nil, err
	}
	return result, nil
}

// APIError represents a failed Allure API response. Code holds a machine-readable
// error identifier parsed from the response body (e.g. "no-job-assigned"), when
// the upstream service returns one, so callers can branch on it with errors.As
// instead of matching substrings in the formatted error text.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("unexpected status %d", e.StatusCode)
	}
	return fmt.Sprintf("unexpected status %d: %s", e.StatusCode, e.Message)
}

func errFromResponse(resp *http.Response) error {
	const limit = 4 * 1024
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return fmt.Errorf("unexpected status %d: read body: %w", resp.StatusCode, err)
	}
	text := strings.TrimSpace(string(body))

	apiErr := &APIError{StatusCode: resp.StatusCode, Message: text}

	// Allure error bodies are typically JSON with a machine-readable code under
	// one of these keys; try each and keep the raw text as Message regardless.
	var parsed struct {
		Code      string `json:"code"`
		ErrorCode string `json:"errorCode"`
		ErrorType string `json:"errorType"`
	}
	if text != "" && json.Unmarshal(body, &parsed) == nil {
		switch {
		case parsed.Code != "":
			apiErr.Code = parsed.Code
		case parsed.ErrorCode != "":
			apiErr.Code = parsed.ErrorCode
		case parsed.ErrorType != "":
			apiErr.Code = parsed.ErrorType
		}
	}

	return apiErr
}
