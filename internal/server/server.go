package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

type Config struct {
	Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
}

type Server struct {
	*chi.Mux
	config  Config
	storage storage.Storage
}

func New(config Config) *Server {
	// Router
	r := chi.NewRouter()

	// Srorage
	memoryDB := storage.NewMemoryStorage()

	// Server
	server := &Server{r, config, memoryDB}
	server.setupRouter()

	return server
}

func (s *Server) setupRouter() {
	// Middleware
	s.Use(logRequest)
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

func (s *Server) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	// Restore metrics from backup
	if s.config.Restore {
		s.restore()
	}

	// Backup metrics periodically
	if s.config.StoreFile != "" && s.config.StoreInterval > time.Duration(0) * time.Second {
		go s.startPeriodicMetricsDump(ctx)
	}

	log.Println(fmt.Errorf("server crashed due to %w", http.ListenAndServe(s.config.Address, s)))
	cancel()
}

func (s *Server) startPeriodicMetricsDump(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.config.StoreInterval) * time.Second)
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

func (s *Server) dump() {
	var metrics []storage.Metric

	producer, err := NewProducer(s.config.StoreFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		log.Println(fmt.Errorf("failed to create file producer %s: %w", s.config.StoreFile, err))
		return
	}

	for _, metric := range s.storage.GetAll() {
		metrics = append(metrics, metric)
	}

	if err := producer.WriteMetrics(&metrics); err != nil {
		log.Println(fmt.Errorf("failed to dump metrics to %s: %w", s.config.StoreFile, err))
		return
	}

	log.Printf("successfully dumped all metrics to %s", s.config.StoreFile)
}

func (s *Server) restore() {
	consumer, err := NewConsumer(s.config.StoreFile, os.O_RDONLY|os.O_CREATE)
	if err != nil {
		log.Println(fmt.Errorf("failed to create file consumer %s: %w", s.config.StoreFile, err))
		return
	}

	metrics, err := consumer.ReadMetrics()
	if err != nil {
		log.Println(fmt.Errorf("failed to restore metrics from %s: %w", s.config.StoreFile, err))
		return
	}

	for _, metric := range metrics {
		err := s.storage.Set(&metric)
		if err != nil {
			log.Println(fmt.Errorf("failed to restore metric %s: %w", metric.ID, err))
		}
	}

	log.Printf("successfully restored all metrics from %s", s.config.StoreFile)
}
