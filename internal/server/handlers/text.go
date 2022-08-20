package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

func HandleSaveTextMetric(w http.ResponseWriter, r *http.Request) {
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

	err := stash.Set(metric)
	if err != nil {
		http.Error(w, "failed to save metric", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Success: metric stored\n"))
}

func HandleLoadTextMetric(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")

	metric, err := stash.Get(metricName)
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
}
