package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
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
	data           = &runtime.MemStats{}
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

	for {
		select {
		case <-agent.ReportTicker.C:
			agent.SendMetricsJSON()
		case <-agent.PollTicker.C:
			agent.PollCounter++

			runtime.ReadMemStats(data)

			agent.UpdateMetrics(data)
		}
	}
}
