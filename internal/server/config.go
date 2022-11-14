package server

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

const (
	defaultListenOn      = "localhost:8080"
	defaultRestoreFlag   = true
	defaultStoreInterval = 300 * time.Second
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
)

type Config struct {
	Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
	Key           string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
}

func ParseConfig() (Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse env vars: %w", err)
	}

	flag.StringVar(&cfg.Address, "a", cfg.Address, "Socket to listen on")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "If should restore metrics on startup")
	flag.DurationVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "backup interval (seconds)")
	flag.StringVar(&cfg.StoreFile, "f", cfg.StoreFile, "Metrics backup file path")
	flag.StringVar(&cfg.Key, "k", "", "Hash key")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database address")
	flag.Parse()

	return cfg, nil
}

