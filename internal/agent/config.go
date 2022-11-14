package agent

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

const (
	defaultAddress        = "localhost:8080"
	defaultReportInterval = 10 * time.Second
	defaultPollInterval   = 2 * time.Second
)

type Config struct {
	Address        string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	Pprof          string        `env:"PPROF"`
	Key            string        `env:"KEY"`
}

func ParseConfig() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.Address, "a", defaultAddress, "Address for sending data to")
	flag.DurationVar(&cfg.ReportInterval, "r", defaultReportInterval, "Metric report to server interval")
	flag.DurationVar(&cfg.PollInterval, "p", defaultPollInterval, "Metric poll interval")
	flag.StringVar(&cfg.Pprof, "P", cfg.Pprof, "Pprof address")
	flag.StringVar(&cfg.Key, "k", "", "Hash key")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse env vars: %w", err)
	}

	return cfg, nil
}
