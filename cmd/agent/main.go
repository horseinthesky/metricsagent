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

	"github.com/horseinthesky/metricsagent/internal/agent"
)

const (
	defaultAddress        = "localhost:8080"
	defaultReportInterval = time.Duration(10 * time.Second)
	defaultPollInterval   = time.Duration(2 * time.Second)
)

func getConfig() *agent.Config {
	cfg := &agent.Config{}

	flag.StringVar(&cfg.Address, "a", defaultAddress, "Address for sending data to")
	flag.DurationVar(&cfg.ReportInterval, "r", defaultReportInterval, "Metric report to server interval")
	flag.DurationVar(&cfg.PollInterval, "p", defaultPollInterval, "Metric poll interval")
	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		log.Fatal(fmt.Errorf("failed to parse env vars: %w", err))
	}

	return cfg
}

func main() {
	cfg := getConfig()
	agent := agent.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go agent.Run(ctx)

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()
	time.Sleep(200 * time.Millisecond)
}
