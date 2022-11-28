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

	metricsServer := server.NewServer(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go metricsServer.Run(ctx)

	// Log build info
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	// Handle graceful shutdown
	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()
	metricsServer.Stop()
}
