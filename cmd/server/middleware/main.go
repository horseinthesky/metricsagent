package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func DropUnsupportedType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")

		if metricType != "gauge" && metricType != "counter" {
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
			return
		}

		next.ServeHTTP(w, r)
	})
}
