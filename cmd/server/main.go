package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/horseinthesky/metricsagent/internal/server"
)

const (
	defaultListenOn      = "localhost:8080"
	defaultRestoreFlag   = true
	defaultStoreInterval = 300 * time.Second
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
)

var (
	address       *string
	restore       *bool
	storeInterval *time.Duration
	storeFile     *string
	cfg           = &server.Config{}
)

func overrideConfig(cfg *server.Config) {
	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		cfg.Address = *address
	}
	if _, ok := os.LookupEnv("STORE_INTERVAL"); !ok {
		cfg.StoreInterval = *storeInterval
	}
	if _, ok := os.LookupEnv("STORE_FILE"); !ok {
		cfg.StoreFile = *storeFile
	}
	if _, ok := os.LookupEnv("RESTORE"); !ok {
		cfg.Restore = *restore
	}
}

func init() {
	// Parse env vars
	if err := env.Parse(cfg); err != nil {
		log.Fatal(fmt.Errorf("failed to parse env vars: %w", err))
	}

	// Parse flags
	address = flag.String("a", defaultListenOn, "Socket to listen on")
	restore = flag.Bool("r", defaultRestoreFlag, "If should restore metrics on startup")
	storeInterval = flag.Duration("i", defaultStoreInterval, "backup interval (seconds)")
	storeFile = flag.String("f", defaultStoreFile, "Metrics backup file path")
	flag.Parse()

	overrideConfig(cfg)
}

func main() {
	// Start server
	metricsServer := server.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go metricsServer.Start(ctx)

	// Handle graceful shutdown
	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()
	time.Sleep(200 *time.Millisecond)
}
