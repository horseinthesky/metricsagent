package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"

	"github.com/horseinthesky/metricsagent/cmd/server/storage"
)

func unsupportedType(mtype string) bool {
	if mtype != "gauge" && mtype != "counter" {
		return false
	}

	return true
}

func SaveHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	valueString := chi.URLParam(r, "value")

	if ok := unsupportedType(metricType); !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	if metricType == "gauge" {
		value, err := strconv.ParseFloat(valueString, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}
		storage.GaugeStorage[metricName] = value
	}

	if metricType == "counter" {
		value, err := strconv.ParseInt(valueString, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(http.StatusText(http.StatusBadRequest)))
			return
		}

		if oldValue, ok := storage.CounterStorage[metricName]; ok {
			newValue := oldValue + value
			storage.CounterStorage[metricName] = newValue
		} else {
			storage.CounterStorage[metricName] = value
		}
	}

	w.Write([]byte("Received a POST request\n"))
}

func LoadHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	if ok := unsupportedType(metricType); !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	if metricType == "gauge" {
		if value, ok := storage.GaugeStorage[metricName]; ok {
			w.Write([]byte(fmt.Sprint(value)))
			return
		}
	}

	if metricType == "counter" {
		if value, ok := storage.CounterStorage[metricName]; ok {
			w.Write([]byte(fmt.Sprint(value)))
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(http.StatusText(http.StatusNotFound)))
}

func AllMetricHandler(w http.ResponseWriter, r *http.Request) {
	allMetrics := map[string]float64{}
	for k, v := range storage.GaugeStorage {
		allMetrics[k] = v
	}
	for k, v := range storage.CounterStorage {
		allMetrics[k] = float64(v)
	}

	htmlPage, err := os.ReadFile("cmd/server/templates/dashboard.html")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.New("").Parse(string(htmlPage)))
	tmpl.Execute(w, allMetrics)
}
