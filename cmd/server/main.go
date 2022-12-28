package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/horseinthesky/metricsagent/internal/server"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	// Start server
	cfg, err := server.ParseConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse server config: %w", err))
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Log build info
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	// Handle graceful shutdown
	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	if cfg.GRPC {
		gRPCMetricsServer, err := server.NewGRPCServer(cfg)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create server: %w", err))
		}

		go gRPCMetricsServer.Run(ctx)

		sig := <-term
		log.Printf("signal received: %v; terminating...\n", sig)

		cancel()
		gRPCMetricsServer.Stop()
	} else {
		httpMetricsServer, err := server.NewServer(cfg)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create server: %w", err))
		}

		go httpMetricsServer.Run(ctx)

		sig := <-term
		log.Printf("signal received: %v; terminating...\n", sig)

		cancel()
		httpMetricsServer.Stop()
	}
}
