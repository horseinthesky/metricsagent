package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/horseinthesky/metricsagent/internal/server/storage"
)

var (
	testServer       *GenericServer
)

func init() {
	testServer, _ = NewGenericServer(Config{
		Address:       defaultListenOn,
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
	})

}

func TestGenericServerRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	testServer.Bootstrap(ctx)

	time.Sleep(2 * time.Second)

	cancel()
	testServer.WorkGroup.Wait()
}

func TestGenericServerSaveMetric(t *testing.T) {
	counterValue1 := int64(10)

	counter1 := storage.Metric{
		ID: "testCounter",
		MType: "counter",
		Delta: &counterValue1,
	}

	counterValue2 := int64(10)

	counter2 := storage.Metric{
		ID: "testCounter",
		MType: "counter",
		Delta: &counterValue2,
	}

	err := testServer.SaveMetric(counter1)
	require.NoError(t, err)

	err = testServer.SaveMetricsBulk([]storage.Metric{counter1, counter2})
	require.NoError(t, err)

	testServer.dump()
	testServer.restore()
}
