package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

type Backuper struct {
	filename string
}

func NewBackuper(filename string) *Backuper {
	return &Backuper{
		filename: filename,
	}
}

func (b Backuper) WriteMetrics(metrics []storage.Metric) error {
	file, err := os.OpenFile(b.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	return encoder.Encode(&metrics)
}

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

// Server backup/restomre methods
func (s *Server) dump() {
	var metrics []storage.Metric

	for _, metric := range s.storage.GetAll() {
		metrics = append(metrics, metric)
	}

	if err := s.backuper.WriteMetrics(metrics); err != nil {
		log.Println(fmt.Errorf("failed to dump metrics to %s: %w", s.backuper.filename, err))
		return
	}

	log.Printf("successfully dumped all metrics to %s", s.backuper.filename)
}

func (s *Server) restore() {
	metrics, err := s.backuper.ReadMetrics()
	if err != nil {
		log.Println(fmt.Errorf("failed to restore metrics from %s: %w", s.backuper.filename, err))
		return
	}

	for _, metric := range metrics {
		err := s.storage.Set(metric)
		if err != nil {
			log.Println(fmt.Errorf("failed to restore metric %s: %w", metric.ID, err))
		}
	}

	log.Printf("successfully restored all metrics from %s", s.backuper.filename)
}
