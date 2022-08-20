package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/caarlos0/env/v6"

	"github.com/horseinthesky/metricsagent/internal/agent"
)

type Config struct {
	Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
}

var (
	cfg  Config
	data = &runtime.MemStats{}
)

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse env vars: %w", err))
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
