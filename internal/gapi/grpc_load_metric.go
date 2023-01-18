package gapi

import (
	"context"
	"encoding/hex"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) LoadMetric(ctx context.Context, req *pb.LoadMetricRequest) (*pb.Metric, error) {
	metricRequest := storage.Metric{
		ID:    req.Id,
		MType: req.Mtype,
	}

	if storage.UnsupportedType(metricRequest.MType) {
		return nil, status.Error(codes.Unimplemented, "unsupported metric type")
	}

	metric, err := s.DB.Get(ctx, metricRequest.ID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "unknown metric id")
	}

	if s.Config.Key != "" {
		metric.Hash = hex.EncodeToString(server.GenerateHash(metric, s.Config.Key))
	}

	return MetricToPB(metric), nil
}
