package server

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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

// Duration is a custom type to help unmarshal time.Duration
type Duration struct {
	time.Duration
}

// ConfigFile is a container to store config file data
type ConfigFile struct {
	Address       string   `json:"address"`
	Restore       bool     `json:"restore"`
	StoreInterval Duration `json:"store_interval"`
	StoreFile     string   `josn:"store_file"`
	CryptoKey     string   `json:"crypto_key"`
	DatabaseDSN   string   `json:"database_dsn"`
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var unmarshalledJSON interface{}

	err := json.Unmarshal(b, &unmarshalledJSON)
	if err != nil {
		return err
	}

	switch value := unmarshalledJSON.(type) {
	case float64:
		d.Duration = time.Duration(value)
	case string:
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid duration: %#v", unmarshalledJSON)
	}

	return nil
}

// Server Agent Config description.
type Config struct {
	ConfigPath    string        `env:"CONFIG"`
	Address       string        `env:"ADDRESS"`
	Restore       bool          `env:"RESTORE"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Key           string        `env:"KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
}

// ParseConfig parses the configuration options.
// Env variables override flag values.
// Default values are used if nothing mentioned above provided.
func ParseConfig() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.ConfigPath, "c", "", "Config file path")
	flag.StringVar(&cfg.Address, "a", defaultListenOn, "Socket to listen on")
	flag.BoolVar(&cfg.Restore, "r", defaultRestoreFlag, "If should restore metrics on startup")
	flag.DurationVar(&cfg.StoreInterval, "i", defaultStoreInterval, "backup interval (seconds)")
	flag.StringVar(&cfg.StoreFile, "f", defaultStoreFile, "Metrics backup file path")
	flag.StringVar(&cfg.Key, "k", "", "Hash key")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database address")
	flag.Parse()

	if cfg.ConfigPath != "" {
		configBytes, err := os.ReadFile(cfg.ConfigPath)
		if err != nil {
			return Config{}, fmt.Errorf("Error reading confg file: %w", err)
		}

		var cfgFromFile ConfigFile
		err = json.Unmarshal(configBytes, &cfgFromFile)
		if err != nil {
			return Config{}, fmt.Errorf("Error parsing config file: %w", err)
		}

		if cfg.Address == defaultListenOn && cfgFromFile.Address != "" {
			cfg.Address = cfgFromFile.Address
		}

		if cfg.Restore == defaultRestoreFlag && cfgFromFile.Restore != true {
			cfg.Restore = cfgFromFile.Restore
		}

		if cfg.StoreInterval == defaultStoreInterval && cfgFromFile.StoreInterval.Duration == 0 {
			cfg.StoreInterval = cfgFromFile.StoreInterval.Duration
		}

		if cfg.StoreFile == defaultStoreFile && cfgFromFile.StoreFile != "" {
			cfg.StoreFile = cfgFromFile.StoreFile
		}

		if cfg.Key == "" && cfgFromFile.CryptoKey != "" {
			cfg.Key = cfgFromFile.CryptoKey
		}

		if cfg.DatabaseDSN == "" && cfgFromFile.DatabaseDSN != "" {
			cfg.DatabaseDSN = cfgFromFile.DatabaseDSN
		}
	}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse env vars: %w", err)
	}

	return cfg, nil
}
