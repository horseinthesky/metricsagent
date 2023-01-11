// Package server describes metrics server  internals.
//
// It consists of the following parts:
//   - server.go - server struct and its lifecycle methods
//   - config.go - server configuration options
//   - backup.go - server periodic backup methods
//   - secure.go - server metrics hash protection
//   - middleware.go - server middleware
//   - handlers.go - server HTTP router endpoints buciness logic
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/horseinthesky/metricsagent/internal/server"
)

// Server main struct.
type Server struct {
	*server.GenericServer
	*chi.Mux
}

// Server constructor.
// Sets things up.
func NewServer(cfg server.Config) (*Server, error) {
	genericServer, err := server.NewGenericServer(cfg)
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()

	server := &Server{genericServer, r}
	server.setupRouter()

	return server, nil
}

// setupRouter builds Server's HTTP router.
// Assembles middleware and handlers.
func (s *Server) setupRouter() {
	s.Use(s.trustedSubnet)
	s.Use(handleGzip)
	// s.Use(logRequest)
	s.Use(s.handleDecrypt)
	s.Use(middleware.RequestID)
	s.Use(middleware.RealIP)
	s.Use(middleware.Logger)
	s.Use(middleware.Recoverer)

	s.Route("/update", func(r chi.Router) {
		s.Route("/{metricType}", func(r chi.Router) {
			s.Use(dropUnsupportedTextType)
			s.Post("/{metricName}/{value}", s.handleSaveTextMetric())
		})
		s.Post("/", s.handleSaveJSONMetric())
	})
	s.Post("/updates/", s.handleSaveJSONMetrics())

	s.Route("/value", func(r chi.Router) {
		s.Route("/{metricType}", func(r chi.Router) {
			s.Use(dropUnsupportedTextType)
			s.Get("/{metricName}", s.handleLoadTextMetric())
		})
		s.Post("/", s.handleLoadJSONMetric())
	})

	s.Get("/", s.handleDashboard())
	s.Get("/ping", s.handlePingDB())
}

// Run is a Server entry point.
// It starts DB, HTTP router and periodic metrics backup.
func (s *Server) Run(ctx context.Context) {
	s.Bootstrap(ctx)

	srv := http.Server{
		Addr:    s.Config.Address,
		Handler: s,
	}

	s.WorkGroup.Add(1)
	go func() {
		defer s.WorkGroup.Done()

		runMsg := fmt.Sprintf("Running HTTP server, listening on %s", s.Config.Address)
		if s.Config.TrustedSubnet != "" {
			addon := fmt.Sprintf(", trusted subnet: %s", s.Config.TrustedSubnet)
			runMsg += addon
		}
		log.Println(runMsg)

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("server crashed: %s", err)
		}

		log.Printf("finished to serve HTTP requests")
	}()

	<-ctx.Done()
	srv.Shutdown(ctx)
}

// Stop is a Server graceful shutdown method.
// Ensures everything is stopped as expected.
func (s *Server) Stop() {
	log.Println("shutting down...")

	s.DB.Close()
	log.Println("connection to database closed")

	s.WorkGroup.Wait()
	log.Println("successfully shut down")
}
