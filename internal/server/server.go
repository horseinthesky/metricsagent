// Package server describes metrics server  internals.
//
// It consists of the following parts:
//  - server.go - server struct and its lifecycle methods
//  - config.go - server configuration options
//  - backup.go - server periodic backup methods
//  - secure.go - server metrics hash protection
//  - middleware.go - server middleware
//  - handlers.go - server HTTP router endpoints buciness logic
package server

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

// Server main struct.
type Server struct {
	*chi.Mux
	config    Config
	db        storage.Storage
	backuper  *Backuper
	workGroup sync.WaitGroup
}

// Server constructor.
// Sets things up.
func NewServer(config Config) *Server {
	r := chi.NewRouter()

	var db storage.Storage
	if config.DatabaseDSN != "" {
		db = storage.NewDBStorage(config.DatabaseDSN)
	} else {
		db = storage.NewMemoryStorage()
	}

	backuper := NewBackuper(config.StoreFile)

	server := &Server{r, config, db, backuper, sync.WaitGroup{}}
	server.setupRouter()

	return server
}

// setupRouter builds Server's HTTP router.
// Assembles middleware and handlers.
func (s *Server) setupRouter() {
	s.Use(logRequest)
	s.Use(handleGzip)
	s.Use(middleware.RequestID)
	s.Use(middleware.RealIP)
	s.Use(middleware.Logger)
	s.Use(middleware.Recoverer)

	s.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(dropUnsupportedTextType)
			r.Post("/{metricName}/{value}", s.handleSaveTextMetric())
		})
		r.Post("/", s.handleSaveJSONMetric())
		r.Post("/*", s.handleNotFound)
	})
	s.Post("/updates/", s.handleSaveJSONMetrics())

	s.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(dropUnsupportedTextType)
			r.Get("/{metricName}", s.handleLoadTextMetric())
		})
		r.Post("/", s.handleLoadJSONMetric())
		r.Get("/*", s.handleNotFound)
	})

	s.Get("/", s.handleDashboard())
	s.Get("/ping", s.handlePingDB())
}

// Run is a Server entry point.
// It starts DB, HTTP router and periodic metrics backup.
func (s *Server) Run(ctx context.Context) {
	if s.config.DatabaseDSN == "" {
		// Restore metrics from backup
		if s.config.Restore {
			s.restore()
		}

		// Backup metrics periodically
		if s.config.StoreFile != "" && s.config.StoreInterval > time.Duration(0)*time.Second {
			s.workGroup.Add(1)
			go func() {
				defer s.workGroup.Done()
				s.startPeriodicMetricsDump(ctx)
			}()
		}
	}

	err := s.db.Init(ctx)
	if err != nil {
		log.Fatalf("failed to init db: %s", err)
	}

	log.Fatalf("server crashed due to %s", http.ListenAndServe(s.config.Address, s))
}

// Stop is a Server graceful shutdown method.
// Ensures everything is stopped as expected.
func (s *Server) Stop() {
	log.Println("shutting down...")

	s.db.Close()
	log.Println("connection to database closed")

	s.workGroup.Wait()
	log.Println("successfully shut down")
}

// startPeriodicMetricsDump handles Server periodic metrics backup to file.
// Only used with memory DB to provide persistent metrics storage
// between Server restart.
func (s *Server) startPeriodicMetricsDump(ctx context.Context) {
	log.Println("pediodic metrics backup started")

	ticker := time.NewTicker(s.config.StoreInterval)

	for {
		select {
		case <-ticker.C:
			s.dump()
		case <-ctx.Done():
			log.Println("metrics backup canceled")
			return
		}
	}
}

// saveMetric handles synchronous metric backup.
// Only used by handleSaveJSONMetric handler when
//  - in-memory storage is in use
//  - no StoreInterval provided
func (s *Server) saveMetric(metric storage.Metric) error {
	err := s.db.Set(metric)

	if s.config.DatabaseDSN == "" {
		if s.config.StoreFile != "" && s.config.StoreInterval == time.Duration(0) {
			s.dump()
		}
	}

	return err
}

// saveMetricBulk handles synchronous bulk metric backup.
// Only used by handleSaveJSONMetrics handler when
//  - in-memory storage is in use
//  - no StoreInterval provided
func (s *Server) saveMetricsBulk(metrics []storage.Metric) error {
	err := s.db.SetBulk(metrics)

	if s.config.DatabaseDSN == "" {
		if s.config.StoreFile != "" && s.config.StoreInterval == time.Duration(0) {
			s.dump()
		}
	}

	return err
}
