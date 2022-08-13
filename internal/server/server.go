package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
			r.Use(dropUnsupportedType)
			r.Post("/{metricName}/{value}", handleSaveMetric)
		})
		r.Post("/*", handleNotFound)
	})

	r.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(dropUnsupportedType)
			r.Get("/{metricName}", handleLoadMetric)
		})
		r.Get("/*", handleNotFound)
	})

	r.Get("/", handleDashboard)

	return &Server{r}
}
