package agent

import (
	"sync"

	"github.com/horseinthesky/metricsagent/internal/pb"
)

// Metrics is an object to marshal metrics to.
type Metric struct {
	ID    string `json:"id"`              // metric name
	MType string `json:"type"`            // metric type, gauge/counter
	Delta *int64 `json:"delta,omitempty"` // metric value if it has a type of counter
	Value *gauge `json:"value,omitempty"` // metric value if it has a type of gauge
	Hash  string `json:"hash,omitempty"`  // hash value
}

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

func MetricToPB(metric Metric) *pb.Metric {
	pbMetric := &pb.Metric{
		Id:    metric.ID,
		Mtype: metric.MType,
		Hash:  metric.Hash,
	}

	if metric.Delta != nil {
		pbMetric.Delta = *metric.Delta
	}

	if metric.Value != nil {
		pbMetric.Value = *metric.Value
	}

	return pbMetric
}
