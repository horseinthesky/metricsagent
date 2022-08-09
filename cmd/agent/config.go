package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL string = "http://localhost:8080"
)

var (
	metrics = &sync.Map{}
	client  = &http.Client{
		Timeout: 1 * time.Second,
	}
)

type gauge float64

type Agent struct {
	poll     *time.Ticker
	report   *time.Ticker
	upstream string
	count    int
}

func newAgent(poll, report time.Duration, url string) *Agent {
	agent := &Agent{
		poll:   time.NewTicker(poll),
		report: time.NewTicker(report),
	}

	if url == "" {
		url = baseURL
	}
	agent.upstream = url

	return agent
}

func (a *Agent) updateMetrics() {
	metrics.Store("Alloc", gauge(data.Alloc))
	metrics.Store("BuckHashSys", gauge(data.BuckHashSys))
	metrics.Store("Frees", gauge(data.Frees))
	metrics.Store("GCCPUFraction", gauge(data.GCCPUFraction))
	metrics.Store("GCSys", gauge(data.GCSys))
	metrics.Store("HeapAlloc", gauge(data.HeapAlloc))
	metrics.Store("HeapIdle", gauge(data.HeapIdle))
	metrics.Store("HeapInuse", gauge(data.HeapInuse))
	metrics.Store("HeapObjects", gauge(data.HeapObjects))
	metrics.Store("HeapReleased", gauge(data.HeapReleased))
	metrics.Store("HeapSys", gauge(data.HeapSys))
	metrics.Store("LastGC", gauge(data.LastGC))
	metrics.Store("Lookups", gauge(data.Lookups))
	metrics.Store("MCacheInuse", gauge(data.MCacheInuse))
	metrics.Store("MCacheSys", gauge(data.MCacheSys))
	metrics.Store("MSpanInuse", gauge(data.MSpanInuse))
	metrics.Store("MSpanSys", gauge(data.MSpanSys))
	metrics.Store("Mallocs", gauge(data.Mallocs))
	metrics.Store("NextGC", gauge(data.NextGC))
	metrics.Store("NumForcedGC", gauge(data.NumForcedGC))
	metrics.Store("NumGC", gauge(data.NumGC))
	metrics.Store("OtherSys", gauge(data.OtherSys))
	metrics.Store("PauseTotalNs", gauge(data.PauseTotalNs))
	metrics.Store("StackInuse", gauge(data.StackInuse))
	metrics.Store("StackSys", gauge(data.StackSys))
	metrics.Store("Sys", gauge(data.Sys))
	metrics.Store("TotalAlloc", gauge(data.TotalAlloc))
	metrics.Store("Rand", gauge(rand.Float64()))
}

func (a Agent) sendMetrics() {
	// Send metrics
	metrics.Range(func(metricName, value interface{}) bool {
		endpoint := fmt.Sprintf("%s/update/%s/%s/%v", a.upstream, "gauge", metricName, value)
		sendPostRequest(endpoint)
		return true
	})

	// Send poll count
	endpoint := fmt.Sprintf("%s/update/%s/%s/%v", a.upstream, "counter", "pollNum", a.count)
	sendPostRequest(endpoint)
}

func sendPostRequest(url string) {
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		fmt.Println("failed to build a request")
	}
	request.Header.Add("Content-Type", "text/plain")

	response, err := client.Do(request)
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
