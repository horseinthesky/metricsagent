package agent

import (
	"context"
	"crypto/rsa"
	"log"
	"sync"
	"time"

	"github.com/horseinthesky/metricsagent/internal/crypto"
	"github.com/horseinthesky/metricsagent/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GPRCAgent description.
type GRPCAgent struct {
	PollTicker   *time.Ticker
	ReportTicker *time.Ticker
	PollCounter  int64
	key          string
	CryptoKey    *rsa.PublicKey
	metrics      *sync.Map
	conn         *grpc.ClientConn
	workGroup    sync.WaitGroup
}

// NewAgent is an Agent constructor.
// Sets things up.
func NewGRPCAgent(cfg Config) (*GRPCAgent, error) {
	var pubKey *rsa.PublicKey
	if cfg.CryptoKey != "" {
		var err error

		pubKey, err = crypto.ParsePubKey(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
	}

	conn, err := grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GRPCAgent{
		PollTicker:   time.NewTicker(cfg.PollInterval),
		ReportTicker: time.NewTicker(cfg.ReportInterval),
		key:          cfg.Key,
		CryptoKey:    pubKey,
		metrics:      &sync.Map{},
		conn:         conn,
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
	go func() {
		defer a.workGroup.Done()
		a.sendMetrics(ctx)
	}()

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

// sendMetricsJSONBulk sends all metrics as one big JSON.
func (a *GRPCAgent) sendMetrics(ctx context.Context) {
	client := pb.NewMetricsAgentClient(a.conn)

	for {
		select {
		case <-ctx.Done():
			log.Println("sending metrics cancelled")
			return
		case <-a.ReportTicker.C:
			metrics := prepareMetrics(a.metrics, a.PollCounter, a.key)

			pbMetics := []*pb.Metric{}
			for _, m := range metrics {
				pbMetics = append(pbMetics, MetricToPB(m))
			}

			res, err := client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{
				Metrics: pbMetics,
			})
			if err != nil {
				log.Printf("failed to send metrics: %s", err)
				continue
			}
			if res.Error != "" {
					log.Printf("server rejected metrics: %s", res.Error)
					continue
			}

			log.Println("successfully updated metrics")
		}
	}
}

// Stop is an Agent graceful shutdown method.
// Ensures everything is stopped as expected.
func (a *GRPCAgent) Stop() {
	a.conn.Close()

	a.workGroup.Wait()
	log.Println("successfully shut down")
}
