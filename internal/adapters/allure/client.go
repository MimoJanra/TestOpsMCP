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
	"time"
)

type Client struct {
	baseURL      string
	userToken    string
	jwtToken     string
	jwtExpiresAt time.Time
	httpClient   *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		userToken: token,
		httpClient: &http.Client{
			Timeout: timeout,
			Jar:     jar,
		},
	}
}

func (c *Client) getJWTToken(ctx context.Context) (string, error) {
	if c.jwtToken != "" && time.Now().Before(c.jwtExpiresAt) {
		return c.jwtToken, nil
	}

	values := url.Values{}
	values.Set("grant_type", "apitoken")
	values.Set("scope", "openid")
	values.Set("token", c.userToken)

	body := strings.NewReader(values.Encode())
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/uaa/oauth/token"), body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Expect", "")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errFromResponse(resp)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	c.jwtToken = result.AccessToken
	if result.ExpiresIn > 0 {
		c.jwtExpiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	} else {
		c.jwtExpiresAt = time.Now().Add(1 * time.Hour)
	}
	return c.jwtToken, nil
}

func (c *Client) CreateLaunch(ctx context.Context, projectID int64, launchName string) (*LaunchResponse, error) {
	body, err := json.Marshal(LaunchCreateRequest{
		Name:      launchName,
		ProjectID: projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/rs/launch"), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errFromResponse(resp)
	}

	var result LaunchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetLaunchStatus(ctx context.Context, launchID int64) (interface{}, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/rs/launch/%d", launchID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result LaunchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Status, nil
}

func (c *Client) GetLaunchStatistics(ctx context.Context, launchID int64) (*StatisticsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/rs/launch/%d/statistic", launchID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result StatisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) CloseLaunch(ctx context.Context, launchID int64) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(fmt.Sprintf("/api/launch/%d/close", launchID)), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errFromResponse(resp)
	}

	return nil
}

func (c *Client) ReopenLaunch(ctx context.Context, launchID int64) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(fmt.Sprintf("/api/launch/%d/reopen", launchID)), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errFromResponse(resp)
	}

	return nil
}

func (c *Client) ListLaunches(ctx context.Context, projectID int64, page, size int) (*LaunchListResponse, error) {
	url := fmt.Sprintf("/api/launch?projectId=%d&page=%d&size=%d", projectID, page, size)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(url), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result LaunchListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetLaunchDetails(ctx context.Context, launchID int64) (*LaunchDetailsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/launch/%d", launchID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result LaunchDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) ListTestResults(ctx context.Context, launchID int64, status string, page, size int) (*TestResultListResponse, error) {
	url := fmt.Sprintf("/api/testresult?launchId=%d&page=%d&size=%d", launchID, page, size)
	if status != "" {
		url += fmt.Sprintf("&status=%s", status)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(url), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result TestResultListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetTestResult(ctx context.Context, testResultID int64) (*TestResultDetailsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/testresult/%d", testResultID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result TestResultDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) AssignTestResult(ctx context.Context, testResultID int64, username string) error {
	body, err := json.Marshal(AssignTestResultRequest{Username: username})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(fmt.Sprintf("/api/testresult/%d/assign", testResultID)), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errFromResponse(resp)
	}

	return nil
}

func (c *Client) MuteTestResult(ctx context.Context, testResultID int64, reason string) error {
	body, err := json.Marshal(MuteTestResultRequest{Reason: reason})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(fmt.Sprintf("/api/testresult/%d/mute", testResultID)), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errFromResponse(resp)
	}

	return nil
}

func (c *Client) ListTestCases(ctx context.Context, projectID int64, page, size int) (*TestCaseListResponse, error) {
	url := fmt.Sprintf("/api/testcase?projectId=%d&page=%d&size=%d", projectID, page, size)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(url), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result TestCaseListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetTestCase(ctx context.Context, testCaseID int64) (*TestCaseDetailsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/testcase/%d", testCaseID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result TestCaseDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) RunTestCase(ctx context.Context, testCaseID, launchID int64) error {
	body, err := json.Marshal(RunTestCaseRequest{
		TestCaseIds: []int64{testCaseID},
		LaunchId:    launchID,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testcase/bulk/run/existing"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errFromResponse(resp)
	}

	return nil
}

func (c *Client) ListProjects(ctx context.Context, page, size int) (*ProjectListResponse, error) {
	url := fmt.Sprintf("/api/project?page=%d&size=%d", page, size)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(url), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result ProjectListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetProject(ctx context.Context, projectID int64) (*ProjectDetailsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/project/%d", projectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result ProjectDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetProjectStats(ctx context.Context, projectID int64) (*ProjectStatsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/project/%d/stats", projectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result ProjectStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetLaunchTrendAnalytics(ctx context.Context, projectID int64) ([]TrendData, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/analytic/%d/statistic_trend", projectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result []TrendData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

func (c *Client) GetLaunchDurationAnalytics(ctx context.Context, projectID int64) (interface{}, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/analytic/%d/launch_duration_histogram", projectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

func (c *Client) GetTestSuccessRateAnalytics(ctx context.Context, projectID int64) (interface{}, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/analytic/%d/tc_success_rate", projectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errFromResponse(resp)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

func (c *Client) CreateTestCase(ctx context.Context, projectID int64, name, description string) (*TestCaseDetailsResponse, error) {
	body, err := json.Marshal(CreateTestCaseRequest{
		Name:        name,
		ProjectID:   projectID,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testcase"), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return nil, fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errFromResponse(resp)
	}

	var result TestCaseDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) UpdateTestCase(ctx context.Context, testCaseID int64, name, description string) error {
	req := UpdateTestCaseRequest{}
	if name != "" {
		req.Name = name
	}
	if description != "" {
		req.Description = description
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.url(fmt.Sprintf("/api/testcase/%d", testCaseID)), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errFromResponse(resp)
	}

	return nil
}

func (c *Client) DeleteTestCase(ctx context.Context, testCaseID int64) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.url(fmt.Sprintf("/api/testcase/%d", testCaseID)), nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errFromResponse(resp)
	}

	return nil
}

func (c *Client) BulkSetTestCaseStatus(ctx context.Context, projectID, statusID, workflowID int64, testCaseIDs []int64) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	body, err := json.Marshal(TestCaseBulkStatusDto{
		Selection:  selection,
		StatusID:   statusID,
		WorkflowID: workflowID,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testcase/bulk/status/set"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) BulkAddTestCaseTags(ctx context.Context, projectID int64, testCaseIDs []int64, tags []TestTagDto) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	body, err := json.Marshal(TestCaseBulkTagDto{
		Selection: selection,
		Tags:      tags,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testcase/bulk/tag/add"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) BulkRemoveTestCaseTags(ctx context.Context, projectID int64, testCaseIDs []int64, tags []TestTagDto) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	body, err := json.Marshal(TestCaseBulkTagDto{
		Selection: selection,
		Tags:      tags,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testcase/bulk/tag/remove"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) BulkAddTestCaseMembers(ctx context.Context, projectID int64, testCaseIDs []int64, members []MemberDto) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	body, err := json.Marshal(TestCaseBulkMemberDto{
		Selection: selection,
		Members:   members,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testcase/bulk/member/add"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) BulkRemoveTestCaseMembers(ctx context.Context, projectID int64, testCaseIDs []int64, members []MemberDto) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	body, err := json.Marshal(TestCaseBulkMemberDto{
		Selection: selection,
		Members:   members,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testcase/bulk/member/remove"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) BulkAssignTestResults(ctx context.Context, launchID int64, testResultIDs []int64, assignees []string) error {
	selection := TestResultTreeSelectionDto{
		LaunchID:     launchID,
		LeafsInclude: testResultIDs,
	}
	body, err := json.Marshal(TestResultBulkAssignDto{
		Selection: selection,
		Assignees: assignees,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testresult/bulk/assign"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) BulkMuteTestResults(ctx context.Context, launchID int64, testResultIDs []int64, reason string) error {
	selection := TestResultTreeSelectionDto{
		LaunchID:     launchID,
		LeafsInclude: testResultIDs,
	}
	body, err := json.Marshal(TestResultBulkMuteDto{
		Selection: selection,
		Reason:    reason,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testresult/bulk/mute"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) BulkUnmuteTestResults(ctx context.Context, launchID int64, testResultIDs []int64) error {
	selection := TestResultTreeSelectionDto{
		LaunchID:     launchID,
		LeafsInclude: testResultIDs,
	}
	body, err := json.Marshal(map[string]interface{}{
		"selection": selection,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testresult/bulk/unmute"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) BulkResolveTestResults(ctx context.Context, launchID int64, testResultIDs []int64) error {
	selection := TestResultTreeSelectionDto{
		LaunchID:     launchID,
		LeafsInclude: testResultIDs,
	}
	body, err := json.Marshal(TestResultBulkResolveDto{
		Selection: selection,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url("/api/testresult/bulk/resolve"), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) AddTestCasesToLaunch(ctx context.Context, launchID int64, projectID int64, testCaseIDs []int64, assignees []string) error {
	selection := TestCaseTreeSelectionDto{
		ProjectID:    projectID,
		LeafsInclude: testCaseIDs,
	}
	body, err := json.Marshal(LaunchTestCasesAddDto{
		Selection: selection,
		Assignees: assignees,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(fmt.Sprintf("/api/launch/%d/testcase/add", launchID)), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errFromResponse(resp)
	}
	return nil
}

func (c *Client) AddTestPlanToLaunch(ctx context.Context, launchID int64, testPlanID int64) error {
	body, err := json.Marshal(LaunchTestPlanAddDto{
		TestPlanID: testPlanID,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(fmt.Sprintf("/api/launch/%d/testplan/add", launchID)), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if err := c.setAuthHeader(ctx, httpReq); err != nil {
		return fmt.Errorf("set auth: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errFromResponse(resp)
	}
	return nil
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

func errFromResponse(resp *http.Response) error {
	const limit = 4 * 1024
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return fmt.Errorf("unexpected status %d: read body: %w", resp.StatusCode, err)
	}
	text := strings.TrimSpace(string(body))
	if text == "" {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, text)
}
