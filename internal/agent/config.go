package agent

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

// Agent default config options.
const (
	defaultAddress        = "localhost:8080"
	defaultReportInterval = 10 * time.Second
	defaultPollInterval   = 2 * time.Second
	defaultPprofAddress   = "localhost:9000"
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
	Address        string   `json:"address"`
	ReportInterval Duration `json:"report_interval"`
	PollInterval   Duration `json:"poll_interval"`
	CryptoKey      string   `json:"crypto_key"`
}

// Agent Config description.
type Config struct {
	ConfigPath     string        `env:"CONFIG"`
	Address        string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	Pprof          string        `env:"PPROF"`
	Key            string        `env:"KEY"`
	CryptoKey      string        `env:"CRYPTO_KEY"`
	GRPC          bool
}

// ParseConfig parses the configuration options.
// Env variables override flag values.
// Default values are used if nothing mentioned above provided.
func ParseConfig() (Config, error) {
	cfg := Config{}

	flag.StringVar(&cfg.ConfigPath, "c", "", "Config file path")
	flag.StringVar(&cfg.Address, "a", defaultAddress, "Address for sending data to")
	flag.DurationVar(&cfg.ReportInterval, "r", defaultReportInterval, "Metric report to server interval")
	flag.DurationVar(&cfg.PollInterval, "p", defaultPollInterval, "Metric poll interval")
	flag.StringVar(&cfg.Pprof, "P", defaultPprofAddress, "Pprof address")
	flag.StringVar(&cfg.Key, "k", "", "Hash key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Crypto public key path")
	flag.BoolVar(&cfg.GRPC, "g", false, "Replace HTTP with gRPC")
	flag.Parse()

	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse env vars: %w", err)
	}

	err := loadConfigFile(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to load config file: %w", err)
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

	if cfg.Address == defaultAddress && cfgFromFile.Address != "" {
		cfg.Address = cfgFromFile.Address
	}

	if cfg.ReportInterval == defaultReportInterval && cfgFromFile.ReportInterval.Duration != 0 {
		cfg.ReportInterval = cfgFromFile.ReportInterval.Duration
	}

	if cfg.PollInterval == defaultPollInterval && cfgFromFile.PollInterval.Duration != 0 {
		cfg.PollInterval = cfgFromFile.PollInterval.Duration
	}

	if cfg.CryptoKey == "" && cfgFromFile.CryptoKey != "" {
		cfg.CryptoKey = cfgFromFile.CryptoKey
	}

	return nil
}
