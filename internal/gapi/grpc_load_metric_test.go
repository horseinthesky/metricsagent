package gapi

import (
	"context"
	"testing"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var testsLoadMetric = []struct {
	name   string
	runner func(context.Context) (pb.MetricsAgentClient, func())
	metric *pb.MetricInfo
	error  error
}{
	{
		name: "test unimplemented mectic type",
		runner: func(ctx context.Context) (pb.MetricsAgentClient, func()) {
			return runTestServer(ctx, "testkey")
		},
		metric: &pb.MetricInfo{
			Id:    "wrongCounter",
			Mtype: "unimplemented",
		},
		error: status.Error(codes.Unimplemented, ""),
	},
	{
		name: "test update metrics",
		runner: func(ctx context.Context) (pb.MetricsAgentClient, func()) {
			return runTestServer(ctx, "testkey")
		},
		metric: &pb.MetricInfo{
			Id:    "testCounter1",
			Mtype: "counter",
		},
		error: status.Error(codes.NotFound, ""),
	},
}

func TestLoadMetrics(t *testing.T) {
	for _, tt := range testsLoadMetric {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, closer := tt.runner(ctx)
			defer closer()

			_, err := client.LoadMetric(ctx, &pb.LoadMetricRequest{Metric: tt.metric})
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
