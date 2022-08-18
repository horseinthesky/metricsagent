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
