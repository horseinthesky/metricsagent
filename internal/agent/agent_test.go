package agent

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	data = &runtime.MemStats{}
)

func TestUpdateMetrics(t *testing.T) {
	runtime.ReadMemStats(data)

	agent := New(&Config{
		Address:        "localhost:8080",
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})

	agent.UpdateMetrics(data)

	storageMetric, loaded := agent.metrics.Load("Alloc")
	assert.True(t, loaded)
	assert.Equal(t, gauge(data.Alloc), storageMetric)
}
