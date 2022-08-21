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

	"github.com/horseinthesky/metricsagent/internal/server/handlers"
	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

type Config struct {
	Address       string `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval int    `env:"STORE_INTERVAL" envDefault:"300"`
	StoreFile     string `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool   `env:"RESTORE" envDefault:"true"`
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

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Handlers
	r.Route("/update", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(dropUnsupportedTextType)
			r.Post("/{metricName}/{value}", handlers.HandleSaveTextMetric(server.storage))
		})
		r.Post("/", handlers.HandleSaveJSONMetric(server.storage))
		r.Post("/*", handlers.HandleNotFound)
	})

	r.Route("/value", func(r chi.Router) {
		r.Route("/{metricType}", func(r chi.Router) {
			r.Use(dropUnsupportedTextType)
			r.Get("/{metricName}", handlers.HandleLoadTextMetric(server.storage))
		})
		r.Post("/", handlers.HandleLoadJSONMetric(server.storage))
		r.Get("/*", handlers.HandleNotFound)
	})

	r.Get("/", handlers.HandleDashboard(server.storage))

	return server
}

func (s *Server) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	// Restore metrics from backup
	if s.config.Restore {
		s.restore()
	}

	// Backup metrics periodically
	if s.config.StoreFile != "" && s.config.StoreInterval > 0 {
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
