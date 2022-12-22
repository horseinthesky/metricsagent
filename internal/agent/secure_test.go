package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAddhash(t *testing.T) {
	agent, err := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
		Key: "testkey",
	})

	require.NoError(t, err)

	testCounter := counter(15)

	counterMetric := Metric{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &testCounter,
	}

	agent.addHash(&counterMetric)
	require.Equal(t, "175b2a772fbf2ad97bb515e10f2c24bdaf75860e18f8999c6825be73acd3e6bc", counterMetric.Hash)

	testGauge := gauge(15.0)

	gaugeMetric := Metric{
		ID:    "TestGauge",
		MType: "gauge",
		Value: &testGauge,
	}

	agent.addHash(&gaugeMetric)
	require.Equal(t, "7300c53d565107966dd4486f13c76cdeda0e31d7f49a62494e5921f8a0faf417", gaugeMetric.Hash)
}
