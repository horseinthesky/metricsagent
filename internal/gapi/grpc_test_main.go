package gapi

import (
	"context"
	"log"
	"net"

	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func runTestServer(ctx context.Context, hashKey string) (pb.MetricsAgentClient, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	testServer, _ := NewGRPCServer(server.Config{Key: hashKey})

	grpcServer := grpc.NewServer()
	pb.RegisterMetricsAgentServer(grpcServer, testServer)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}

		grpcServer.Stop()
	}

	client := pb.NewMetricsAgentClient(conn)

	return client, closer
}
