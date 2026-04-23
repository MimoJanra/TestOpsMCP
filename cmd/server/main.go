package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
	"github.com/MimoJanra/TestOpsMCP/internal/config"
	"github.com/MimoJanra/TestOpsMCP/internal/core"
	"github.com/MimoJanra/TestOpsMCP/internal/mcp"
	"github.com/MimoJanra/TestOpsMCP/internal/tools"
)

func main() {
	cfg := config.Load()
	logger := core.NewLogger()

	if cfg.AllureBaseURL == "" {
		logger.Error("ALLURE_BASE_URL not set", nil, nil)
		os.Exit(1)
	}

	if cfg.AllureToken == "" {
		logger.Error("ALLURE_TOKEN not set", nil, nil)
		os.Exit(1)
	}

	logger.Info("Starting Allure MCP Server", map[string]interface{}{
		"base_url": cfg.AllureBaseURL,
		"timeout":  cfg.RequestTimeout.String(),
	})

	allureClient := allure.NewClient(cfg.AllureBaseURL, cfg.AllureToken, cfg.RequestTimeout)
	registry := tools.NewRegistry(allureClient, logger)
	server := mcp.NewServer(registry, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received", nil)
		cancel()
	}()

	if err := server.Start(ctx); err != nil {
		logger.Error("Server error", err, nil)
		os.Exit(1)
	}
}
