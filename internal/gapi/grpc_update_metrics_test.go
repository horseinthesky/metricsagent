package gapi

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
	"github.com/stretchr/testify/require"
)

var testsUpdateMetric = []struct {
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
	for _, tt := range testsUpdateMetric {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, closer := tt.runner(ctx)
			defer closer()

			tt.metric.Delta = &tt.value
			payload := MetricToPB(tt.metric)

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
	for _, tt := range testsUpdateMetric {
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
