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

func TestLoad_TokenOptional(t *testing.T) {
	t.Setenv("ALLURE_BASE_URL", "https://allure.example.com")
	t.Setenv("ALLURE_TOKEN", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.AllureToken != "" {
		t.Errorf("expected empty token, got %q", cfg.AllureToken)
	}
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("ALLURE_BASE_URL", "https://allure.example.com/")
	t.Setenv("ALLURE_TOKEN", "")
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
	if cfg.CORSAllowOrigin != "" {
		t.Errorf("default CORS = %q", cfg.CORSAllowOrigin)
	}
}

func TestLoad_PortNormalization(t *testing.T) {
	t.Setenv("ALLURE_BASE_URL", "https://allure.example.com")
	t.Setenv("ALLURE_TOKEN", "")

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
	t.Setenv("ALLURE_TOKEN", "")

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

func TestParseUsers(t *testing.T) {
	cases := []struct {
		name        string
		tokensEnv   string
		singleToken string
		want        []UserConfig
	}{
		{"empty", "", "", nil},
		{"single_legacy_token", "", "  legacy-tok  ", []UserConfig{{Name: "default", Token: "legacy-tok"}}},
		{"multi", "alice:token1,bob:token2", "", []UserConfig{
			{Name: "alice", Token: "token1"},
			{Name: "bob", Token: "token2"},
		}},
		{"multi_takes_precedence_over_single", "alice:token1", "ignored", []UserConfig{
			{Name: "alice", Token: "token1"},
		}},
		{"skips_blank_entries", "alice:token1,,bob:token2,", "", []UserConfig{
			{Name: "alice", Token: "token1"},
			{Name: "bob", Token: "token2"},
		}},
		{"skips_malformed_entries", "alice:token1,noseparator,bob:,:empty-name", "", []UserConfig{
			{Name: "alice", Token: "token1"},
		}},
		{"trims_whitespace", " alice : token1 , bob:token2", "", []UserConfig{
			{Name: "alice", Token: "token1"},
			{Name: "bob", Token: "token2"},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := parseUsers(tc.tokensEnv, tc.singleToken)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d users %+v, want %d %+v", len(got), got, len(tc.want), tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("user[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParseRetentionDays(t *testing.T) {
	if n, err := parseRetentionDays(""); err != nil || n != defaultAuditRetentionDays {
		t.Errorf("empty: got (%d, %v), want (%d, nil)", n, err, defaultAuditRetentionDays)
	}
	if n, err := parseRetentionDays("14"); err != nil || n != 14 {
		t.Errorf("14: got (%d, %v), want (14, nil)", n, err)
	}
	for _, bad := range []string{"0", "-5", "abc", "3.5"} {
		if _, err := parseRetentionDays(bad); err == nil {
			t.Errorf("AUDIT_RETENTION_DAYS=%q: expected error", bad)
		}
	}
}
