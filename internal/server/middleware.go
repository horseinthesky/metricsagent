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
		log.Printf("Got %s request from %s for %s", r.Method, r.RemoteAddr, r.URL.Path)

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Body: failed to read", )
			r.Body.Close()
			next.ServeHTTP(w, r)
		}
		defer r.Body.Close()

		log.Print("Body:", string(bodyBytes))
		log.Print("Headers:")
		for header, values := range r.Header {
			log.Print(header, values)

		}

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
