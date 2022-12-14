package agent

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrypto(t *testing.T) {
	curDir, err := os.Getwd()
	assert.NoError(t, err)

	testAgent, err := NewAgent(Config{
		Address:        defaultAddress,
		PollInterval:   1 * time.Second,
		ReportInterval: 10 * time.Second,
		Pprof:          defaultPprofAddress,
		CryptoKey:      curDir + "/testdata/public.pem",
	})
	require.NoError(t, err)

	encryptedBytes, err := encryptWithPublicKey([]byte("plaintext"), testAgent.cryptoKey)
	require.NoError(t, err)
	require.NotZero(t, encryptedBytes)
}
