package server

import (
	"context"
	"encoding/hex"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) LoadMetric(ctx context.Context, req *pb.LoadMetricRequest) (*pb.LoadMetricResponse, error) {
	metricRequest := storage.Metric{
		ID:    req.Metric.Id,
		MType: req.Metric.Mtype,
	}

	if storage.UnsupportedType(metricRequest.MType) {
		return nil, status.Error(codes.Unimplemented, "unsupported metric type")
	}

	metric, err := s.db.Get(ctx, metricRequest.ID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "unknown metric id")
	}

	if s.config.Key != "" {
		metric.Hash = hex.EncodeToString(generateHash(metric, s.config.Key))
	}

	return &pb.LoadMetricResponse{
		Metric: MetricToPB(metric),
	}, nil
}
