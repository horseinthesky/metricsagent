package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/horseinthesky/metricsagent/internal/server"
)

func Example() {
	cfg, err := server.ParseConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse server config: %w", err))
	}

	httpServer, err := NewServer(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create server: %w", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	go httpServer.Run(ctx)

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()
	httpServer.Stop()
}
