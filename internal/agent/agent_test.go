package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAgent(t *testing.T) {
	testAgent := NewAgent(Config{
		Address:        defaultAddress,
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Pprof:          defaultPprofAddress,
	})

	ctx, cancel := context.WithCancel(context.Background())
	go testAgent.Run(ctx)

	time.Sleep(3 * time.Second)
	cancel()

	testAgent.Stop()

	require.True(t, true)
}
