package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/horseinthesky/metricsagent/internal/server/handlers"
)

const (
	listenOn = ":8080"
)

type Server struct {
	*chi.Mux
}

func New() *Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(dropUnsupportedTextType)
			r.Post("/{metricName}/{value}", handlers.HandleSaveMetric)
		})
		r.Post("/", handlers.HandleSaveJSONMetric)
		r.Post("/*", handlers.HandleNotFound)
	})

	r.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(dropUnsupportedTextType)
			r.Get("/{metricName}", handlers.HandleLoadMetric)
		})
		r.Post("/", handlers.HandleLoadJSONMetric)
		r.Get("/*", handlers.HandleNotFound)
	})

	r.Get("/", handlers.HandleDashboard)

	return &Server{r}
}
