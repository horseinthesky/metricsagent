package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Example() {
	// Start agent
	cfg, err := ParseConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse agent config: %w", err))
	}

	agent, err := NewAgent(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create agent: %w", err))
	}

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
