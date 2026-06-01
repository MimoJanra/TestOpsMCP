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
	defaultTimeoutSec         = 30
	maxTimeoutSec             = 600
	defaultPort               = ":3000"
	defaultLogLevel           = "INFO"
	defaultCORSOrigin         = "" // disabled by default; set CORS_ALLOWED_ORIGIN explicitly in production
	defaultAuditRetentionDays = 30
	defaultAuditLogPath       = "audit"
)

// UserConfig holds a named user and their MCP bearer token.
type UserConfig struct {
	Name  string
	Token string
}

type Config struct {
	AllureBaseURL      string
	AllureToken        string
	RequestTimeout     time.Duration
	Port               string
	LogLevel           string
	Users              []UserConfig
	CORSAllowOrigin    string
	AuditLogPath       string
	AuditRetentionDays int
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

	users := parseUsers(os.Getenv("MCP_AUTH_TOKENS"), os.Getenv("MCP_AUTH_TOKEN"))

	auditLogPath := strings.TrimSpace(os.Getenv("AUDIT_LOG_PATH"))
	if auditLogPath == "" {
		auditLogPath = defaultAuditLogPath
	}

	auditRetentionDays, err := parseRetentionDays(os.Getenv("AUDIT_RETENTION_DAYS"))
	if err != nil {
		return nil, err
	}

	return &Config{
		AllureBaseURL:      baseURL,
		AllureToken:        token,
		RequestTimeout:     timeout,
		Port:               port,
		LogLevel:           logLevel,
		Users:              users,
		CORSAllowOrigin:    corsOrigin,
		AuditLogPath:       auditLogPath,
		AuditRetentionDays: auditRetentionDays,
	}, nil
}

// parseUsers builds the user list from MCP_AUTH_TOKENS (preferred) or MCP_AUTH_TOKEN (legacy).
// MCP_AUTH_TOKENS format: "alice:token1,bob:token2"
func parseUsers(tokensEnv, singleToken string) []UserConfig {
	tokensEnv = strings.TrimSpace(tokensEnv)
	if tokensEnv != "" {
		var users []UserConfig
		for _, entry := range strings.Split(tokensEnv, ",") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			parts := strings.SplitN(entry, ":", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				continue
			}
			users = append(users, UserConfig{Name: strings.TrimSpace(parts[0]), Token: strings.TrimSpace(parts[1])})
		}
		return users
	}
	if t := strings.TrimSpace(singleToken); t != "" {
		return []UserConfig{{Name: "default", Token: t}}
	}
	return nil
}

func parseRetentionDays(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultAuditRetentionDays, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("AUDIT_RETENTION_DAYS must be a positive integer, got %q", raw)
	}
	return n, nil
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
