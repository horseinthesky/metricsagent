package agent

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var storage = &sync.Map{}

func TestUpdateRuntimeMetrics(t *testing.T) {
	updateRuntimeMetrics(storage)

	AllocMetric, loaded := storage.Load("Alloc")
	require.True(t, loaded)
	require.NotEqual(t, 0, AllocMetric)
}

func TestUpdatePSUtilMetrics(t *testing.T) {
	updatePSUtilMetrics(storage)

	TotalMemoryMetric, loaded := storage.Load("TotalMemory")
	require.True(t, loaded)
	require.Greater(t, TotalMemoryMetric, gauge(0))
}
