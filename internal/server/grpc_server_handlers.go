package server

import (
	"context"
	"crypto/hmac"
	"encoding/hex"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponce, error) {
	metrics := []storage.Metric{}

	for _, pbMetric := range req.Metrics {
		metric := MetricFromPB(pbMetric)

		if storage.UnsupportedType(metric.MType) {
			return nil, status.Error(codes.Unimplemented, "unsupported metric type")
		}

		if s.config.Key != "" {
			localHash := generateHash(metric, s.config.Key)
			remoteHash, err := hex.DecodeString(metric.Hash)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to decode hash")
			}

			if !hmac.Equal(localHash, remoteHash) {
				return nil, status.Error(codes.Internal, "invalid hash")
			}
		}

		metrics = append(metrics, metric)
	}

	err := s.saveMetricsBulk(metrics)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to store metric")
	}

	return &pb.UpdateMetricsResponce{}, nil
}
func (s *GRPCServer) LoadMetric(ctx context.Context, req *pb.LoadMetricRequest) (*pb.LoadMetricResponce, error) {
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

	return &pb.LoadMetricResponce{
		Metric: MetricToPB(metric),
	}, nil
}
