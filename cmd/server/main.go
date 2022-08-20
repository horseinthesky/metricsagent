package main

import (
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v6"

	"github.com/horseinthesky/metricsagent/internal/server"
)

const (
	listenOn = "localhost:8080"
)

var (
	cfg     server.Config
	address string
)

func init() {
	address = os.Getenv("ADDRESS")
	if address == "" {
		address = listenOn
	}

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse env vars: %w", err))
	}
}

func main() {
	metricsServer := server.New(cfg)
	metricsServer.Start()
}
