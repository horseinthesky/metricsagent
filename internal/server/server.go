package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

type Config struct {
	Address       string        `env:"ADDRESS"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
}

type Server struct {
	*chi.Mux
	config   *Config
	storage  storage.Storage
	backuper *Backuper
}

func New(config *Config) *Server {
	// Router
	r := chi.NewRouter()

	// Srorage
	memoryDB := storage.NewMemoryStorage()

	// Backuper
	backuper := NewBackuper(config.StoreFile)

	// Server
	server := &Server{r, config, memoryDB, backuper}
	server.setupRouter()

	return server
}

func (s *Server) setupRouter() {
	// Middleware
	s.Use(logRequest)
	s.Use(handleGzip)
	s.Use(middleware.RequestID)
	s.Use(middleware.RealIP)
	s.Use(middleware.Logger)
	s.Use(middleware.Recoverer)

	// Handlers
	s.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(dropUnsupportedTextType)
			r.Post("/{metricName}/{value}", s.HandleSaveTextMetric())
		})
		r.Post("/", s.HandleSaveJSONMetric())
		r.Post("/*", s.HandleNotFound)
	})

	s.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(dropUnsupportedTextType)
			r.Get("/{metricName}", s.HandleLoadTextMetric())
		})
		r.Post("/", s.HandleLoadJSONMetric())
		r.Get("/*", s.HandleNotFound)
	})

	s.Get("/", s.HandleDashboard())
}

func (s *Server) Start(rootCtx context.Context) {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	// Restore metrics from backup
	if s.config.Restore {
		s.restore()
	}

	// Backup metrics periodically
	if s.config.StoreFile != "" && s.config.StoreInterval > time.Duration(0)*time.Second {
		go s.startPeriodicMetricsDump(ctx)
	}

	log.Println(fmt.Errorf("server crashed due to %w", http.ListenAndServe(s.config.Address, s)))
}

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

func (s *Server) saveMetric(metric *storage.Metric) error {
	err := s.storage.Set(metric)

	if s.config.StoreFile != "" && s.config.StoreInterval == time.Duration(0) {
		s.dump()
	}

	return err
}
