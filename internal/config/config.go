package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTimeoutSec = 30
	maxTimeoutSec     = 600
	defaultPort       = ":3000"
	defaultLogLevel   = "INFO"
	defaultCORSOrigin = "*"
)

type Config struct {
	AllureBaseURL   string
	AllureToken     string
	RequestTimeout  time.Duration
	Port            string
	LogLevel        string
	AuthToken       string
	CORSAllowOrigin string
}

func Load() (*Config, error) {
	baseURL := strings.TrimSpace(os.Getenv("ALLURE_BASE_URL"))
	if baseURL == "" {
		return nil, errors.New("ALLURE_BASE_URL not set")
	}
	u, err := url.Parse(baseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("ALLURE_BASE_URL must be a valid http(s) URL, got %q", baseURL)
	}
	baseURL = strings.TrimRight(baseURL, "/")

	token := strings.TrimSpace(os.Getenv("ALLURE_TOKEN"))
	if token == "" {
		return nil, errors.New("ALLURE_TOKEN not set")
	}

	timeout, err := parseTimeout(os.Getenv("REQUEST_TIMEOUT"))
	if err != nil {
		return nil, err
	}

	port := normalizePort(os.Getenv("PORT"))

	logLevel := strings.ToUpper(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	if logLevel == "" {
		logLevel = defaultLogLevel
	}

	corsOrigin := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGIN"))
	if corsOrigin == "" {
		corsOrigin = defaultCORSOrigin
	}

	return &Config{
		AllureBaseURL:   baseURL,
		AllureToken:     token,
		RequestTimeout:  timeout,
		Port:            port,
		LogLevel:        logLevel,
		AuthToken:       strings.TrimSpace(os.Getenv("MCP_AUTH_TOKEN")),
		CORSAllowOrigin: corsOrigin,
	}, nil
}

func parseTimeout(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultTimeoutSec * time.Second, nil
	}
	sec, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("REQUEST_TIMEOUT must be an integer (seconds), got %q", raw)
	}
	if sec <= 0 || sec > maxTimeoutSec {
		return 0, fmt.Errorf("REQUEST_TIMEOUT must be in (0, %d], got %d", maxTimeoutSec, sec)
	}
	return time.Duration(sec) * time.Second, nil
}

func normalizePort(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultPort
	}
	if strings.HasPrefix(raw, ":") {
		return raw
	}
	return ":" + raw
}
