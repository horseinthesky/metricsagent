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

var (
	address        *string
	reportInterval *time.Duration
	pollInterval   *time.Duration
	cfg            = &agent.Config{}
	data           = &runtime.MemStats{}
)

func overrideConfig(cfg *agent.Config) {
	if _, present := os.LookupEnv("ADDRESS"); !present {
		cfg.Address = *address
	}
	if _, present := os.LookupEnv("REPORT_INTERVAL"); !present {
		cfg.ReportInterval = *reportInterval
	}
	if _, present := os.LookupEnv("POLL_INTERVAL"); !present {
		cfg.PollInterval = *pollInterval
	}
}

func init() {
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse env vars: %w", err))
	}

	address = flag.String("a", "localhost:8080", "Address for sending data to")
	reportInterval = flag.Duration("r", time.Duration(10*time.Second), "Metric report to server interval")
	pollInterval = flag.Duration("p", time.Duration(2*time.Second), "Metric poll interval")
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
			agent.Count++

			runtime.ReadMemStats(data)

			agent.UpdateMetrics(data)
		}
	}
}
