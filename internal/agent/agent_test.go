package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAgent(t *testing.T) {
	testAgent, err := NewAgent(Config{
		Address:        defaultAddress,
		PollInterval:   1 * time.Second,
		ReportInterval: 10 * time.Second,
		Pprof:          defaultPprofAddress,
	})

	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	go testAgent.Run(ctx)

	time.Sleep(2 * time.Second)
	cancel()

	testAgent.Stop()

	require.True(t, true)
}

func TestInvalidPublicKey(t *testing.T) {
	testAgent, err := NewAgent(Config{
		Address:        defaultAddress,
		PollInterval:   1 * time.Second,
		ReportInterval: 10 * time.Second,
		Pprof:          defaultPprofAddress,
		CryptoKey:      "notexists.txt",
	})

	require.Nil(t, testAgent)
	require.Error(t, err)
}
