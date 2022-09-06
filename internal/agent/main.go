package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type gauge = float64
type counter = int64

type Config struct {
	Address        string        `env:"ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	Key            string        `env:"KEY"`
}

type Agent struct {
	PollTicker   *time.Ticker
	ReportTicker *time.Ticker
	PollCounter  int64
	key          string
	metrics      *sync.Map
	upstream     string
	client       *http.Client
}

type Metric struct {
	ID    string `json:"id"`              // имя метрики
	MType string `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64 `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *gauge `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string `json:"hash,omitempty"`  // значение хеш-функции
}

func New(cfg *Config) *Agent {
	return &Agent{
		PollTicker:   time.NewTicker(cfg.PollInterval),
		ReportTicker: time.NewTicker(cfg.ReportInterval),
		key:          cfg.Key,
		metrics:      &sync.Map{},
		upstream:     fmt.Sprintf("http://%s", cfg.Address),
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
	}
}

func (a Agent) Run(ctx context.Context) {
	data := &runtime.MemStats{}

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down agent")
			return
		case <-a.ReportTicker.C:
			a.SendMetricsJSONBulk()
		case <-a.PollTicker.C:
			a.PollCounter++

			runtime.ReadMemStats(data)

			a.UpdateMetrics(data)
		}
	}
}

func (a *Agent) UpdateMetrics(data *runtime.MemStats) {
	a.metrics.Store("Alloc", gauge(data.Alloc))
	a.metrics.Store("BuckHashSys", gauge(data.BuckHashSys))
	a.metrics.Store("Frees", gauge(data.Frees))
	a.metrics.Store("GCCPUFraction", gauge(data.GCCPUFraction))
	a.metrics.Store("GCSys", gauge(data.GCSys))
	a.metrics.Store("HeapAlloc", gauge(data.HeapAlloc))
	a.metrics.Store("HeapIdle", gauge(data.HeapIdle))
	a.metrics.Store("HeapInuse", gauge(data.HeapInuse))
	a.metrics.Store("HeapObjects", gauge(data.HeapObjects))
	a.metrics.Store("HeapReleased", gauge(data.HeapReleased))
	a.metrics.Store("HeapSys", gauge(data.HeapSys))
	a.metrics.Store("LastGC", gauge(data.LastGC))
	a.metrics.Store("Lookups", gauge(data.Lookups))
	a.metrics.Store("MCacheInuse", gauge(data.MCacheInuse))
	a.metrics.Store("MCacheSys", gauge(data.MCacheSys))
	a.metrics.Store("MSpanInuse", gauge(data.MSpanInuse))
	a.metrics.Store("MSpanSys", gauge(data.MSpanSys))
	a.metrics.Store("Mallocs", gauge(data.Mallocs))
	a.metrics.Store("NextGC", gauge(data.NextGC))
	a.metrics.Store("NumForcedGC", gauge(data.NumForcedGC))
	a.metrics.Store("NumGC", gauge(data.NumGC))
	a.metrics.Store("OtherSys", gauge(data.OtherSys))
	a.metrics.Store("PauseTotalNs", gauge(data.PauseTotalNs))
	a.metrics.Store("StackInuse", gauge(data.StackInuse))
	a.metrics.Store("StackSys", gauge(data.StackSys))
	a.metrics.Store("Sys", gauge(data.Sys))
	a.metrics.Store("TotalAlloc", gauge(data.TotalAlloc))
	a.metrics.Store("RandomValue", gauge(rand.Float64()))
}
