package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

func dropUnsupportedTextType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")

		if metricType != storage.Gauge.String() && metricType != storage.Counter.String() {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
			return
		}

		next.ServeHTTP(w, r)
	})
}
