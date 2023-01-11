package gapi

import (
	"context"
	"crypto/hmac"
	"encoding/hex"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {

	metric := MetricFromPB(req.Metric)
	if storage.UnsupportedType(metric.MType) {
		return nil, status.Error(codes.Unimplemented, "unsupported metric type")
	}

	if s.Config.Key != "" {
		localHash := server.GenerateHash(metric, s.Config.Key)
		remoteHash, err := hex.DecodeString(metric.Hash)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to decode hash")
		}

		if !hmac.Equal(localHash, remoteHash) {
			return nil, status.Error(codes.Internal, "invalid hash")
		}
	}


	err := s.SaveMetric(metric)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to store metric")
	}

	return &pb.UpdateMetricResponse{}, nil
}
