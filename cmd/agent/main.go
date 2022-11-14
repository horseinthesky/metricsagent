package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/horseinthesky/metricsagent/internal/agent"
)

func main() {
	// Start agent
	cfg, err := agent.ParseConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse agent config: %w", err))
	}

	agent := agent.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go agent.Run(ctx)

	// Handle graceful shutdown
	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()
	agent.Stop()
}
