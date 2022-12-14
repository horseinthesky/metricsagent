package server

import (
	"os"
	"testing"
	"time"

	"github.com/horseinthesky/metricsagent/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {
	curDir, err := os.Getwd()
	assert.NoError(t, err)

	testServer, err := NewServer(Config{
		Address:       defaultListenOn,
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
		CryptoKey:      curDir + "/testdata/private.pem",
	})
	require.NoError(t, err)

	testAgent, err := agent.NewAgent(agent.Config{
		PollInterval:   1 * time.Second,
		ReportInterval: 10 * time.Second,
		CryptoKey:      curDir + "/testdata/public.pem",
	})
	require.NoError(t, err)

	testMsg := []byte("plaintext")

	encryptedBytes, err := agent.EncryptWithPublicKey(testMsg, testAgent.CryptoKey)
	require.NoError(t, err)
	require.NotZero(t, encryptedBytes)

	decryptedBytes, err := decryptWithPrivateKey(encryptedBytes, testServer.CryptoKey)
	require.NoError(t, err)
	require.Equal(t, testMsg, decryptedBytes)
}
