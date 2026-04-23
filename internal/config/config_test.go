package config

import (
	"testing"
	"time"
)

func TestLoad_RequiresBaseURL(t *testing.T) {
	t.Setenv("ALLURE_BASE_URL", "")
	t.Setenv("ALLURE_TOKEN", "tok")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for missing ALLURE_BASE_URL")
	}
}

func TestLoad_RequiresValidURL(t *testing.T) {
	t.Setenv("ALLURE_BASE_URL", "not-a-url")
	t.Setenv("ALLURE_TOKEN", "tok")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for malformed URL")
	}
}

func TestLoad_RequiresToken(t *testing.T) {
	t.Setenv("ALLURE_BASE_URL", "https://allure.example.com")
	t.Setenv("ALLURE_TOKEN", "")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for missing token")
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("ALLURE_BASE_URL", "https://allure.example.com/")
	t.Setenv("ALLURE_TOKEN", "tok")
	t.Setenv("REQUEST_TIMEOUT", "")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("CORS_ALLOWED_ORIGIN", "")
	t.Setenv("MCP_AUTH_TOKEN", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AllureBaseURL != "https://allure.example.com" {
		t.Errorf("trailing slash not trimmed: %q", cfg.AllureBaseURL)
	}
	if cfg.RequestTimeout != 30*time.Second {
		t.Errorf("default timeout = %v", cfg.RequestTimeout)
	}
	if cfg.Port != ":3000" {
		t.Errorf("default port = %q", cfg.Port)
	}
	if cfg.LogLevel != "INFO" {
		t.Errorf("default log level = %q", cfg.LogLevel)
	}
	if cfg.CORSAllowOrigin != "*" {
		t.Errorf("default CORS = %q", cfg.CORSAllowOrigin)
	}
}

func TestLoad_PortNormalization(t *testing.T) {
	t.Setenv("ALLURE_BASE_URL", "https://allure.example.com")
	t.Setenv("ALLURE_TOKEN", "tok")

	cases := map[string]string{
		"8080":  ":8080",
		":8080": ":8080",
	}
	for in, want := range cases {
		t.Setenv("PORT", in)
		cfg, err := Load()
		if err != nil {
			t.Fatalf("in=%q: %v", in, err)
		}
		if cfg.Port != want {
			t.Errorf("in=%q: got %q want %q", in, cfg.Port, want)
		}
	}
}

func TestLoad_TimeoutBounds(t *testing.T) {
	t.Setenv("ALLURE_BASE_URL", "https://allure.example.com")
	t.Setenv("ALLURE_TOKEN", "tok")

	cases := []string{"0", "-1", "abc", "601"}
	for _, v := range cases {
		t.Setenv("REQUEST_TIMEOUT", v)
		if _, err := Load(); err == nil {
			t.Errorf("REQUEST_TIMEOUT=%q: expected error", v)
		}
	}

	t.Setenv("REQUEST_TIMEOUT", "45")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RequestTimeout != 45*time.Second {
		t.Errorf("timeout = %v, want 45s", cfg.RequestTimeout)
	}
}
