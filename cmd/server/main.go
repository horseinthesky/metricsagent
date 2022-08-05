package main

import (
	"net/http"

	"github.com/horseinthesky/metricsagent/cmd/server/handlers"
)

func main() {
	http.HandleFunc("/update/", handlers.MetricsHandler)
	http.ListenAndServe(":8080", nil)
}
