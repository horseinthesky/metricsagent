package agent

import "sync"

// prepareMetrics converts metrics data to Metric objects.
func prepareMetrics(storage *sync.Map, pollCounter int64, hashKey string) []Metric {
	metrics := []Metric{}

	storage.Range(func(metricName, value interface{}) bool {
		m, _ := metricName.(string)
		v, _ := value.(gauge)

		metric := Metric{
			ID:    m,
			MType: "gauge",
			Value: &v,
		}

		if hashKey != "" {
			metric = addHash(metric, hashKey)
		}

		metrics = append(metrics, metric)

		return true
	})

	metric := Metric{
		ID:    "PollCount",
		MType: "counter",
		Delta: &pollCounter,
	}

	if hashKey != "" {
		metric = addHash(metric, hashKey)
	}

	metrics = append(metrics, metric)

	return metrics
}
