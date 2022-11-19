package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, payload string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer([]byte(payload)))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestRouter(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		payload string
		expected    int
	}{
		{
			name:   "test no route",
			method: http.MethodGet,
			path:   "/notexists",
			expected:   http.StatusNotFound,
		},
		{
			name:   "test ping db",
			method: http.MethodGet,
			path:   "/ping",
			expected:   http.StatusOK,
		},
		{
			name:   "test save unsupported metric type plain",
			method: http.MethodPost,
			path:   "/update/unsupported/testUnsupported/100",
			expected:   http.StatusNotImplemented,
		},
		{
			name:   "test save valid metric counter",
			method: http.MethodPost,
			path:   "/update/counter/testCounter/100",
			expected:   http.StatusOK,
		},
		{
			name:   "test save valid metric gauge",
			method: http.MethodPost,
			path:   "/update/gauge/testGauge/10.0",
			expected:   http.StatusOK,
		},
		{
			name:   "test save unsupported metric type JSON",
			method: http.MethodPost,
			path:   "/update",
			payload: `{
				"id": "testJSONGauge",
				"type": "unsupported",
				"value": 1
			}`,
			expected: http.StatusNotImplemented,
		},
		{
			name:   "test save valid metric counter JSON",
			method: http.MethodPost,
			path:   "/update",
			payload: `{
				"id": "testJSONCounter",
				"type": "counter",
				"delta": 110
			}`,
			expected: http.StatusOK,
		},
		{
			name:   "test save valid metric gauge JSON",
			method: http.MethodPost,
			path:   "/update",
			payload: `{
				"id": "testJSONGauge",
				"type": "gauge",
				"value": 11.0
			}`,
			expected: http.StatusOK,
		},
		{
			name:   "test save unsupported JSON metrics",
			method: http.MethodPost,
			path:   "/updates/",
			payload: `[
				{
					"id": "testJSONCounter1",
					"type": "counter",
					"delta": 210
				},
				{
					"id": "testJSONGauge1",
					"type": "unsupported",
					"value": 400
				}
			]`,
			expected: http.StatusNotImplemented,
		},
		{
			name:   "test save valid JSON metrics",
			method: http.MethodPost,
			path:   "/updates/",
			payload: `[
				{
					"id": "testJSONCounter1",
					"type": "counter",
					"delta": 210
				},
				{
					"id": "testJSONGauge1",
					"type": "gauge",
					"value": 21.0
				}
			]`,
			expected: http.StatusOK,
		},
		{
			name:   "test get not existing counter plain",
			method: http.MethodGet,
			path:   "/value/counter/testNotExist",
			expected:   http.StatusNotFound,
		},
		{
			name:   "test get counter plain",
			method: http.MethodGet,
			path:   "/value/counter/testCounter",
			expected:   http.StatusOK,
		},
		{
			name:   "test get not existing gauge JSON",
			method: http.MethodPost,
			path:   "/value/",
			payload: `
				{
					"id": "testNotExistingJSON",
					"type": "gauge"
				}
			`,
			expected: http.StatusNotFound,
		},
		{
			name:   "test get gauge JSON",
			method: http.MethodPost,
			path:   "/value/",
			payload: `
				{
					"id": "testJSONGauge",
					"type": "gauge"
				}
			`,
			expected: http.StatusOK,
		},
	}

	testServer := NewServer(Config{
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
	})

	ts := httptest.NewServer(testServer)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := testRequest(t, ts, tt.method, tt.path, tt.payload)
			require.Equal(t, tt.expected, code)
		})
	}
}
