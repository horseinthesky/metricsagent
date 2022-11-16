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

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestSendMetricsJSONBulk(t *testing.T) {
	agent := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})
	agent.client = NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusCreated,
			Body:       io.NopCloser(strings.NewReader(`{"error":"Character not found"}`)),
		}
	})

	testValue := counter(15)
	metrics := []Metric{{
		ID:    "TestCounter",
		MType: "counter",
		Delta: &testValue,
	}}

	code, body, err := agent.sendPostJSONBulk(context.Background(), metrics)
	assert.NoError(t, err)
	assert.Equal(t, 201, code)
	assert.NotEqual(t, "", body)
}
