package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

type Metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func extractMetic(r *http.Request) (*Metric, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	metric := &Metric{}

	err = json.Unmarshal(body, metric)
	if err != nil {
		return nil, err
	}

	return metric, nil
}

func HandleSaveJSONMetric(w http.ResponseWriter, r *http.Request) {
	metric, err := extractMetic(r)
	if err != nil {
		http.Error(w, "failed to unmarshal payload", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	if storage.UnsupportedType(metric.MType) {
		http.Error(w, "Unknown type", http.StatusNotImplemented)
		return
	}

	var value string
	if metric.MType == storage.Counter.String() {
		value = fmt.Sprint(*metric.Delta)
	} else {
		value = fmt.Sprint(*metric.Value)
	}

	err = stash.Set(metric.ID, value)
	if err != nil {
		http.Error(w, "Invalid value", http.StatusBadRequest)
		return
	}

	w.Write([]byte("Received a POST request\n"))
}

func HandleLoadJSONMetric(w http.ResponseWriter, r *http.Request) {
	metricRequest, err := extractMetic(r)
	if err != nil {
		http.Error(w, "failed to unmarshal payload", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	if storage.UnsupportedType(metricRequest.MType) {
		http.Error(w, "Unknown type", http.StatusNotImplemented)
		return
	}

	value, err := stash.Get(metricRequest.ID)
	if err != nil {
		http.Error(w, "Unknown name", http.StatusNotFound)
		return
	}

	metric := &Metric{
		ID: metricRequest.ID,
		MType: metricRequest.MType,
	}
	if metricRequest.MType == storage.Counter.String() {
		v, _ := value.(int64)
		metric.Delta = &v
	} else if metricRequest.MType == storage.Gauge.String() {
		v, _ := value.(float64)
		metric.Value = &v
	}

	res, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, "Internal error during JSON marshal", http.StatusInternalServerError)
	}

	w.Write([]byte(res))
}
