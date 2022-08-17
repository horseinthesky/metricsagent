package storage

type Storage interface {
	Set(name, value string) error
	Get(name string) (string, error)
	GetAll() map[string]float64
}
