package storage

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
}

type Storage interface {
	Set(metric *Metric) error
	Get(name string) (Metric, error)
	GetAll() map[string]float64
}

func UnsupportedType(mtype string) bool {
	if mtype != Gauge.String() && mtype != Counter.String() {
		return true
	}

	return false
}
