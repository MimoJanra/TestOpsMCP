package allure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: timeout},
	}
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
	c.setAuthHeader(httpReq)
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

func (c *Client) GetLaunchStatus(ctx context.Context, launchID int64) (string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/rs/launch/%d", launchID)), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	c.setAuthHeader(httpReq)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errFromResponse(resp)
	}

	var result LaunchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Status, nil
}

func (c *Client) GetLaunchStatistics(ctx context.Context, launchID int64) (*StatisticsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(fmt.Sprintf("/api/rs/launch/%d/statistic", launchID)), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)
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
	c.setAuthHeader(httpReq)

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

func (c *Client) url(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.baseURL + path
}

func (c *Client) setAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
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
