package gapi

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	*server.GenericServer
	pb.UnimplementedMetricsAgentServer
}

func NewGRPCServer(cfg server.Config) (*GRPCServer, error) {
	genericServer, err := server.NewGenericServer(cfg)
	if err != nil {
		return nil, err
	}

	server := &GRPCServer{
		genericServer,
		pb.UnimplementedMetricsAgentServer{},
	}

	return server, nil
}

func (s *GRPCServer) Run(ctx context.Context) {
	s.Bootstrap(ctx)

	listener, err := net.Listen("tcp", s.Config.Address)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(s.protectInterceptor))
	pb.RegisterMetricsAgentServer(grpcServer, s)
	reflection.Register(grpcServer)

	s.WorkGroup.Add(1)
	go func() {
		defer s.WorkGroup.Done()

		runMsg := fmt.Sprintf("Running gRPC server, listening on %s", s.Config.Address)
		if s.Config.TrustedSubnet != "" {
			addon := fmt.Sprintf(", trusted subnet: %s", s.Config.TrustedSubnet)
			runMsg += addon
		}
		log.Println(runMsg)

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("server crashed: %s", err)
		}

		log.Printf("finished to serve gRPC requests")
	}()

	<-ctx.Done()
	grpcServer.GracefulStop()
}

func (s *GRPCServer) Stop() {
	log.Println("shutting down...")

	s.DB.Close()
	log.Println("connection to database closed")

	s.WorkGroup.Wait()
	log.Println("successfully shut down")
}
