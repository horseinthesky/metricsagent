package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go testServer.Run(ctx)

	time.Sleep(2 * time.Second)

	cancel()
	testServer.Stop()
}

func TestRouter(t *testing.T) {
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

func TestDashBoard(t *testing.T) {
	ts := httptest.NewServer(testServer)
	defer ts.Close()

	code, body := testRequest(t, ts, http.MethodGet, "/", "")
	require.Equal(t, http.StatusOK, code)
	require.NotEmpty(t, body)
}
