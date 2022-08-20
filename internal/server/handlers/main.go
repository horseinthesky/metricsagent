package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

const dashboardTemplate = "internal/server/templates/dashboard.html"

var stash storage.Storage = storage.NewMemoryStorage()

func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(http.StatusText(http.StatusNotFound)))
}

func HandleDashboard(db storage.Storage) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		floatedMetrics := map[string]float64{}

		for name, metric := range db.GetAll() {
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
