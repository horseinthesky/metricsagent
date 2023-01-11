package server

import (
	"context"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) PingDB(ctx context.Context, req *pb.PingDBRequest) (*pb.PingDBResponse, error) {
	if err := s.DB.Check(ctx); err != nil {
		return nil, status.Error(codes.Internal, "failed to ping DB")
	}

	return &pb.PingDBResponse{
	}, nil
}
