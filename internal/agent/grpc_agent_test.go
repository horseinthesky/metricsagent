package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGRPCAgent(t *testing.T) {
	testAgent, err := NewGRPCAgent(Config{
		Address:        defaultAddress,
		PollInterval:   1 * time.Second,
		ReportInterval: 10 * time.Second,
	})

	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go testAgent.Run(ctx)

	time.Sleep(2 * time.Second)
	cancel()

	testAgent.Stop()

	require.True(t, true)
}

func TestGRPCInvalidPublicKey(t *testing.T) {
	testAgent, err := NewGRPCAgent(Config{
		Address:        defaultAddress,
		PollInterval:   1 * time.Second,
		ReportInterval: 10 * time.Second,
		CryptoKey:      "notexists.txt",
	})

	require.Nil(t, testAgent)
	require.Error(t, err)
}
