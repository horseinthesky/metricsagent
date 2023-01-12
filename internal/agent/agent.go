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
	"os"
	"os/signal"
	"syscall"
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
func (a *Agent) Run() {
	ctx, cancel := context.WithCancel(context.Background())

	go a.Work(ctx)

	a.workGroup.Add(1)
	go func() {
		defer a.workGroup.Done()
		a.sendMetricsJSONBulk(ctx)
	}()

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()

	a.workGroup.Wait()
	log.Println("successfully shut down")
}
