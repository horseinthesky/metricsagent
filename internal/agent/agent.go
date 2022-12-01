// Package agent describes metrics agent internals.
//
// It consists of the following parts:
//   - agent.go - agent struct and its lifecycle methods
//   - config.go - agent configuration options
//   - collect.go - agent metrics and collect methods
//   - secure.go - agent metrics hash protection
//   - send.go - agent metrics send methods
package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

// Agent description.
type Agent struct {
	PollTicker   *time.Ticker
	ReportTicker *time.Ticker
	PollCounter  int64
	pprofServer  *http.Server
	key          string
	metrics      *sync.Map
	upstream     string
	client       *http.Client
	workGroup    sync.WaitGroup
}

// NewAgent is an Agent constructor.
// Sets things up.
func NewAgent(cfg Config) *Agent {
	return &Agent{
		PollTicker:   time.NewTicker(cfg.PollInterval),
		ReportTicker: time.NewTicker(cfg.ReportInterval),
		pprofServer:  &http.Server{Addr: cfg.Pprof},
		key:          cfg.Key,
		metrics:      &sync.Map{},
		upstream:     fmt.Sprintf("http://%s", cfg.Address),
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
	}
}

// Run is an Agent starting point.
// Runs an agent.
func (a *Agent) Run(ctx context.Context) {
	a.workGroup.Add(4)
	go func() {
		defer a.workGroup.Done()
		a.pprofServer.ListenAndServe()
	}()
	go func() {
		defer a.workGroup.Done()
		a.collectRuntimeMetrics(ctx)
	}()
	go func() {
		defer a.workGroup.Done()
		a.collectPSUtilMetrics(ctx)
	}()
	go func() {
		defer a.workGroup.Done()
		a.sendMetricsJSONBulk(ctx)
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	a.pprofServer.Shutdown(ctx)
}

// Stop is an Agent graceful shutdown method.
// Ensures everything is stopped as expected.
func (a *Agent) Stop() {
	a.workGroup.Wait()
	log.Println("successfully shut down")
}
