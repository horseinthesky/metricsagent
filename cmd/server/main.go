package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/horseinthesky/metricsagent/cmd/server/handlers"
	cmiddleware "github.com/horseinthesky/metricsagent/cmd/server/middleware"
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

	r.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(cmiddleware.DropUnsupportedType)
			r.Post("/{metricName}/{value}", handlers.HandleSaveMetric)
		})
		r.Post("/*", handlers.HandleNotFound)
	})

	r.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(cmiddleware.DropUnsupportedType)
			r.Get("/{metricName}", handlers.HandleLoadMetric)
		})
		r.Get("/*", handlers.HandleNotFound)
	})

	r.Get("/", handlers.HandleDashboard)

	log.Fatal(http.ListenAndServe(listenOn, r))
}
