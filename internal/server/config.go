package server

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
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

// ConfigFile is a container to store config file data
type ConfigFile struct {
	Address       string   `json:"address"`
	Restore       bool     `json:"restore"`
	TrustedSubnet string   `json:"trusted_subnet"`
	StoreInterval Duration `json:"store_interval"`
	StoreFile     string   `json:"store_file"`
	CryptoKey     string   `json:"crypto_key"`
	DatabaseDSN   string   `json:"database_dsn"`
}

// Server Agent Config description.
type Config struct {
	ConfigPath    string        `env:"CONFIG"`
	Address       string        `env:"ADDRESS"`
	Restore       bool          `env:"RESTORE"`
	TrustedSubnet string        `env:"TRUSTED_SUBNET"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Key           string        `env:"KEY"`
	CryptoKey     string        `env:"CRYPTO_KEY"`
	DatabaseDSN   string        `env:"DATABASE_DSN"`
	GRPC          bool
}

// ParseConfig parses the configuration options.
// Env variables override flag values.
// Default values are used if nothing mentioned above provided.
func ParseConfig() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.ConfigPath, "c", "", "Config file path")
	flag.StringVar(&cfg.Address, "a", defaultListenOn, "Socket to listen on")
	flag.BoolVar(&cfg.Restore, "r", defaultRestoreFlag, "If should restore metrics on startup")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "Trusted subnet of IPs to accept requests from")
	flag.DurationVar(&cfg.StoreInterval, "i", defaultStoreInterval, "backup interval (seconds)")
	flag.StringVar(&cfg.StoreFile, "f", defaultStoreFile, "Metrics backup file path")
	flag.StringVar(&cfg.Key, "k", "", "Hash key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Crypto private key path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database address")
	flag.BoolVar(&cfg.GRPC, "g", false, "Replace HTTP with gRPC")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse env vars: %w", err)
	}

	err := loadConfigFile(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load config file: %w", err)
	}

	if cfg.TrustedSubnet != "" {
		_, _, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

func loadConfigFile(cfg *Config) error {
	if cfg.ConfigPath == "" {
		return nil
	}

	configBytes, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return err
	}

	var cfgFromFile ConfigFile
	err = json.Unmarshal(configBytes, &cfgFromFile)
	if err != nil {
		return err
	}

	if cfg.Address == defaultListenOn && cfgFromFile.Address != "" {
		cfg.Address = cfgFromFile.Address
	}

	if cfg.Restore && !cfgFromFile.Restore {
		cfg.Restore = cfgFromFile.Restore
	}

	if cfg.TrustedSubnet == "" && cfgFromFile.TrustedSubnet != "" {
		cfg.TrustedSubnet = cfgFromFile.TrustedSubnet
	}

	if cfg.StoreInterval == defaultStoreInterval && cfgFromFile.StoreInterval.Duration != 0 {
		cfg.StoreInterval = cfgFromFile.StoreInterval.Duration
	}

	if cfg.StoreFile == defaultStoreFile && cfgFromFile.StoreFile != "" {
		cfg.StoreFile = cfgFromFile.StoreFile
	}

	if cfg.CryptoKey == "" && cfgFromFile.CryptoKey != "" {
		cfg.CryptoKey = cfgFromFile.CryptoKey
	}

	if cfg.DatabaseDSN == "" && cfgFromFile.DatabaseDSN != "" {
		cfg.DatabaseDSN = cfgFromFile.DatabaseDSN
	}

	return nil
}
