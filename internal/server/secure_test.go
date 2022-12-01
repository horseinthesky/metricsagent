package server

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
	"github.com/stretchr/testify/require"
)

func TestGeneratehash(t *testing.T) {
	testServer := NewServer(Config{
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
		Key:           "testkey",
	})

	testCounter := `{
		"id": "TestCounter",
		"type": "counter",
		"delta": 15,
		"hash": "175b2a772fbf2ad97bb515e10f2c24bdaf75860e18f8999c6825be73acd3e6bc"
	}`

	testCounterMetric := storage.Metric{}
	err := json.Unmarshal([]byte(testCounter), &testCounterMetric)
	require.NoError(t, err)

	localHash := testServer.generateHash(testCounterMetric)
	remoteHash, err := hex.DecodeString(testCounterMetric.Hash)
	require.NoError(t, err)

	require.True(t, hmac.Equal(localHash, remoteHash), "Local and remote hashes differ")

	testGauge := `{
		"id": "TestGauge",
		"type": "gauge",
		"value": 15,
		"hash": "7300c53d565107966dd4486f13c76cdeda0e31d7f49a62494e5921f8a0faf417"
	}`

	testGaugeMeric := storage.Metric{}
	err = json.Unmarshal([]byte(testGauge), &testGaugeMeric)
	require.NoError(t, err)

	localHash = testServer.generateHash(testGaugeMeric)
	remoteHash, err = hex.DecodeString(testGaugeMeric.Hash)
	require.NoError(t, err)

	require.True(t, hmac.Equal(localHash, remoteHash), "Local and remote hashes differ")
}
