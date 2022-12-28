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
