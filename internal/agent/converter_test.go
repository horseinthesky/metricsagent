package agent

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareMetrics(t *testing.T) {
	storage := &sync.Map{}
	testKey := "testkey"

	metrics := prepareMetrics(storage, 1, testKey)
	require.Equal(t, len(metrics), 1)

	updateRuntimeMetrics(storage)
	metrics = prepareMetrics(storage, 2, testKey)
	require.Greater(t, len(metrics), 1)
}

func TestMetricToPB(t *testing.T) {
	counter := int64(10)

	counterMmetric := Metric{
		ID: "TestCounter",
		MType: "counter",
		Delta: &counter,
	}

	gauge := gauge(10.5)

	gaugeMmetric := Metric{
		ID: "TestGauge",
		MType: "gauge",
		Value: &gauge,
	}

	pbCounter := MetricToPB(counterMmetric)
	require.Equal(t, counterMmetric.ID, pbCounter.Id)
	require.Equal(t, *counterMmetric.Delta, pbCounter.Delta)

	pbGauge := MetricToPB(gaugeMmetric)
	require.Equal(t, gaugeMmetric.ID, pbGauge.Id)
	require.Equal(t, *gaugeMmetric.Value, pbGauge.Value)
}
