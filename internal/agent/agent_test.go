package agent

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	data = &runtime.MemStats{}
)

func TestUpdateMetrics(t *testing.T) {
	runtime.ReadMemStats(data)

	agent := New(2, 10, "")
	agent.UpdateMetrics(data)

	storageMetric, loaded := agent.metrics.Load("Alloc")
	assert.True(t, loaded)
	assert.Equal(t, gauge(data.Alloc), storageMetric)
}
