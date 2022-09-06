package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func (a *Agent) SendMetricsJSONBulk() {
	metrics := []Metric{}

	// Send runtime metrics
	a.metrics.Range(func(metricName, value interface{}) bool {
		m, _ := metricName.(string)
		v, _ := value.(gauge)

		metric := Metric{
			ID:    m,
			MType: "gauge",
			Value: &v,
		}

		if a.key != "" {
			a.addHash(&metric)
		}

		metrics = append(metrics, metric)

		return true
	})

	// Send poll count
	metric := Metric{
		ID:    "PollCount",
		MType: "counter",
		Delta: &a.PollCounter,
	}

	if a.key != "" {
		a.addHash(&metric)
	}

	metrics = append(metrics, metric)

	a.sendPostJSONBulk(metrics)
}

func (a *Agent) SendMetricsJSON() {
	// Send runtime metrics
	a.metrics.Range(func(metricName, value interface{}) bool {
		m, _ := metricName.(string)
		v, _ := value.(gauge)

		metric := Metric{
			ID:    m,
			MType: "gauge",
			Value: &v,
		}

		if a.key != "" {
			a.addHash(&metric)
		}

		a.sendPostJSON(metric)
		return true
	})

	// Send poll count
	metric := Metric{
		ID:    "PollCount",
		MType: "counter",
		Delta: &a.PollCounter,
	}

	if a.key != "" {
		a.addHash(&metric)
	}

	a.sendPostJSON(metric)
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

func (a *Agent) sendPostJSON(metric Metric) {
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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(fmt.Errorf("failed to read response body: %w", err))
		return
	}
	defer response.Body.Close()

	log.Printf("Code: %v: %s", response.Status, string(body))
}

func (a *Agent) sendPostJSONBulk(metrics []Metric) {
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(metrics)

	endpoint := fmt.Sprintf("%s/updates/", a.upstream)
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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(fmt.Errorf("failed to read response body: %w", err))
		return
	}
	defer response.Body.Close()

	log.Printf("Code: %v: %s", response.Status, string(body))
}
