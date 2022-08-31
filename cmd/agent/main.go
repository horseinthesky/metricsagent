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

var (
	address        *string
	reportInterval *time.Duration
	pollInterval   *time.Duration
	cfg            = &agent.Config{}
)

func overrideConfig(cfg *agent.Config) {
	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		cfg.Address = *address
	}
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); !ok {
		cfg.ReportInterval = *reportInterval
	}
	if _, ok := os.LookupEnv("POLL_INTERVAL"); !ok {
		cfg.PollInterval = *pollInterval
	}
}

func init() {
	if err := env.Parse(cfg); err != nil {
		log.Fatal(fmt.Errorf("failed to parse env vars: %w", err))
	}

	address = flag.String("a", defaultAddress, "Address for sending data to")
	reportInterval = flag.Duration("r", defaultReportInterval, "Metric report to server interval")
	pollInterval = flag.Duration("p", defaultPollInterval, "Metric poll interval")
	flag.Parse()

	overrideConfig(cfg)
}

func main() {
	agent := agent.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go agent.Run(ctx)

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()
	time.Sleep(200 *time.Millisecond)
}
