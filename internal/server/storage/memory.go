package storage

import (
	"fmt"
	"sync"
)

type Memory struct {
	sync.Mutex
	db map[string]Metric
}

func NewMemoryStorage() *Memory {
	return &Memory{db: map[string]Metric{}}
}

func (m *Memory) Set(metric *Metric) error {
	m.Lock()
	defer m.Unlock()

	switch metric.MType {
	case Counter.String():
		oldMetric, ok := m.db[metric.ID]
		if ok {
			*oldMetric.Delta += *metric.Delta
			return nil

		}
		m.db[metric.ID] = *metric
		return nil
	case Gauge.String():
		m.db[metric.ID] = *metric
		return nil
	}

	return fmt.Errorf("failed to save metric")
}

func (m *Memory) Get(name string) (Metric, error) {
	m.Lock()
	defer m.Unlock()

	metric, ok := m.db[name]
	if !ok {
		return Metric{}, fmt.Errorf("no value found")
	}

	return metric, nil
}

func (m *Memory) GetAll() map[string]Metric {
	m.Lock()
	defer m.Unlock()

	newDb := map[string]Metric{}
	for k, v := range m.db {
		newDb[k] = v
	}

	return newDb
}
