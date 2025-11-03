package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rmrfslashbin/openplantbook-mcp/internal/server"
)

var (
	// Injected at build time via ldflags
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

func main() {
	// Parse flags
	configPath := flag.String("config", "", "Path to config file (default: ~/.config/openplantbook-mcp/config.json)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("openplantbook-mcp %s\n", version)
		fmt.Printf("  commit: %s\n", gitCommit)
		fmt.Printf("  built:  %s\n", buildTime)
		os.Exit(0)
	}

	// Load configuration
	config, err := server.LoadConfig(*configPath)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nProvide credentials via environment variables:\n")
		fmt.Fprintf(os.Stderr, "  OPENPLANTBOOK_API_KEY=xxx  (for API key auth)\n")
		fmt.Fprintf(os.Stderr, "OR\n")
		fmt.Fprintf(os.Stderr, "  OPENPLANTBOOK_CLIENT_ID=xxx OPENPLANTBOOK_CLIENT_SECRET=xxx  (for OAuth2)\n")
		os.Exit(1)
	}

	// Create server
	srv, err := server.New(config, version)
	if err != nil {
		slog.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Run(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		slog.Info("shutdown signal received", "signal", sig)
		cancel()
	case err := <-errChan:
		if err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}

	slog.Info("shutdown complete")
}
