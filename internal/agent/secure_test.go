package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddhash(t *testing.T) {
	agent := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
		Key: "testkey",
	})

	testValue := counter(15)

	metric := Metric{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &testValue,
	}

	agent.addHash(&metric)
	assert.Equal(t, "175b2a772fbf2ad97bb515e10f2c24bdaf75860e18f8999c6825be73acd3e6bc", metric.Hash)
}
