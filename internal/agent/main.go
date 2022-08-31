package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type gauge float64

type Config struct {
	Address        string        `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
}

type Agent struct {
	PollTicker   *time.Ticker
	ReportTicker *time.Ticker
	PollCounter  int64
	metrics      *sync.Map
	upstream     string
	client       *http.Client
}

type Metric struct {
	ID    string `json:"id"`              // имя метрики
	MType string `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64 `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *gauge `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func New(cfg *Config) *Agent {
	return &Agent{
		PollTicker:   time.NewTicker(cfg.PollInterval),
		ReportTicker: time.NewTicker(cfg.ReportInterval),
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
		case <-a.ReportTicker.C:
			a.SendMetricsJSON()
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

func (a *Agent) SendMetricsJSON() {
	a.metrics.Range(func(metricName, value interface{}) bool {
		m, _ := metricName.(string)
		v, _ := value.(gauge)

		a.sendPostJSON(
			&Metric{
				ID:    m,
				MType: "gauge",
				Value: &v,
			},
		)
		return true
	})

	// Send poll count
	a.sendPostJSON(
		&Metric{
			ID:    "PollCount",
			MType: "counter",
			Delta: &a.PollCounter,
		},
	)
}

func (a *Agent) SendMetricsPlain() {
	// Send metrics
	a.metrics.Range(func(metricName, value interface{}) bool {
		endpoint := fmt.Sprintf("%s/update/%s/%s/%v", a.upstream, "gauge", metricName, value)
		a.sendPostPlain(endpoint)
		return true
	})

	// Send poll count
	endpoint := fmt.Sprintf("%s/update/%s/%s/%v", a.upstream, "counter", "PollCount", a.PollCounter)
	a.sendPostPlain(endpoint)
}

func (a *Agent) sendPostPlain(url string) {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Println(fmt.Errorf("failed to build a request: %w", err))
		return
	}
	request.Header.Add("Content-Type", "text/plain")

	response, err := a.client.Do(request)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to make a request: %w", err))
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(fmt.Errorf("failed to read response body: %w", err))
		return
	}
	defer response.Body.Close()

	log.Printf("Code: %v: %s", response.Status, string(body))
}

func (a *Agent) sendPostJSON(metric *Metric) {
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(metric)

	endpoint := fmt.Sprintf("%s/update/", a.upstream)
	request, err := http.NewRequest(http.MethodPost, endpoint, payloadBuf)
	if err != nil {
		log.Println(fmt.Errorf("failed to build a request: %w", err))
		return
	}
	request.Header.Add("Content-Type", "application/json")

	response, err := a.client.Do(request)
	if err != nil {
		log.Println(fmt.Errorf("failed to make a request: %w", err))
		return
	}
	defer response.Body.Close()

	log.Printf("Code: %v", response.Status)
}
