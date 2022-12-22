package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpdateRuntimeMetrics(t *testing.T) {
	agent, err := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})

	require.NoError(t, err)

	agent.updateRuntimeMetrics()

	AllocMetric, loaded := agent.metrics.Load("Alloc")
	require.True(t, loaded)
	require.Equal(t, counter(1), agent.PollCounter)
	require.NotEqual(t, 0, AllocMetric)
}

func TestUpdatePSUtilMetrics(t *testing.T) {
	agent, err := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})

	require.NoError(t, err)

	agent.updatePSUtilMetrics()

	TotalMemoryMetric, loaded := agent.metrics.Load("TotalMemory")
	require.True(t, loaded)
	require.Greater(t, TotalMemoryMetric, gauge(0))
}
