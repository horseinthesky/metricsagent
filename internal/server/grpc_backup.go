package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

// dump is a Server's method to save metrics from DB to filesystem.
// Only used if in-memory DB is in use.
// Uses Backuper to do his job.
func (s *GRPCServer) dump() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	allMetrics, err := s.db.GetAll(ctx)
	if err != nil {
		log.Printf("failed to get stored metrics: %s", err)
		return
	}

	var metrics []storage.Metric

	for _, metric := range allMetrics {
		metrics = append(metrics, metric)
	}

	if err := s.backuper.WriteMetrics(metrics); err != nil {
		log.Println(fmt.Errorf("failed to dump metrics to %s: %w", s.backuper.filename, err))
		return
	}

	log.Printf("successfully dumped all metrics to %s", s.backuper.filename)
}

// restore is a Server's metohd to restore metrics from filesystem to DB.
// Only used if in-memory DB is in use.
// Uses Backuper to do his job.
func (s *GRPCServer) restore() {
	metrics, err := s.backuper.ReadMetrics()
	if err != nil {
		log.Println(fmt.Errorf("failed to restore metrics from %s: %w", s.backuper.filename, err))
		return
	}

	for _, metric := range metrics {
		err := s.db.Set(metric)
		if err != nil {
			log.Println(fmt.Errorf("failed to restore metric %s: %w", metric.ID, err))
		}
	}

	log.Printf("successfully restored all metrics from %s", s.backuper.filename)
}
