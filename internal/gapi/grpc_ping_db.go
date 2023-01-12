package gapi

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *GRPCServer) PingDB(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.DB.Check(ctx); err != nil {
		return nil, status.Error(codes.Internal, "failed to ping DB")
	}

	return &emptypb.Empty{
	}, nil
}
