package gapi

import (
	"context"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
	"github.com/stretchr/testify/require"
)

func runTestServer(ctx context.Context, hashKey string) (pb.MetricsAgentClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	testServer, err := NewGRPCServer(server.Config{Key: hashKey})

	grpcServer := grpc.NewServer()
	pb.RegisterMetricsAgentServer(grpcServer, testServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}

		grpcServer.Stop()
	}

	client := pb.NewMetricsAgentClient(conn)

	return client, closer
}

var tests = []struct {
	name    string
	runner  func(context.Context) (pb.MetricsAgentClient, func())
	value   int64
	metric  storage.Metric
	payload *pb.UpdateMetricsRequest
	error   error
}{
	{
		name: "test unimplemented mectic type",
		runner: func(ctx context.Context) (pb.MetricsAgentClient, func()) {
			return runTestServer(ctx, "")
		},
		value: int64(1),
		metric: storage.Metric{
			ID:    "wrongCounter",
			MType: "unimplemented",
		},
		error: status.Error(codes.Unimplemented, ""),
	},
	{
		name: "test update metrics",
		runner: func(ctx context.Context) (pb.MetricsAgentClient, func()) {
			return runTestServer(ctx, "")
		},
		value: int64(10),
		metric: storage.Metric{
			ID:    "testCounter1",
			MType: "counter",
		},
		error: nil,
	},
	{
		name: "test update metrics hashed",
		runner: func(ctx context.Context) (pb.MetricsAgentClient, func()) {
			return runTestServer(ctx, "testkey")
		},
		value: int64(15),
		metric: storage.Metric{
			ID:    "TestCounter",
			MType: "counter",
			Hash:  "175b2a772fbf2ad97bb515e10f2c24bdaf75860e18f8999c6825be73acd3e6bc",
		},
		error: nil,
	},
	{
		name: "test update metrics invalid hash",
		runner: func(ctx context.Context) (pb.MetricsAgentClient, func()) {
			return runTestServer(ctx, "testkey")
		},
		value: int64(15),
		metric: storage.Metric{
			ID:    "TestCounter",
			MType: "counter",
			Hash:  "dfmsdskgdfuf",
		},
		error: status.Error(codes.Internal, ""),
	},
	{
		name: "test update metrics wrong hash",
		runner: func(ctx context.Context) (pb.MetricsAgentClient, func()) {
			return runTestServer(ctx, "testkey")
		},
		value: int64(15),
		metric: storage.Metric{
			ID:    "TestCounter",
			MType: "counter",
			Hash:  "175b2a772fbf2ad97bb515e10f2c24bdaf75860e18f8999c6825be73acd3e6bd",
		},
		error: status.Error(codes.Internal, ""),
	},
}

func TestUpdateMetric(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, closer := tt.runner(ctx)
			defer closer()

			tt.metric.Delta = &tt.value
			payload := &pb.UpdateMetricRequest{Metric: MetricToPB(tt.metric)}

			_, err := client.UpdateMetric(ctx, payload)
			if tt.error == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)

			e, _ := status.FromError(err)
			expectedErr, _ := status.FromError(tt.error)

			require.Equal(t, expectedErr.Code(), e.Code())
		})
	}
}
func TestUpdateMetrics(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, closer := tt.runner(ctx)
			defer closer()

			tt.metric.Delta = &tt.value
			payload := &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{MetricToPB(tt.metric)}}

			_, err := client.UpdateMetrics(ctx, payload)
			if tt.error == nil {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)

			e, _ := status.FromError(err)
			expectedErr, _ := status.FromError(tt.error)

			require.Equal(t, expectedErr.Code(), e.Code())
		})
	}
}
