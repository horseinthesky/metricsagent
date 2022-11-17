package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, payload string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer([]byte(payload)))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	testServer := NewServer(Config{
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
	})

	ts := httptest.NewServer(testServer)
	defer ts.Close()

	// test no route
	resp, body := testRequest(t, ts, "GET", "/notexists", "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "404 page not found\n", body)

	// test ping db
	resp, body = testRequest(t, ts, "GET", "/ping", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, http.StatusText(http.StatusOK), body)

	// test save unsupported metric type plain
	resp, body = testRequest(t, ts, "POST", "/update/unsupported/testUnsupported/100", "")
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Equal(t, http.StatusText(http.StatusNotImplemented), body)

	// test save valid metric counter
	resp, body = testRequest(t, ts, "POST", "/update/counter/testCounter/100", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Success: metric stored\n", body)

	// test save valid metric gauge
	resp, body = testRequest(t, ts, "POST", "/update/gauge/testGauge/10.0", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Success: metric stored\n", body)

	// test save unsupported metric type JSON
	resp, body = testRequest(t, ts, "POST", "/update", `{"id": "testJSONGauge", "type": "unsupported", "value": 1}`)
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Equal(t, "{\"error\": \"unsupported metric type\"}\n", body)

	// test save valid metric counter JSON
	resp, body = testRequest(t, ts, "POST", "/update", `{"id": "testJSONCounter", "type": "counter", "delta": 110}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{\"result\": \"metric saved\"}", body)

	// test save valid metric gauge JSON
	resp, body = testRequest(t, ts, "POST", "/update", `{"id": "testJSONGauge", "type": "gauge", "value": 11.0}`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{\"result\": \"metric saved\"}", body)

	// test save unsupported JSON metrics
	unsupportedMetrics := `
		[
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
		]
	`
	resp, body = testRequest(t, ts, "POST", "/updates/", unsupportedMetrics)
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	assert.Equal(t, "{\"error\": \"unsupported metric type\"}\n", body)

	// test save valid JSON metrics
	metrics := `
		[
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
		]
	`
	resp, body = testRequest(t, ts, "POST", "/updates/", metrics)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{\"result\": \"metric saved\"}", body)

	// test get counter plain
	resp, body = testRequest(t, ts, "GET", "/value/counter/testCounter", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "100", body)

	// test get not existing counter plain
	resp, body = testRequest(t, ts, "GET", "/value/counter/testNotExist", "")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, http.StatusText(http.StatusNotFound), body)

	// test get gauge JSON
	jsonMetric := `
		{
		  "id": "testJSONGauge",
		  "type": "gauge"
		}
	`
	resp, body = testRequest(t, ts, "POST", "/value/", jsonMetric)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, `{"id":"testJSONGauge","type":"gauge","value":11}`, body)

	// test get not existing gauge JSON
	notExistingJsonMetric := `
		{
		  "id": "testNotExistingJSON",
		  "type": "gauge"
		}
	`
	resp, body = testRequest(t, ts, "POST", "/value/", notExistingJsonMetric)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
