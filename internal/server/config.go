package server

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

// Server default cofig options.
const (
	defaultListenOn      = "localhost:8080"
	defaultRestoreFlag   = true
	defaultStoreInterval = 300 * time.Second
	defaultStoreFile     = "/tmp/devops-metrics-db.json"
)

// Server Agent Config description.
type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
	Key           string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
}

// ParseConfig parses the configuration options.
// Env variables override flag values.
// Default values are used if nothing mentioned above provided.
func ParseConfig() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.Address, "a", defaultListenOn, "Socket to listen on")
	flag.BoolVar(&cfg.Restore, "r", defaultRestoreFlag, "If should restore metrics on startup")
	flag.DurationVar(&cfg.StoreInterval, "i", defaultStoreInterval, "backup interval (seconds)")
	flag.StringVar(&cfg.StoreFile, "f", defaultStoreFile, "Metrics backup file path")
	flag.StringVar(&cfg.Key, "k", "", "Hash key")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database address")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse env vars: %w", err)
	}

	return cfg, nil
}
