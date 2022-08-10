package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/horseinthesky/metricsagent/cmd/server/handlers"
)

const (
	listenOn = ":8080"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/update/{metricType}/{metricName}/{value}", handlers.SaveMetric)
	r.Post("/update/*", handlers.Null)

	r.Get("/value/{metricType}/{metricName}", handlers.LoadMetric)
	r.Get("/value/*", handlers.Null)

	r.Get("/", handlers.AllMetricHandler)

	log.Fatal(http.ListenAndServe(listenOn, r))
}
