package server

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request. Headers are:")
		for header, values := range r.Header {
			log.Println(header, values)

		}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Body: failed to read", )
			r.Body.Close()
			next.ServeHTTP(w, r)
		}

		log.Println("Body:", string(bodyBytes))

		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(w, r)
	})
}

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
