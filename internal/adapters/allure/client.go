package allure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	timeout    time.Duration
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{},
		timeout:    timeout,
	}
}

func (c *Client) CreateLaunch(ctx context.Context, projectID int64, launchName string) (*LaunchResponse, error) {
	req := LaunchCreateRequest{
		Name:      launchName,
		ProjectID: projectID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/rs/launch", c.baseURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setAuthHeader(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyText))
	}

	var result LaunchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetLaunchStatus(ctx context.Context, launchID int64) (string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/rs/launch/%d", c.baseURL, launchID), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyText))
	}

	var result LaunchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Status, nil
}

func (c *Client) GetLaunchStatistics(ctx context.Context, launchID int64) (*StatisticsResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/rs/launch/%d/statistic", c.baseURL, launchID), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyText, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyText))
	}

	var result StatisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) setAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
}
