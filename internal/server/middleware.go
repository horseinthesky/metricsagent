package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

// gzipWriter provides compression interface for middleware.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write writes compressed bytes to HTTP writer.
func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// handleGzip provides gzip compression.
func handleGzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

// logRequest logs some HTTP request data.
// Stores:
//  - method
//  - client address
//  - headers
//  - URL path
//  - body
//  - headers
func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Got %s request from %s for %s", r.Method, r.RemoteAddr, r.URL.Path)

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Body: failed to read")
			r.Body.Close()
			next.ServeHTTP(w, r)
		}
		defer r.Body.Close()

		log.Print("Body:", string(bodyBytes))
		log.Print("Headers:")
		for header, values := range r.Header {
			log.Print(header, values)

		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next.ServeHTTP(w, r)
	})
}

// dropUnsupportedTextType provides early request drop
// if metric type is not supported.
// Only used with handlers which get metrics data from URL params.
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
