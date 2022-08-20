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

	return nil
}

func (m *Memory) Get(name string) (Metric, error) {
	metric, ok := m.db[name]
	if !ok {
		return Metric{}, fmt.Errorf("no value found")
	}

	return metric, nil
}

func (m *Memory) GetAll() map[string]float64 {
	res := map[string]float64{}

	for name, metric := range m.db {
		switch metric.MType {
		case Counter.String():
			res[name] = float64(*metric.Delta)
		case Gauge.String():
			res[name] = *metric.Value
		}
	}

	return res
}
