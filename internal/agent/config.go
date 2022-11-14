package agent

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	Pprof          string        `env:"PPROF" envDefault:"localhost:9000"`
	Key            string        `env:"KEY"`
}

func ParseConfig() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse env vars: %w", err)
	}

	flag.StringVar(&cfg.Address, "a", cfg.Address, "Address for sending data to")
	flag.DurationVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "Metric report to server interval")
	flag.DurationVar(&cfg.PollInterval, "p", cfg.PollInterval, "Metric poll interval")
	flag.StringVar(&cfg.Pprof, "P", cfg.Pprof, "Pprof address")
	flag.StringVar(&cfg.Key, "k", "", "Hash key")
	flag.Parse()

	return cfg, nil
}
