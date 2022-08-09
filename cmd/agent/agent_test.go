package main

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestUpdateMetrics(t *testing.T) {
	runtime.ReadMemStats(data)

	agent := newAgent(pollInterval, reportInterval, "")
	agent.updateMetrics()

	storageMetric, loaded := metrics.Load("Alloc")
	assert.True(t, loaded)
	assert.Equal(t, gauge(data.Alloc), storageMetric)
}
