package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetrics(t *testing.T) {
	agent := New(&Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})

	agent.updateRuntimeMetrics()

	storageMetric, loaded := agent.metrics.Load("Alloc")
	assert.True(t, loaded)
	assert.NotEqual(t, 0, storageMetric)
}
