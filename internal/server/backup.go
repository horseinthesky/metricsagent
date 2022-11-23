package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

// Backuper provides metrics backups to filesystem.
// Only used when in-memory DB is in use.
type Backuper struct {
	filename string
}

// NewBackuper is a Backuper constructor.
func NewBackuper(filename string) *Backuper {
	return &Backuper{
		filename: filename,
	}
}

// WriteMetrics aves metrics to filesystem.
func (b Backuper) WriteMetrics(metrics []storage.Metric) error {
	file, err := os.OpenFile(b.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	return encoder.Encode(&metrics)
}

// ReadMetrics reads metrics from filesystem.
func (b Backuper) ReadMetrics() ([]storage.Metric, error) {
	file, err := os.OpenFile(b.filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	metrics := []storage.Metric{}
	if err := decoder.Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

// dump is a Server's method to save metrics from DB to filesystem.
// Only used if in-memory DB is in use.
// Uses Backuper to do his job.
func (s *Server) dump() {
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
func (s *Server) restore() {
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
