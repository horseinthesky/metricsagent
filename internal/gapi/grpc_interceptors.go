package gapi

import (
	"context"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) protectInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if s.Config.TrustedSubnet == "" {
		return handler(ctx, req)
	}

	requestIP, err := getClientIP(ctx)
	if requestIP == nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, trustedNet, _ := net.ParseCIDR(s.Config.TrustedSubnet)

	if !trustedNet.Contains(requestIP) {
		return nil, status.Errorf(codes.PermissionDenied, "request from %s is forbidden", requestIP.String())
	}

	return handler(ctx, req)
}

// GetClientIP inspects the context to retrieve the ip address of the client
func getClientIP(ctx context.Context) (net.IP, error) {
	addrPort, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no peer address info found")
	}

	ip := strings.Split(addrPort.Addr.String(), ":")[0]

	return net.ParseIP(ip), nil
}
