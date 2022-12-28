package agent

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/horseinthesky/metricsagent/internal/crypto"
)

// GPRCAgent description.
type GRPCAgent struct {
	PollTicker   *time.Ticker
	ReportTicker *time.Ticker
	PollCounter  int64
	key          string
	CryptoKey    *rsa.PublicKey
	metrics      *sync.Map
	upstream     string
	workGroup    sync.WaitGroup
}

// NewAgent is an Agent constructor.
// Sets things up.
func NewGRPCAgent(cfg Config) (*Agent, error) {
	var pubKey *rsa.PublicKey
	if cfg.CryptoKey != "" {
		var err error

		pubKey, err = crypto.ParsePubKey(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
	}

	return &Agent{
		PollTicker:   time.NewTicker(cfg.PollInterval),
		ReportTicker: time.NewTicker(cfg.ReportInterval),
		key:          cfg.Key,
		CryptoKey:    pubKey,
		metrics:      &sync.Map{},
		upstream:     fmt.Sprintf("http://%s", cfg.Address),
	}, nil
}

// Run is an Agent starting point.
// Runs an agent.
func (a *GRPCAgent) Run(ctx context.Context) {
	a.workGroup.Add(3)
	go func() {
		defer a.workGroup.Done()
		a.collectRuntimeMetrics(ctx)
	}()
	go func() {
		defer a.workGroup.Done()
		a.collectPSUtilMetrics(ctx)
	}()
	// go func() {
	// 	defer a.workGroup.Done()
	// 	a.sendMetricsJSONBulk(ctx)
	// }()

	<-ctx.Done()
	log.Println("shutting down...")
}

// collectPSUtilMetrics runs updatePSUtilMetrics every config.PollInterval.
// Also handles graceful shutdown.
func (a *GRPCAgent) collectPSUtilMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("psutil data collection cancelled")
			return
		case <-a.PollTicker.C:
			updatePSUtilMetrics(a.metrics)

			log.Println("successfully collected psutil metrics")
		}
	}
}

// collectRuntimeMetrics runs updateRuntimeMetrics every config.PollInterval.
// Also handles graceful shutdown.
func (a *GRPCAgent) collectRuntimeMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("runtime data collection cancelled")
			return
		case <-a.PollTicker.C:
			updateRuntimeMetrics(a.metrics)

			a.PollCounter++

			log.Println("successfully collected runtime metrics")
		}
	}
}

// Stop is an Agent graceful shutdown method.
// Ensures everything is stopped as expected.
func (a *GRPCAgent) Stop() {
	a.workGroup.Wait()
	log.Println("successfully shut down")
}
