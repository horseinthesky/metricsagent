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

type Storage interface {
	Set(name, value string) error
	Get(name string) (any, error)
	GetAll() map[string]float64
}

func UnsupportedType(mtype string) bool {
	if mtype != Gauge.String() && mtype != Counter.String() {
		return true
	}

	return false
}
