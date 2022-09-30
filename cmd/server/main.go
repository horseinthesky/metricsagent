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

func getConfig() *server.Config {
	cfg := &server.Config{}

	flag.StringVar(&cfg.Address, "a", defaultListenOn, "Socket to listen on")
	flag.BoolVar(&cfg.Restore, "r", defaultRestoreFlag, "If should restore metrics on startup")
	flag.DurationVar(&cfg.StoreInterval, "i", defaultStoreInterval, "backup interval (seconds)")
	flag.StringVar(&cfg.StoreFile, "f", defaultStoreFile, "Metrics backup file path")
	flag.StringVar(&cfg.Key, "k", "", "Hash key")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database address")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		log.Fatal(fmt.Errorf("failed to parse env vars: %w", err))
	}

	return cfg
}

func main() {
	// Start server
	cfg := getConfig()
	metricsServer := server.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go metricsServer.Run(ctx)

	// Handle graceful shutdown
	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()
	metricsServer.Stop()
}
