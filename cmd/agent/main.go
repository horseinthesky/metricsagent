package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

type counter int64
type gauge float64

var (
	data = &runtime.MemStats{}
	metrics map[string]gauge
	pollNum uint
	baseUrl string = "http://localhost:8080/"
	client         = &http.Client{
		Timeout: 1 * time.Second,
	}
)

func updateMetrcis() {
	metrics = map[string]gauge{
		"Alloc":         gauge(data.Alloc),
		"BuckHashSys":   gauge(data.BuckHashSys),
		"Frees":         gauge(data.Frees),
		"GCCPUFraction": gauge(data.GCCPUFraction),
		"GCSys":         gauge(data.GCSys),
		"HeapAlloc":     gauge(data.HeapAlloc),
		"HeapIdle":      gauge(data.HeapIdle),
		"HeapInuse":     gauge(data.HeapInuse),
		"HeapObjects":   gauge(data.HeapObjects),
		"HeapReleased":  gauge(data.HeapReleased),
		"HeapSys":       gauge(data.HeapSys),
		"LastGC":        gauge(data.LastGC),
		"Lookups":       gauge(data.Lookups),
		"MCacheInuse":   gauge(data.MCacheInuse),
		"MCacheSys":     gauge(data.MCacheSys),
		"MSpanInuse":    gauge(data.MSpanInuse),
		"MSpanSys":      gauge(data.MSpanSys),
		"Mallocs":       gauge(data.Mallocs),
		"NextGC":        gauge(data.NextGC),
		"NumForcedGC":   gauge(data.NumForcedGC),
		"NumGC":         gauge(data.NumGC),
		"OtherSys":      gauge(data.OtherSys),
		"PauseTotalNs":  gauge(data.PauseTotalNs),
		"StackInuse":    gauge(data.StackInuse),
		"StackSys":      gauge(data.StackSys),
		"Sys":           gauge(data.Sys),
		"TotalAlloc":    gauge(data.TotalAlloc),
		"Rand":          gauge(rand.Float64()),
	}
}

func sendRequest(request *http.Request) {
	request.Header.Add("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Code: ", response.Status)
		defer response.Body.Close()
	}
}

func sendMetrics() {
	// Send metrics
	for metricName, value := range metrics {
		endpoint := fmt.Sprintf("%s/update/%s/%s/%v", baseUrl, "gauge", metricName, value)

		request, err := http.NewRequest(http.MethodPost, endpoint, nil)
		if err != nil {
			fmt.Println("failed to build a request")
		}

		sendRequest(request)
	}

	// Send poll count
	endpoint := fmt.Sprintf("%s/update/%s/%s/%v", baseUrl, "counter", "pollNum", pollNum)
	request, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		fmt.Println("failed to build a request")
	}

	sendRequest(request)
}

func main() {
	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	for {
		select {
		case <-reportTicker.C:
			sendMetrics()
		case <-pollTicker.C:
			pollNum++

			runtime.ReadMemStats(data)

			updateMetrcis()
		}
	}
}
