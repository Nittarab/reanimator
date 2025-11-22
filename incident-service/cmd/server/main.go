package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/ai-sre-platform/incident-service/internal/api"
	"github.com/your-org/ai-sre-platform/incident-service/internal/config"
	"github.com/your-org/ai-sre-platform/incident-service/internal/database"
	"github.com/your-org/ai-sre-platform/incident-service/internal/github"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := database.Connect(cfg.Database.DatabaseDSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Connect to Redis
	redis, err := database.ConnectRedis(cfg.Redis.RedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to redis: %v\n", err)
		os.Exit(1)
	}
	defer redis.Close()

	// Create GitHub client
	githubClient := github.NewClient(
		cfg.GitHub.APIURL,
		cfg.GitHub.Token,
		cfg.GitHub.WorkflowName,
		cfg.Concurrency.MaxWorkflowsPerRepo,
	)

	// Create server
	server := api.NewServer(cfg, db, redis, githubClient)
	logger := server.Logger()

	// Log startup
	logger.Info("starting incident service", map[string]interface{}{
		"port":    cfg.Server.Port,
		"version": "0.1.0",
	})

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      server.Router(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("server listening", map[string]interface{}{
			"addr": httpServer.Addr,
		})
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server", nil)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("server stopped", nil)
}
