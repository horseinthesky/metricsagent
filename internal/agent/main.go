package agent

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const (
	baseURL string = "http://localhost:8080"
)

type gauge float64

type Agent struct {
	PollTicker   *time.Ticker
	ReportTicker *time.Ticker
	Count        int
	metrics      *sync.Map
	upstream     string
	client       *http.Client
}

func New(poll, report int, url string) *Agent {
	if url == "" {
		url = baseURL
	}

	agent := &Agent{
		PollTicker:   time.NewTicker(time.Duration(poll) * time.Second),
		ReportTicker: time.NewTicker(time.Duration(report) * time.Second),
		metrics:      &sync.Map{},
		upstream:     url,
		client: &http.Client{
			Timeout: 1 * time.Second,
		},
	}

	return agent
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
	a.metrics.Store("Rand", gauge(rand.Float64()))
}

func (a *Agent) SendMetrics() {
	// Send metrics
	a.metrics.Range(func(metricName, value interface{}) bool {
		endpoint := fmt.Sprintf("%s/update/%s/%s/%v", a.upstream, "gauge", metricName, value)
		a.sendPostRequest(endpoint)
		return true
	})

	// Send poll count
	endpoint := fmt.Sprintf("%s/update/%s/%s/%v", a.upstream, "counter", "pollNum", a.Count)
	a.sendPostRequest(endpoint)
}

func (a *Agent) sendPostRequest(url string) {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("failed to build a request")
	}
	request.Header.Add("Content-Type", "text/plain")

	response, err := a.client.Do(request)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to make a request. error: %w", err))
		return
	}

	fmt.Println("Code: ", response.Status)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer response.Body.Close()
	fmt.Println(string(body))
}
