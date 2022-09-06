package storage

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type Memory struct {
	sync.RWMutex
	db map[string]Metric
}

func NewMemoryStorage() *Memory {
	return &Memory{db: map[string]Metric{}}
}

func (m *Memory) Init(ctx context.Context) error {
	log.Println("memory database initialized")

	return nil
}

func (m *Memory) Check(ctx context.Context) error {
	return nil
}

func (m *Memory) Set(metric Metric) error {
	m.Lock()
	defer m.Unlock()

	switch metric.MType {
	case Counter.String():
		oldMetric, ok := m.db[metric.ID]
		if ok {
			*oldMetric.Delta += *metric.Delta
			return nil

		}
		m.db[metric.ID] = metric
		return nil
	case Gauge.String():
		m.db[metric.ID] = metric
		return nil
	}

	return fmt.Errorf("failed to save metric")
}

func (m *Memory) SetBulk(metrics []Metric) error {
	m.Lock()
	defer m.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case Counter.String():
			oldMetric, ok := m.db[metric.ID]
			if ok {
				*oldMetric.Delta += *metric.Delta
				continue

			}
			m.db[metric.ID] = metric
		case Gauge.String():
			m.db[metric.ID] = metric
		}
	}

	return nil
}
func (m *Memory) Get(name string) (Metric, error) {
	m.RLock()
	defer m.RUnlock()

	metric, ok := m.db[name]
	if !ok {
		return Metric{}, fmt.Errorf("no value found")
	}

	return metric, nil
}

func (m *Memory) GetAll() (map[string]Metric, error) {
	m.RLock()
	defer m.RUnlock()

	newDB := map[string]Metric{}
	for k, v := range m.db {
		newDB[k] = v
	}

	return newDB, nil
}

func (m *Memory) Close() {
}
