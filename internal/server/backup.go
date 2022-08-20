package server

import (
	"encoding/json"
	"os"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filename string, flags int) (*producer, error) {
	file, err := os.OpenFile(filename, flags, 0777)
	if err != nil {
		return nil, err
	}

	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p producer) Close() error {
	return p.file.Close()
}

func (p producer) WriteMetrics(metrics *[]storage.Metric) error {
	defer p.Close()

	return p.encoder.Encode(&metrics)
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(filename string, flags int) (*consumer, error) {
	file, err := os.OpenFile(filename, flags, 0777)
	if err != nil {
		return nil, err
	}

	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *consumer) ReadMetrics() ([]storage.Metric, error) {
	defer c.Close()

	metrics := []storage.Metric{}
	if err := c.decoder.Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

func (c consumer) Close() error {
	return c.file.Close()
}
