package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateRuntimeMetrics(t *testing.T) {
	agent := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})

	agent.updateRuntimeMetrics()

	AllocMetric, loaded := agent.metrics.Load("Alloc")
	assert.True(t, loaded)
	assert.Equal(t, counter(1), agent.PollCounter)
	assert.NotEqual(t, 0, AllocMetric)
}

func TestUpdatePSUtilMetrics(t *testing.T) {
	agent := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})

	agent.updatePSUtilMetrics()

	TotalMemoryMetric, loaded := agent.metrics.Load("TotalMemory")
	assert.True(t, loaded)
	assert.Greater(t, TotalMemoryMetric, gauge(0))
}
