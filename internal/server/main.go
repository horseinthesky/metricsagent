package server

import (
	"context"
	"crypto/rsa"
	"log"
	"sync"
	"time"

	"github.com/horseinthesky/metricsagent/internal/crypto"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

// GenericServer main struct.
type GenericServer struct {
	Config    Config
	CryptoKey *rsa.PrivateKey
	DB        storage.Storage
	backuper  *Backuper
	WorkGroup sync.WaitGroup
}

// Server constructor.
// Sets things up.
func NewGenericServer(cfg Config) (*GenericServer, error) {
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
		db = storage.NewDBStorage(cfg.DatabaseDriver, cfg.DatabaseDSN)
	} else {
		db = storage.NewMemoryStorage()
	}

	backuper := NewBackuper(cfg.StoreFile)

	server := &GenericServer{cfg, privKey, db, backuper, sync.WaitGroup{}}

	return server, nil
}

func (s *GenericServer) Bootstrap(ctx context.Context) {
	if s.Config.DatabaseDSN == "" {
		// Restore metrics from backup
		if s.Config.Restore {
			s.restore()
		}

		// Backup metrics periodically
		if s.Config.StoreFile != "" && s.Config.StoreInterval > time.Duration(0)*time.Second {
			s.WorkGroup.Add(1)
			go func() {
				defer s.WorkGroup.Done()
				s.startPeriodicMetricsDump(ctx)
			}()
		}
	}

	err := s.DB.Init(ctx)
	if err != nil {
		log.Fatalf("failed to init db: %s", err)
	}
}

// startPeriodicMetricsDump handles Server periodic metrics backup to file.
// Only used with memory DB to provide persistent metrics storage
// between Server restart.
func (s *GenericServer) startPeriodicMetricsDump(ctx context.Context) {
	log.Println("pediodic metrics backup started")

	ticker := time.NewTicker(s.Config.StoreInterval)

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
func (s *GenericServer) SaveMetric(metric storage.Metric) error {
	err := s.DB.Set(metric)

	if s.Config.DatabaseDSN == "" {
		if s.Config.StoreFile != "" && s.Config.StoreInterval == time.Duration(0) {
			s.dump()
		}
	}

	return err
}

// saveMetricBulk handles synchronous bulk metric backup.
// Only used by handleSaveJSONMetrics handler when
//   - in-memory storage is in use
//   - no StoreInterval provided
func (s *GenericServer) SaveMetricsBulk(metrics []storage.Metric) error {
	err := s.DB.SetBulk(metrics)

	if s.Config.DatabaseDSN == "" {
		if s.Config.StoreFile != "" && s.Config.StoreInterval == time.Duration(0) {
			s.dump()
		}
	}

	return err
}
