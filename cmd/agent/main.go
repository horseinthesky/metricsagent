package main

import (
	"log"
	"runtime"

	"github.com/caarlos0/env/v6"

	"github.com/horseinthesky/metricsagent/internal/agent"
)

// Seconds
const (
	baseURL               = "http://localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

var (
	cfg  Config
	data = &runtime.MemStats{}
)

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Address == "" {
		cfg.Address = baseURL
	}

	if cfg.PollInterval == 0 {
		cfg.PollInterval = defaultPollInterval
	}

	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = defaultReportInterval
	}
}

func main() {
	agent := agent.New(cfg.PollInterval, cfg.ReportInterval, cfg.Address)

	for {
		select {
		case <-agent.ReportTicker.C:
			agent.SendMetrics()
		case <-agent.PollTicker.C:
			agent.Count++

			runtime.ReadMemStats(data)

			agent.UpdateMetrics(data)
		}
	}
}
