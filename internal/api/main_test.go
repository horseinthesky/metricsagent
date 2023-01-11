package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/horseinthesky/metricsagent/internal/server"
	"github.com/stretchr/testify/require"
)

var (
	testServer        *Server
	testHashedServer  *Server
	testTrustedServer *Server
)

func init() {
	testServer, _ = NewServer(server.Config{
		Address:       "localhost:8080",
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
	})

	testHashedServer, _ = NewServer(server.Config{
		Address:       "localhost:8085",
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
		Key:           "testkey",
	})

	testTrustedServer, _ = NewServer(server.Config{
		Address:       "localhost:8090",
		Restore:       false,
		StoreInterval: 10 * time.Minute,
		StoreFile:     "/tmp/test-metrics-db.json",
		TrustedSubnet: "10.10.10.0/24",
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
