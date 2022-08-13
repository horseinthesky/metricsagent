package server

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

const dashboardTemplate = "internal/server/templates/dashboard.html"

var stash = storage.NewMemoryStorage()

func handleSaveMetric(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")
	valueString := chi.URLParam(r, "value")

	err := stash.Set(metricName, valueString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(http.StatusText(http.StatusBadRequest)))
		return
	}

	w.Write([]byte("Received a POST request\n"))
}

func handleLoadMetric(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "metricName")

	value, err := stash.Get(metricName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	w.Write([]byte(value))
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(http.StatusText(http.StatusNotFound)))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
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
