package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

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

func dropUnsupportedType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")

		if metricType != Gauge.String() && metricType != Counter.String() {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
			return
		}

		next.ServeHTTP(w, r)
	})
}
