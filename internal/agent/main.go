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

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	CPUPollTime = 10 * time.Second
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
	data         *runtime.MemStats
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
		data:         &runtime.MemStats{},
		key:          cfg.Key,
		metrics:      &sync.Map{},
		upstream:     fmt.Sprintf("http://%s", cfg.Address),
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
	}
}

func (a Agent) Run(ctx context.Context) {
	go a.collectRuntimeMetrics(ctx)
	go a.collectPSUtilMetrics(ctx)
	go a.SendMetricsJSONBulk(ctx)

	<-ctx.Done()
	log.Println("shutting down agent")
}

func (a *Agent) collectRuntimeMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("finished collecting runtime data")
			return
		case <-a.PollTicker.C:
			a.updateRuntimeMetrics()
		}
	}
}

func (a *Agent) updateRuntimeMetrics() {
	a.PollCounter++

	runtime.ReadMemStats(a.data)

	a.metrics.Store("Alloc", gauge(a.data.Alloc))
	a.metrics.Store("BuckHashSys", gauge(a.data.BuckHashSys))
	a.metrics.Store("Frees", gauge(a.data.Frees))
	a.metrics.Store("GCCPUFraction", gauge(a.data.GCCPUFraction))
	a.metrics.Store("GCSys", gauge(a.data.GCSys))
	a.metrics.Store("HeapAlloc", gauge(a.data.HeapAlloc))
	a.metrics.Store("HeapIdle", gauge(a.data.HeapIdle))
	a.metrics.Store("HeapInuse", gauge(a.data.HeapInuse))
	a.metrics.Store("HeapObjects", gauge(a.data.HeapObjects))
	a.metrics.Store("HeapReleased", gauge(a.data.HeapReleased))
	a.metrics.Store("HeapSys", gauge(a.data.HeapSys))
	a.metrics.Store("LastGC", gauge(a.data.LastGC))
	a.metrics.Store("Lookups", gauge(a.data.Lookups))
	a.metrics.Store("MCacheInuse", gauge(a.data.MCacheInuse))
	a.metrics.Store("MCacheSys", gauge(a.data.MCacheSys))
	a.metrics.Store("MSpanInuse", gauge(a.data.MSpanInuse))
	a.metrics.Store("MSpanSys", gauge(a.data.MSpanSys))
	a.metrics.Store("Mallocs", gauge(a.data.Mallocs))
	a.metrics.Store("NextGC", gauge(a.data.NextGC))
	a.metrics.Store("NumForcedGC", gauge(a.data.NumForcedGC))
	a.metrics.Store("NumGC", gauge(a.data.NumGC))
	a.metrics.Store("OtherSys", gauge(a.data.OtherSys))
	a.metrics.Store("PauseTotalNs", gauge(a.data.PauseTotalNs))
	a.metrics.Store("StackInuse", gauge(a.data.StackInuse))
	a.metrics.Store("StackSys", gauge(a.data.StackSys))
	a.metrics.Store("Sys", gauge(a.data.Sys))
	a.metrics.Store("TotalAlloc", gauge(a.data.TotalAlloc))
	a.metrics.Store("RandomValue", gauge(rand.Float64()))

	log.Println("successfully collected runtime metrics")
}

func (a *Agent) collectPSUtilMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("finished collecting psutil data")
			return
		case <-a.PollTicker.C:
			memory, _ := mem.VirtualMemory()
			a.metrics.Store("TotalMemory", gauge(memory.Total))
			a.metrics.Store("FreeMemory", gauge(memory.Free))

			cpusUtilization, _ := cpu.Percent(0, true)
			for i, c := range cpusUtilization {
				a.metrics.Store(fmt.Sprintf("CPUutilization%d", i), gauge(c))
			}

			log.Println("successfully collected psutil metrics")
		}
	}
}
