package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Metrics is an object to marshal metrics to.
type Metric struct {
	ID    string `json:"id"`              // metric name
	MType string `json:"type"`            // metric type, gauge/counter
	Delta *int64 `json:"delta,omitempty"` // metric value if it has a type of counter
	Value *gauge `json:"value,omitempty"` // metric value if it has a type of gauge
	Hash  string `json:"hash,omitempty"`  // hash value
}

// sendMetricsJSONBulk sends all metrics as one big JSON.
func (a *Agent) sendMetricsJSONBulk(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("sending metrics cancelled")
			return
		case <-a.ReportTicker.C:
			metrics := a.prepareMetrics()

			code, body, err := a.sendPostJSONBulk(ctx, metrics)
			if err != nil {
				log.Println(err)
				continue
			}

			log.Printf("Code: %v: %s", code, body)
		}
	}
}

// prepareMetrics() converts metrics data to Metric objects.
func (a *Agent) prepareMetrics() []Metric {
	metrics := []Metric{}

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

	metric := Metric{
		ID:    "PollCount",
		MType: "counter",
		Delta: &a.PollCounter,
	}

	if a.key != "" {
		a.addHash(&metric)
	}

	metrics = append(metrics, metric)

	return metrics
}

// sendPostJSONBulk serves as a HTTP helper for sendMetricsJSONBulk.
func (a *Agent) sendPostJSONBulk(ctx context.Context, metrics []Metric) (int, string, error) {
	endpoint := fmt.Sprintf("%s/updates/", a.upstream)

	payloadBytes, err := json.Marshal(metrics)
	if err != nil {
			return 0, "", fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if a.cryptoKey != nil {
		payloadBytes, err = encryptWithPublicKey(payloadBytes, a.cryptoKey)
		if err != nil {
			return 0, "", fmt.Errorf("failed to encrypt payload: %w", err)
		}
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return 0, "", fmt.Errorf("failed to build a request: %w", err)
	}
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(ctx)

	response, err := a.client.Do(request)
	if err != nil {
		return 0, "", fmt.Errorf("failed to make a request: %w", err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read response body: %w", err)
	}
	defer response.Body.Close()

	return response.StatusCode, string(body), nil
}

// // sendMetricsJSON sends metrics as JSON by one at a time.
// func (a *Agent) sendMetricsJSON() {
// 	// Send runtime metrics
// 	a.metrics.Range(func(metricName, value interface{}) bool {
// 		m, _ := metricName.(string)
// 		v, _ := value.(gauge)
//
// 		metric := Metric{
// 			ID:    m,
// 			MType: "gauge",
// 			Value: &v,
// 		}
//
// 		if a.key != "" {
// 			a.addHash(&metric)
// 		}
//
// 		a.sendPostJSON(metric)
// 		return true
// 	})
//
// 	// Send poll count
// 	metric := Metric{
// 		ID:    "PollCount",
// 		MType: "counter",
// 		Delta: &a.PollCounter,
// 	}
//
// 	if a.key != "" {
// 		a.addHash(&metric)
// 	}
//
// 	a.sendPostJSON(metric)
// }
//
// // sendMetricsPlain sends metrics by one at a time as URL params.
// func (a *Agent) sendMetricsPlain() {
// 	// Send metrics
// 	a.metrics.Range(func(metricName, value interface{}) bool {
// 		endpoint := fmt.Sprintf("%s/update/%s/%s/%v", a.upstream, "gauge", metricName, value)
// 		a.sendPostPlain(endpoint)
// 		return true
// 	})
//
// 	// Send poll count
// 	endpoint := fmt.Sprintf("%s/update/%s/%s/%v", a.upstream, "counter", "PollCount", a.PollCounter)
// 	a.sendPostPlain(endpoint)
// }
//
// // sendPostPlain serves as a HTTP helper for sendMetricsPlain.
// func (a *Agent) sendPostPlain(url string) {
// 	request, err := http.NewRequest(http.MethodPost, url, nil)
// 	if err != nil {
// 		log.Println(fmt.Errorf("failed to build a request: %w", err))
// 		return
// 	}
// 	request.Header.Add("Content-Type", "text/plain")
//
// 	response, err := a.client.Do(request)
// 	if err != nil {
// 		fmt.Println(fmt.Errorf("failed to make a request: %w", err))
// 		return
// 	}
//
// 	body, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		log.Println(fmt.Errorf("failed to read response body: %w", err))
// 		return
// 	}
// 	defer response.Body.Close()
//
// 	log.Printf("Code: %v: %s", response.Status, string(body))
// }
//
// // sendPostJSON serves as a HTTP helper for sendMetricsJSON.
// func (a *Agent) sendPostJSON(metric Metric) {
// 	payloadBuf := new(bytes.Buffer)
// 	json.NewEncoder(payloadBuf).Encode(metric)
//
// 	endpoint := fmt.Sprintf("%s/update/", a.upstream)
// 	request, err := http.NewRequest(http.MethodPost, endpoint, payloadBuf)
// 	if err != nil {
// 		log.Println(fmt.Errorf("failed to build a request: %w", err))
// 		return
// 	}
// 	request.Header.Add("Content-Type", "application/json")
//
// 	response, err := a.client.Do(request)
// 	if err != nil {
// 		log.Println(fmt.Errorf("failed to make a request: %w", err))
// 		return
// 	}
//
// 	body, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		log.Println(fmt.Errorf("failed to read response body: %w", err))
// 		return
// 	}
// 	defer response.Body.Close()
//
// 	log.Printf("Code: %v: %s", response.Status, string(body))
// }
