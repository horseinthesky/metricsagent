package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Example() {
	// Start server
	cfg, err := ParseConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse server config: %w", err))
	}

	server, err := NewServer(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create server: %w", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	go server.Run(ctx)

	// Handle graceful shutdown
	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()
	server.Stop()
}
