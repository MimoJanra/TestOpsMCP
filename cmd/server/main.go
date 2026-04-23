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

	allureClient := allure.NewClient(cfg.AllureBaseURL, cfg.AllureToken, cfg.RequestTimeout)
	registry := tools.NewRegistry(allureClient, logger)
	mcpServer := mcp.NewServer(registry, logger, mcp.Options{
		AuthToken:       cfg.AuthToken,
		CORSAllowOrigin: cfg.CORSAllowOrigin,
	})

	if *httpMode {
		runHTTP(mcpServer, cfg, logger)
	} else {
		runStdio(mcpServer, logger)
	}
}

func runHTTP(mcpServer *mcp.Server, cfg *config.Config, logger *core.Logger) {
	logger.Info("starting Allure MCP HTTP server", map[string]any{
		"base_url":  cfg.AllureBaseURL,
		"timeout":   cfg.RequestTimeout.String(),
		"port":      cfg.Port,
		"log_level": cfg.LogLevel,
		"auth":      cfg.AuthToken != "",
		"cors":      cfg.CORSAllowOrigin,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", mcpServer.HandleSSE)
	mux.HandleFunc("/messages", mcpServer.HandleMessages)

	httpServer := &http.Server{
		Addr:              cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
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
