package main

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestUpdateMetrics(t *testing.T) {
	runtime.ReadMemStats(data)
	updateMetrcis()

	assert.Equal(t, gauge(data.Alloc), metrics["Alloc"])
}
