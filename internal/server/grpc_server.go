package server

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/horseinthesky/metricsagent/internal/crypto"
	"github.com/horseinthesky/metricsagent/internal/pb"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	pb.UnimplementedMetricsAgentServer
	config    Config
	CryptoKey *rsa.PrivateKey
	db        storage.Storage
	backuper  *Backuper
	workGroup sync.WaitGroup
}

func NewGRPCServer(cfg Config) (*GRPCServer, error) {
	var privKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		var err error

		privKey, err = crypto.ParsePrivKey(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
	}

	var db storage.Storage
	if cfg.DatabaseDSN != "" {
		db = storage.NewDBStorage(cfg.DatabaseDSN)
	} else {
		db = storage.NewMemoryStorage()
	}

	backuper := NewBackuper(cfg.StoreFile)

	server := &GRPCServer{
		config:    cfg,
		CryptoKey: privKey,
		db:        db,
		backuper:  backuper,
		workGroup: sync.WaitGroup{},
	}

	return server, nil
}

func (s *GRPCServer) Run(ctx context.Context) {
	if s.config.DatabaseDSN == "" {
		// Restore metrics from backup
		if s.config.Restore {
			s.restore()
		}

		// Backup metrics periodically
		if s.config.StoreFile != "" && s.config.StoreInterval > time.Duration(0)*time.Second {
			s.workGroup.Add(1)
			go func() {
				defer s.workGroup.Done()
				s.startPeriodicMetricsDump(ctx)
			}()
		}
	}

	err := s.db.Init(ctx)
	if err != nil {
		log.Fatalf("failed to init db: %s", err)
	}

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(s.protectInterceptor))
	pb.RegisterMetricsAgentServer(grpcServer, s)
	reflection.Register(grpcServer)

	s.workGroup.Add(1)
	go func() {
		defer s.workGroup.Done()

		runMsg := fmt.Sprintf("Running gRPC server, listening on %s", s.config.Address)
		if s.config.TrustedSubnet != "" {
			addon := fmt.Sprintf(", trusted subnet: %s", s.config.TrustedSubnet)
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

	s.db.Close()
	log.Println("connection to database closed")

	s.workGroup.Wait()
	log.Println("successfully shut down")
}

// startPeriodicMetricsDump handles Server periodic metrics backup to file.
// Only used with memory DB to provide persistent metrics storage
// between Server restart.
func (s *GRPCServer) startPeriodicMetricsDump(ctx context.Context) {
	log.Println("pediodic metrics backup started")

	ticker := time.NewTicker(s.config.StoreInterval)

	for {
		select {
		case <-ticker.C:
			s.dump()
		case <-ctx.Done():
			log.Println("metrics backup canceled")
			return
		}
	}
}

// saveMetric handles synchronous metric backup.
// Only used by handleSaveJSONMetric handler when
//   - in-memory storage is in use
//   - no StoreInterval provided
func (s *GRPCServer) saveMetric(metric storage.Metric) error {
	err := s.db.Set(metric)

	// if s.config.DatabaseDSN == "" {
	// 	if s.config.StoreFile != "" && s.config.StoreInterval == time.Duration(0) {
	// 		s.dump()
	// 	}
	// }

	return err
}

// saveMetricBulk handles synchronous bulk metric backup.
// Only used by handleSaveJSONMetrics handler when
//   - in-memory storage is in use
//   - no StoreInterval provided
func (s *GRPCServer) saveMetricsBulk(metrics []storage.Metric) error {
	err := s.db.SetBulk(metrics)

	// if s.config.DatabaseDSN == "" {
	// 	if s.config.StoreFile != "" && s.config.StoreInterval == time.Duration(0) {
	// 		s.dump()
	// 	}
	// }

	return err
}
