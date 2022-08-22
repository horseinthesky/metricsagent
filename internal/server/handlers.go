package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

const dashboardTemplate = "internal/server/templates/dashboard.html"

func (s *Server) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(http.StatusText(http.StatusNotFound)))
}

func (s *Server) HandleDashboard() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		floatedMetrics := map[string]float64{}

		for name, metric := range s.storage.GetAll() {
			switch metric.MType {
			case storage.Counter.String():
				floatedMetrics[name] = float64(*metric.Delta)
			case storage.Gauge.String():
				floatedMetrics[name] = *metric.Value
			}
		}

		htmlPage, err := os.ReadFile(dashboardTemplate)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		w.Header().Set("Content-Type", "text/html")

		tmpl := template.Must(template.New("").Parse(string(htmlPage)))
		tmpl.Execute(w, floatedMetrics)
	})
}

func (s *Server) HandleSaveTextMetric() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		valueString := chi.URLParam(r, "value")

		var metric *storage.Metric

		switch metricType {
		case storage.Counter.String():
			value, err := strconv.ParseInt(valueString, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(http.StatusText(http.StatusBadRequest)))
				return
			}
			metric = &storage.Metric{
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
			metric = &storage.Metric{
				ID:    metricName,
				MType: metricType,
				Value: &value,
			}
		}

		err := s.storage.Set(metric)
		if err != nil {
			http.Error(w, "failed to save metric", http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Success: metric stored\n"))
	})
}

func (s *Server) HandleLoadTextMetric() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricName := chi.URLParam(r, "metricName")

		metric, err := s.storage.Get(metricName)
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

func (s *Server) HandleSaveJSONMetric() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		metric := &storage.Metric{}
		err := json.NewDecoder(r.Body).Decode(metric)
		if err != nil {
			http.Error(w, `{"error": "bad or no payload"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if storage.UnsupportedType(metric.MType) {
			http.Error(w, `{"error": "unsupported metric type"}`, http.StatusNotImplemented)
			return
		}

		err = s.saveMetric(metric)
		if err != nil {
			http.Error(w, `{"error": "unsupported metric type"}`, http.StatusBadRequest)
			return
		}

		w.Write([]byte(`{"result": "metric saved"}`))
	})
}

func (s *Server) HandleLoadJSONMetric() http.HandlerFunc {
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

		metric, err := s.storage.Get(metricRequest.ID)
		if err != nil {
			http.Error(w, `{"result": "unknown metric id"}`, http.StatusNotFound)
			return
		}

		res, err := json.Marshal(metric)
		if err != nil {
			http.Error(w, `{"error": "faied to marshal metric"}`, http.StatusInternalServerError)
			return
		}

		w.Write([]byte(res))
	})
}
