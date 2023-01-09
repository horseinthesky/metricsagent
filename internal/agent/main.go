package agent

import (
	"context"
	"crypto/rsa"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/horseinthesky/metricsagent/internal/crypto"
)

// Agent description.
type GenericAgent struct {
	PollTicker   *time.Ticker
	ReportTicker *time.Ticker
	PollCounter  int64
	pprofServer  *http.Server
	key          string
	CryptoKey    *rsa.PublicKey
	metrics      *sync.Map
	upstream     string
	workGroup    sync.WaitGroup
}

// NewAgent is an Agent constructor.
// Sets things up.
func NewGenericAgent(cfg Config) (*GenericAgent, error) {
	var pubKey *rsa.PublicKey
	if cfg.CryptoKey != "" {
		var err error

		pubKey, err = crypto.ParsePubKey(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
	}

	return &GenericAgent{
		PollTicker:   time.NewTicker(cfg.PollInterval),
		ReportTicker: time.NewTicker(cfg.ReportInterval),
		pprofServer:  &http.Server{Addr: cfg.Pprof},
		key:          cfg.Key,
		CryptoKey:    pubKey,
		metrics:      &sync.Map{},
	}, nil
}

// Run is an Agent starting point.
// Runs an agent.
func (a *GenericAgent) Collect(ctx context.Context) {
	a.workGroup.Add(3)
	go func() {
		defer a.workGroup.Done()
		a.pprofServer.ListenAndServe()
	}()
	go func() {
		defer a.workGroup.Done()
		collectRuntimeMetrics(ctx, a.PollTicker, a.metrics, &a.PollCounter)
	}()
	go func() {
		defer a.workGroup.Done()
		collectPSUtilMetrics(ctx, a.PollTicker, a.metrics)
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	a.pprofServer.Shutdown(ctx)
}
