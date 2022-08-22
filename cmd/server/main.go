package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/horseinthesky/metricsagent/internal/server"
)

const (
	listenOn = "localhost:8080"
)

var (
	address       *string
	restore       *bool
	storeInterval *time.Duration
	storeFile     *string
	cfg           = &server.Config{}
)

func overrideConfig(cfg *server.Config) {
	if _, present := os.LookupEnv("ADDRESS"); !present {
		cfg.Address = *address
	}
	if _, present := os.LookupEnv("STORE_INTERVAL"); !present {
		cfg.StoreInterval = *storeInterval
	}
	if _, present := os.LookupEnv("STORE_FILE"); !present {
		cfg.StoreFile = *storeFile
	}
	if _, present := os.LookupEnv("RESTORE"); !present {
		cfg.Restore = *restore
	}
}

func init() {
	// Parse env vars
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse env vars: %w", err))
	}

	// Parse flags
	address = flag.String("a", "localhost:8080", "Socket to listen on")
	restore = flag.Bool("r", true, "If should restore metrics on startup")
	storeInterval = flag.Duration("i", time.Duration(300*time.Second), "backup interval (seconds)")
	storeFile = flag.String("f", "/tmp/devops-metrics-db.json", "Metrics backup file path")
	flag.Parse()

	overrideConfig(cfg)
}

func main() {
	metricsServer := server.New(cfg)
	metricsServer.Start()
}
