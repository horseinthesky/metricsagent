package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/horseinthesky/metricsagent/internal/crypto"
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

// trustedSubnet checks if a request came from an IP of stusted subnet.
// Drops a request and returns 403 if not true.
func (s *Server) trustedSubnet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.Config.TrustedSubnet == "" {
			next.ServeHTTP(w, r)
			return
		}

		_, trustedNet, _ := net.ParseCIDR(s.Config.TrustedSubnet)
		requestIP := r.Header.Get("X-Real-IP")

		if !trustedNet.Contains(net.ParseIP(requestIP)) {
			log.Printf("request from %s is forbidden", requestIP)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(http.StatusText(http.StatusForbidden)))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleDecrypt provides RSA decryption.
func (s *Server) handleDecrypt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.CryptoKey == nil {
			next.ServeHTTP(w, r)
			return
		}

		encryptedBody, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("failed to read body")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}
		defer r.Body.Close()

		decryptedBody, err := crypto.DecryptWithPrivateKey(encryptedBody, s.CryptoKey)
		if err != nil {
			log.Println("failed to decrypt body")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))

		next.ServeHTTP(w, r)
	})
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
//   - method
//   - client address
//   - headers
//   - URL path
//   - body
//   - headers
// func logRequest(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Printf("Got %s request from %s for %s", r.Method, r.RemoteAddr, r.URL.Path)
//
// 		bodyBytes, err := io.ReadAll(r.Body)
// 		if err != nil {
// 			log.Println("Body: failed to read")
// 			next.ServeHTTP(w, r)
// 		}
// 		defer r.Body.Close()
//
// 		log.Print("Body:", string(bodyBytes))
// 		log.Print("Headers:")
// 		for header, values := range r.Header {
// 			log.Print(header, values)
//
// 		}
//
// 		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
//
// 		next.ServeHTTP(w, r)
// 	})
// }

// // dropUnsupportedTextType provides early request drop
// // if metric type is not supported.
// // Only used with handlers which get metrics data from URL params.
func dropUnsupportedTextType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")

		if metricType != storage.Gauge.String() && metricType != storage.Counter.String() {
			log.Printf("metric has unsupported type: %s", metricType)
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
			return
		}

		next.ServeHTTP(w, r)
	})
}
