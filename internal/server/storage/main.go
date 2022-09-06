package storage

import "context"

type MetricType int

const (
	Gauge MetricType = iota
	Counter
)

func (mt MetricType) String() string {
	return [...]string{
		"gauge",
		"counter",
	}[mt]
}

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type Storage interface {
	Init(context.Context) error
	Check(context.Context) error
	Set(Metric) error
	SetBulk([]Metric) error
	Get(string) (Metric, error)
	GetAll() (map[string]Metric, error)
	Close()
}

func UnsupportedType(mtype string) bool {
	if mtype != Gauge.String() && mtype != Counter.String() {
		return true
	}

	return false
}
