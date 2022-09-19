package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetrics(t *testing.T) {
	agent := New(&Config{
		Address:        "localhost:8080",
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})

	agent.updateRuntimeMetrics()

	storageMetric, loaded := agent.metrics.Load("Alloc")
	assert.True(t, loaded)
	assert.Equal(t, gauge(agent.data.Alloc), storageMetric)
}
