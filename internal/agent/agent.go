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
	"time"
)

// Agent description.
type Agent struct {
	*GenericAgent
	upstream string
	client   *http.Client
}

// NewAgent is an Agent constructor.
// Sets things up.
func NewAgent(cfg Config) (Agent, error) {
	genericAgent, err := NewGenericAgent(cfg)
	if err != nil {
		return Agent{}, err
	}

	return Agent{
		genericAgent,
		fmt.Sprintf("http://%s", cfg.Address),
		&http.Client{Timeout: 1 * time.Second},
	}, nil
}

// Run is an Agent starting point.
// Runs an agent.
func (a *Agent) Run(ctx context.Context) {
	go a.Collect(ctx)

	a.workGroup.Add(1)
	go func() {
		defer a.workGroup.Done()
		a.sendMetricsJSONBulk(ctx)
	}()
}

// Stop is an Agent graceful shutdown method.
// Ensures everything is stopped as expected.
func (a *Agent) Stop() {
	a.workGroup.Wait()
	log.Println("successfully shut down")
}
