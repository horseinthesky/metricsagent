package agent

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

func TestPrepareMetrics(t *testing.T) {
	agent := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
		Key: "testkey",
	})

	metrics := agent.prepareMetrics()
	assert.Equal(t, len(metrics), 1)

	agent.updateRuntimeMetrics()
	metrics = agent.prepareMetrics()
	assert.Greater(t, len(metrics), 1)
}

func TestSendMetricsJSONBulk(t *testing.T) {
	agent := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})
	agent.client = NewTestClient(func(*http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"test": "passed"}`)),
		}
	})

	code, body, err := agent.sendPostJSONBulk(context.Background(), []Metric{{}})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, `{"test": "passed"}`, body)
}
