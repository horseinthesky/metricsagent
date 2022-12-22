package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	testServer       *Server
	testHashedServer *Server
)

func init() {
	testServer, _ = NewServer(Config{
		Address:       defaultListenOn,
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
	})

	testHashedServer, _ = NewServer(Config{
		Address:       "localhost:8085",
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
		Key:           "testkey",
	})
}

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

func TestServerRun(t *testing.T) {
	testServer.restore()

	ctx, cancel := context.WithCancel(context.Background())
	go testServer.Run(ctx)

	time.Sleep(2 * time.Second)
	testServer.dump()

	cancel()
	testServer.Stop()
}

func TestGeneral(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		payload  string
		expected int
	}{
		{
			name:     "test no route",
			method:   http.MethodGet,
			path:     "/notexists",
			expected: http.StatusNotFound,
		},
		{
			name:     "test ping db",
			method:   http.MethodGet,
			path:     "/ping",
			expected: http.StatusOK,
		},
	}

	ts := httptest.NewServer(testServer)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := testRequest(t, ts, tt.method, tt.path, tt.payload)
			require.Equal(t, tt.expected, code)
		})
	}
}

func TestTextHandlers(t *testing.T) {
	saveTests := []struct {
		name     string
		method   string
		path     string
		payload  string
		expected int
	}{
		{
			name:     "test save unsupported metric type plain",
			method:   http.MethodPost,
			path:     "/update/unsupported/testUnsupported/100",
			expected: http.StatusNotImplemented,
		},
		{
			name:     "test save invalid metric counter",
			method:   http.MethodPost,
			path:     "/update/counter/invalidCounter/invalidCounter",
			expected: http.StatusBadRequest,
		},
		{
			name:     "test save invalid metric gauge",
			method:   http.MethodPost,
			path:     "/update/gauge/invalidCounter/invalidCounter",
			expected: http.StatusBadRequest,
		},
		{
			name:     "test save valid metric counter",
			method:   http.MethodPost,
			path:     "/update/counter/testCounter/100",
			expected: http.StatusOK,
		},
		{
			name:     "test save valid metric gauge",
			method:   http.MethodPost,
			path:     "/update/gauge/testGauge/10.0",
			expected: http.StatusOK,
		},
	}

	loadTests := []struct {
		name     string
		method   string
		path     string
		payload  string
		expected int
		body     string
	}{
		{
			name:     "test get not existing counter plain",
			method:   http.MethodGet,
			path:     "/value/counter/testNotExists",
			expected: http.StatusNotFound,
			body:     http.StatusText(http.StatusNotFound),
		},
		{
			name:     "test get counter plain",
			method:   http.MethodGet,
			path:     "/value/counter/testCounter",
			expected: http.StatusOK,
			body:     "100",
		},
		{
			name:     "test get gauge plain",
			method:   http.MethodGet,
			path:     "/value/gauge/testGauge",
			expected: http.StatusOK,
			body:     "10",
		},
	}

	ts := httptest.NewServer(testServer)
	defer ts.Close()

	for _, tt := range saveTests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := testRequest(t, ts, tt.method, tt.path, tt.payload)
			require.Equal(t, tt.expected, code)
		})
	}

	for _, tt := range loadTests {
		t.Run(tt.name, func(t *testing.T) {
			code, body := testRequest(t, ts, tt.method, tt.path, tt.payload)
			require.Equal(t, tt.expected, code)
			require.Equal(t, tt.body, body)
		})
	}
}

func TestJSONHandlers(t *testing.T) {
	saveTests := []struct {
		name     string
		method   string
		path     string
		payload  string
		expected int
	}{
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
	}

	loadTests := []struct {
		name     string
		method   string
		path     string
		payload  string
		expected int
		body     string
	}{
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
			body:     `{"result": "unknown metric id"}`,
		},
		{
			name:   "test get counter JSON",
			method: http.MethodPost,
			path:   "/value/",
			payload: `
				{
					"id": "testJSONCounter",
					"type": "counter"
				}
			`,
			expected: http.StatusOK,
			body:     `{"id":"testJSONCounter","type":"counter","delta":110}`,
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
			body:     `{"id":"testJSONGauge","type":"gauge","value":11}`,
		},
	}

	ts := httptest.NewServer(testServer)
	defer ts.Close()

	for _, tt := range saveTests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := testRequest(t, ts, tt.method, tt.path, tt.payload)
			require.Equal(t, tt.expected, code)
		})
	}

	for _, tt := range loadTests {
		t.Run(tt.name, func(t *testing.T) {
			code, body := testRequest(t, ts, tt.method, tt.path, tt.payload)
			require.Equal(t, tt.expected, code)
			require.Equal(t, tt.body, body)
		})
	}
}

func TestJSONHandlersHashed(t *testing.T) {
	saveTests := []struct {
		name     string
		method   string
		path     string
		payload  string
		expected int
	}{
		{
			name:   "test save hashed counter JSON",
			method: http.MethodPost,
			path:   "/update",
			payload: `{
				"id": "TestCounter",
				"type": "counter",
				"delta": 15,
				"hash": "175b2a772fbf2ad97bb515e10f2c24bdaf75860e18f8999c6825be73acd3e6bc"
			}`,
			expected: http.StatusOK,
		},
		{
			name:   "test save hashed Gauge JSON",
			method: http.MethodPost,
			path:   "/update",
			payload: `{
				"id": "TestGauge",
				"type": "gauge",
				"value": 15,
				"hash": "7300c53d565107966dd4486f13c76cdeda0e31d7f49a62494e5921f8a0faf417"
			}`,
			expected: http.StatusOK,
		},
		{
			name:   "test save invalid hashed Gauge JSON",
			method: http.MethodPost,
			path:   "/update",
			payload: `{
				"id": "TestInvalidHashedGauge",
				"type": "gauge",
				"value": 15,
				"hash": "invalidhash"
			}`,
			expected: http.StatusInternalServerError,
		},
		{
			name:   "test save wrong hashed Gauge JSON",
			method: http.MethodPost,
			path:   "/update",
			payload: `{
				"id": "TestWrongHashedGauge",
				"type": "gauge",
				"value": 15,
				"hash": "7300c53d565107966dd4486f13c76cdeda0e31d7f49a62494e5921f8a0faf417"
			}`,
			expected: http.StatusBadRequest,
		},
		{
			name:   "test save a pair of hashed metrics",
			method: http.MethodPost,
			path:   "/updates/",
			payload: `[
				{
					"id": "TestCounter",
					"type": "counter",
					"delta": 15,
					"hash": "175b2a772fbf2ad97bb515e10f2c24bdaf75860e18f8999c6825be73acd3e6bc"
				},
				{
					"id": "TestGauge",
					"type": "gauge",
					"value": 15,
					"hash": "7300c53d565107966dd4486f13c76cdeda0e31d7f49a62494e5921f8a0faf417"
				}
			]`,
			expected: http.StatusOK,
		},
		{
			name:   "test save invalid hashed metrics",
			method: http.MethodPost,
			path:   "/updates/",
			payload: `[
				{
					"id": "TestInvalidHashCounter",
					"type": "counter",
					"delta": 15,
					"hash": "invalidhash"
				}
			]`,
			expected: http.StatusInternalServerError,
		},
		{
			name:   "test save wrong hashed metrics",
			method: http.MethodPost,
			path:   "/updates/",
			payload: `[
				{
					"id": "TestWrongHashedCounter",
					"type": "counter",
					"delta": 15,
					"hash": "175b2a772fbf2ad97bb515e10f2c24bdaf75860e18f8999c6825be73acd3e6bc"
				}
			]`,
			expected: http.StatusBadRequest,
		},
	}

	ts := httptest.NewServer(testHashedServer)
	defer ts.Close()

	for _, tt := range saveTests {
		t.Run(tt.name, func(t *testing.T) {
			code, _ := testRequest(t, ts, tt.method, tt.path, tt.payload)
			require.Equal(t, tt.expected, code)
		})
	}
}

func TestDashBoard(t *testing.T) {
	ts := httptest.NewServer(testServer)
	defer ts.Close()

	code, body := testRequest(t, ts, http.MethodGet, "/", "")
	require.Equal(t, http.StatusOK, code)
	require.NotEmpty(t, body)
}
