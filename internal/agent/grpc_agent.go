package agent

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GPRCAgent description.
type GRPCAgent struct {
	*GenericAgent
	conn *grpc.ClientConn
}

// NewAgent is an Agent constructor.
// Sets things up.
func NewGRPCAgent(cfg Config) (GRPCAgent, error) {
	genericAgent, err := NewGenericAgent(cfg)
	if err != nil {
		return GRPCAgent{}, err
	}

	conn, err := grpc.Dial(cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return GRPCAgent{}, err
	}

	return GRPCAgent{
		genericAgent,
		conn,
	}, nil
}

// Run is an Agent starting point.
// Runs an agent.
func (a *GRPCAgent) Run() {
	ctx, cancel := context.WithCancel(context.Background())

	go a.Work(ctx)

	a.workGroup.Add(1)
	go func() {
		defer a.workGroup.Done()
		a.sendMetrics(ctx)
	}()

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	sig := <-term
	log.Printf("signal received: %v; terminating...\n", sig)

	cancel()

	a.conn.Close()
	a.workGroup.Wait()
	log.Println("successfully shut down")
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
