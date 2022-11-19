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
	testServer := NewServer(Config{
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
	})

	ts := httptest.NewServer(testServer)
	defer ts.Close()

	// test no route
	code, _ := testRequest(t, ts, "GET", "/notexists", "")
	assert.Equal(t, http.StatusNotFound, code)

	// test ping db
	code, payload := testRequest(t, ts, "GET", "/ping", "")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, http.StatusText(http.StatusOK), payload)

	// test save unsupported metric type plain
	code, payload = testRequest(t, ts, "POST", "/update/unsupported/testUnsupported/100", "")
	assert.Equal(t, http.StatusNotImplemented, code)
	assert.Equal(t, http.StatusText(http.StatusNotImplemented), payload)

	// test save valid metric counter
	code, _ = testRequest(t, ts, "POST", "/update/counter/testCounter/100", "")
	assert.Equal(t, http.StatusOK, code)

	// test save valid metric gauge
	code, _ = testRequest(t, ts, "POST", "/update/gauge/testGauge/10.0", "")
	assert.Equal(t, http.StatusOK, code)

	// test save unsupported metric type JSON
	code, _ = testRequest(t, ts, "POST", "/update", `{"id": "testJSONGauge", "type": "unsupported", "value": 1}`)
	assert.Equal(t, http.StatusNotImplemented, code)

	// test save valid metric counter JSON
	code, _ = testRequest(t, ts, "POST", "/update", `{"id": "testJSONCounter", "type": "counter", "delta": 110}`)
	assert.Equal(t, http.StatusOK, code)

	// test save valid metric gauge JSON
	code, _ = testRequest(t, ts, "POST", "/update", `{"id": "testJSONGauge", "type": "gauge", "value": 11.0}`)
	assert.Equal(t, http.StatusOK, code)

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
	code, _ = testRequest(t, ts, "POST", "/updates/", unsupportedMetrics)
	assert.Equal(t, http.StatusNotImplemented, code)

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
	code, _ = testRequest(t, ts, "POST", "/updates/", metrics)
	assert.Equal(t, http.StatusOK, code)

	// test get not existing counter plain
	code, payload = testRequest(t, ts, "GET", "/value/counter/testNotExist", "")
	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, http.StatusText(http.StatusNotFound), payload)

	// test get counter plain
	code, payload = testRequest(t, ts, "GET", "/value/counter/testCounter", "")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "100", payload)

	// test get not existing gauge JSON
	notExistingJSONMetric := `
		{
		  "id": "testNotExistingJSON",
		  "type": "gauge"
		}
	`
	code, _ = testRequest(t, ts, "POST", "/value/", notExistingJSONMetric)
	assert.Equal(t, http.StatusNotFound, code)
	// test get gauge JSON
	jsonMetric := `
		{
		  "id": "testJSONGauge",
		  "type": "gauge"
		}
	`
	code, payload = testRequest(t, ts, "POST", "/value/", jsonMetric)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, `{"id":"testJSONGauge","type":"gauge","value":11}`, payload)
}
