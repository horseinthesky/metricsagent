package agent

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(fn),
	}
}

func TestSendMetricsJSONBulk(t *testing.T) {
	agent, err := NewAgent(Config{
		PollInterval:   time.Duration(2 * time.Second),
		ReportInterval: time.Duration(10 * time.Second),
	})

	require.NoError(t, err)

	agent.client = newTestClient(func(*http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"test": "passed"}`)),
		}
	})

	code, body, err := agent.sendPostJSONBulk(context.Background(), []Metric{{}})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, code)
	require.Equal(t, `{"test": "passed"}`, body)
}
