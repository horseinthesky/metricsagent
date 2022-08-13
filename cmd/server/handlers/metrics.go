package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/horseinthesky/metricsagent/cmd/server/storage"
)

const dashboardTemplate = "cmd/server/templates/dashboard.html"

var stash = &storage.Memory{}

func unsupportedType(mtype string) bool {
	if mtype != "gauge" && mtype != "counter" {
		return true
	}

	return false
}

func HandleSaveMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	valueString := chi.URLParam(r, "value")

	if unsupportedType(metricType) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(http.StatusText(http.StatusNotImplemented)))
		return
	}

	err := stash.Set(metricName, valueString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	w.Write([]byte("Received a POST request\n"))
}

func HandleLoadMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	if unsupportedType(metricType) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	value, err := stash.Get(metricName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	w.Write([]byte(value))
}

func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(http.StatusText(http.StatusNotFound)))
}

func HandleDashboard(w http.ResponseWriter, r *http.Request) {
	allMetrics := stash.GetAll()

	htmlPage, err := os.ReadFile(dashboardTemplate)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	w.Header().Set("Content-Type", "text/html")

	tmpl := template.Must(template.New("").Parse(string(htmlPage)))
	tmpl.Execute(w, allMetrics)
}
