package storage

import (
	"fmt"
	"strconv"
	"sync"
)

// var GaugeStorage = map[string]float64{}
// var CounterStorage = map[string]int64{}

type Memory struct {
	sync.Map
}

func (m *Memory) Set(name, value string) error {
	counter, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		oldValue, loaded := m.Load(name)
		if loaded {
			oldCounter, _ := oldValue.(int64)
			m.Store(name, counter + oldCounter)
			return nil

		}

		m.Store(name, counter)
		return nil
	}

	gauge, err := strconv.ParseFloat(value, 64)
	if err == nil {
		m.Store(name, gauge)
		return nil
	}

	return fmt.Errorf("failed to store")
}

func (m *Memory) Get(name string) (string, error) {
	value, loaded := m.Load(name)
	if !loaded {
		return "", fmt.Errorf("no value found")
	}

	counter, ok := value.(int64)
	if ok {
		return fmt.Sprint(counter), nil
	}

	gauge, ok := value.(float64)
	if ok {
		return fmt.Sprint(gauge), nil
	}

	return "", fmt.Errorf("unknown type")
}

func (m *Memory) GetAll() map[string]float64 {
	res := map[string]float64{}
	m.Range(func(key, value interface{}) bool {
		gauge, ok := value.(float64)
		if ok {
			res[key.(string)] = gauge
			return true
		}

		counter, ok := value.(int64)
		if ok {
			res[key.(string)] = float64(counter)
		}

		return true
	})

	return res
}
