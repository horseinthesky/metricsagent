package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/horseinthesky/metricsagent/cmd/server/handlers"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/update/{metricType}/{metricName}/{value}", handlers.SaveHandler)
	r.Get("/value/{metricType}/{metricName}", handlers.LoadHandler)
	r.Get("/", handlers.AllMetricHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
