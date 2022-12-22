package server

import (
	"crypto/hmac"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

//go:embed templates/dashboard.html
var dashboardTemplate string

// handleDashboard handles metrics dashboard rendering.
func (s *Server) handleDashboard() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		floatedMetrics := map[string]float64{}

		allMetrics, err := s.db.GetAll(r.Context())
		if err != nil {
			log.Printf("failed to get stored metrics: %s", err)
			return
		}

		for name, metric := range allMetrics {
			switch metric.MType {
			case storage.Counter.String():
				floatedMetrics[name] = float64(*metric.Delta)
			case storage.Gauge.String():
				floatedMetrics[name] = *metric.Value
			}
		}

		w.Header().Set("Content-Type", "text/html")

		tmpl, err := template.New("").Parse(dashboardTemplate)
		if err != nil {
			log.Printf("failed to parse a template: %s", err)
			return
		}

		err = tmpl.Execute(w, floatedMetrics)
		if err != nil {
			log.Printf("failed to render a template: %s", err)
		}
	})
}

// handleSaveTextMetric provides single metric receiver.
// Metric type, name and value are obtained from URL params.
func (s *Server) handleSaveTextMetric() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		valueString := chi.URLParam(r, "value")

		var metric storage.Metric

		switch metricType {
		case storage.Counter.String():
			value, err := strconv.ParseInt(valueString, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(http.StatusText(http.StatusBadRequest)))
				return
			}
			metric = storage.Metric{
				ID:    metricName,
				MType: metricType,
				Delta: &value,
			}
		case storage.Gauge.String():
			value, err := strconv.ParseFloat(string(valueString), 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(http.StatusText(http.StatusBadRequest)))
				return
			}
			metric = storage.Metric{
				ID:    metricName,
				MType: metricType,
				Value: &value,
			}
		}

		err := s.db.Set(metric)
		if err != nil {
			http.Error(w, "failed to save metric", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Success: metric stored\n"))
	})
}

// handleLoadTextMetric provides single metric loader.
// Metric name is obtained from URL param.
func (s *Server) handleLoadTextMetric() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")

		metric, err := s.db.Get(r.Context(), metricName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(http.StatusText(http.StatusNotFound)))
			return
		}

		var value string

		switch metric.MType {
		case storage.Counter.String():
			value = fmt.Sprint(*metric.Delta)
		case storage.Gauge.String():
			value = fmt.Sprint(*metric.Value)
		}

		w.Write([]byte(value))
	})
}

// handleSaveJSONMetrics provides multiple metrics receiver.
// Metrics type, name and value are obtained from JSON payload.
func (s *Server) handleSaveJSONMetrics() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		metrics := []storage.Metric{}
		err := json.NewDecoder(r.Body).Decode(&metrics)
		if err != nil {
			http.Error(w, `{"error": "bad or no payload"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		for _, metric := range metrics {
			if storage.UnsupportedType(metric.MType) {
				http.Error(w, `{"error": "unsupported metric type"}`, http.StatusNotImplemented)
				return
			}

			if s.config.Key != "" {
				localHash := s.generateHash(metric)
				remoteHash, err := hex.DecodeString(metric.Hash)
				if err != nil {
					http.Error(w, `{"error": "failed to decode hash"}`, http.StatusInternalServerError)
					return
				}

				if !hmac.Equal(localHash, remoteHash) {
					http.Error(w, `{"error": "invalid hash"}`, http.StatusBadRequest)
					return
				}
			}
		}

		err = s.saveMetricsBulk(metrics)
		if err != nil {
			log.Printf("failed to store metric: %s", err)
			http.Error(w, `{"error": "failed to store metric"}`, http.StatusBadRequest)
			return
		}

		w.Write([]byte(`{"result": "metric saved"}`))
	})
}

// handleSaveJSONMetric provides single metric receiver.
// Metric type, name and value are obtained from JSON payload.
func (s *Server) handleSaveJSONMetric() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		metric := storage.Metric{}
		err := json.NewDecoder(r.Body).Decode(&metric)
		if err != nil {
			http.Error(w, `{"error": "bad or no payload"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if storage.UnsupportedType(metric.MType) {
			http.Error(w, `{"error": "unsupported metric type"}`, http.StatusNotImplemented)
			return
		}

		if s.config.Key != "" {
			localHash := s.generateHash(metric)
			remoteHash, err := hex.DecodeString(metric.Hash)

			if err != nil {
				http.Error(w, `{"error": "failed to decode hash"}`, http.StatusInternalServerError)
				return
			}

			if !hmac.Equal(localHash, remoteHash) {
				http.Error(w, `{"error": "invalid hash"}`, http.StatusBadRequest)
				return
			}
		}

		err = s.saveMetric(metric)
		if err != nil {
			log.Printf("failed to store metric: %s", err)
			http.Error(w, `{"error": "failed to store metric"}`, http.StatusBadRequest)
			return
		}

		w.Write([]byte(`{"result": "metric saved"}`))
	})
}

// handleLoadJSONMetric provides metric JSON loader.
// Metric ID is obtained from JSON payload.
func (s *Server) handleLoadJSONMetric() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		metricRequest := &storage.Metric{}
		err := json.NewDecoder(r.Body).Decode(metricRequest)
		if err != nil {
			http.Error(w, `{"error": "bad or no payload"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if storage.UnsupportedType(metricRequest.MType) {
			http.Error(w, `{"error": "unsupported metric type"}`, http.StatusNotImplemented)
			return
		}

		metric, err := s.db.Get(r.Context(), metricRequest.ID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"result": "unknown metric id"}`))
			return
		}

		if s.config.Key != "" {
			metric.Hash = hex.EncodeToString(s.generateHash(metric))
		}

		res, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, `{"error": "faied to marshal metric"}`, http.StatusInternalServerError)
			return
		}

		w.Write([]byte(res))
	})
}

// handlePingDB provides Server's DB healthcheck.
func (s *Server) handlePingDB() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.db.Check(r.Context()); err != nil {
			log.Printf("failed to ping DB: %s", err)
			http.Error(w, "failed to ping DB", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	})
}
