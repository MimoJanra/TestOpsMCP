package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/audit"
	"github.com/MimoJanra/TestOpsMCP/internal/config"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
	"github.com/MimoJanra/TestOpsMCP/internal/mcp"
	"github.com/MimoJanra/TestOpsMCP/internal/tools"
)

const shutdownTimeout = 10 * time.Second

func main() {
	httpMode := flag.Bool("http", false, "run in HTTP mode (default: stdio)")
	flag.Parse()

	bootLogger := core.NewLogger(core.LevelInfo)

	cfg, err := config.Load()
	if err != nil {
		bootLogger.Error("load config", err, nil)
		os.Exit(1)
	}

	logger := core.NewLogger(core.ParseLevel(cfg.LogLevel))

	auditLog, err := audit.NewLogger(cfg.AuditLogPath, cfg.AuditRetentionDays)
	if err != nil {
		logger.Warn("audit log disabled: cannot create directory", map[string]any{
			"path":  cfg.AuditLogPath,
			"error": err.Error(),
		})
	} else {
		defer auditLog.Close()
	}

	users := make([]mcp.User, len(cfg.Users))
	for i, u := range cfg.Users {
		users[i] = mcp.User{Name: u.Name, Token: u.Token}
	}

	allureClient := allure.NewClient(cfg.AllureBaseURL, cfg.AllureToken, cfg.RequestTimeout)
	registry := tools.NewRegistry(allureClient, logger)
	mcpServer := mcp.NewServer(registry, logger, mcp.Options{
		Users:           users,
		CORSAllowOrigin: cfg.CORSAllowOrigin,
		AuditLog:        auditLog,
	})

	if *httpMode {
		runHTTP(mcpServer, cfg, logger)
	} else {
		runStdio(mcpServer, logger)
	}
}

func runHTTP(mcpServer *mcp.Server, cfg *config.Config, logger *core.Logger) {
	logger.Info("starting Allure MCP HTTP server", map[string]any{
		"base_url":        cfg.AllureBaseURL,
		"timeout":         cfg.RequestTimeout.String(),
		"port":            cfg.Port,
		"log_level":       cfg.LogLevel,
		"users":           len(cfg.Users),
		"cors":            cfg.CORSAllowOrigin,
		"audit_path":      cfg.AuditLogPath,
		"audit_retention": cfg.AuditRetentionDays,
	})
	if len(cfg.Users) == 0 {
		logger.Warn("no auth tokens configured (MCP_AUTH_TOKENS / MCP_AUTH_TOKEN) — server accepts unauthenticated requests", nil)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", mcpServer.HandleHealth)
	// Streamable HTTP transport (MCP spec 2025-03-26) — recommended
	mux.HandleFunc("/mcp", mcpServer.HandleMCP)
	// Legacy HTTP+SSE transport (MCP spec 2024-11-05) — kept for backward compat
	mux.HandleFunc("/sse", mcpServer.HandleSSE)
	mux.HandleFunc("/messages", mcpServer.HandleMessages)

	httpServer := &http.Server{
		Addr:              cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      0, // SSE streams are long-lived; rely on client disconnect
		IdleTimeout:       120 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	logger.Info("server listening", map[string]any{"addr": cfg.Port})

	select {
	case err := <-serverErr:
		if err != nil {
			logger.Error("server error", err, nil)
			os.Exit(1)
		}
	case sig := <-sigChan:
		logger.Info("shutdown signal received", map[string]any{"signal": sig.String()})
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown", err, nil)
			os.Exit(1)
		}
		logger.Info("server stopped", nil)
	}
}

func runStdio(mcpServer *mcp.Server, logger *core.Logger) {
	logger.Info("starting Allure MCP stdio server", nil)

	handler := mcp.NewStdioHandler(mcpServer, logger)
	if err := handler.Run(); err != nil {
		logger.Error("stdio handler error", err, nil)
		os.Exit(1)
	}
}
