package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrustedSubnet(t *testing.T) {
	ts := httptest.NewServer(testTrustedServer)
	defer ts.Close()

	// trusted request
	trustedReq, err := http.NewRequest(http.MethodPost, ts.URL+"/update/counter/trustedCounter/100", nil)
	require.NoError(t, err)

	trustedReq.Header.Add("X-Real-IP", "10.10.10.10")

	resp, err := http.DefaultClient.Do(trustedReq)
	require.NoError(t, err)

	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// no ip request
	noIPReq, err := http.NewRequest(http.MethodPost, ts.URL+"/update/counter/noIPCounter/100", nil)
	require.NoError(t, err)

	resp, err = http.DefaultClient.Do(noIPReq)
	require.NoError(t, err)

	resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	// untrusted request
	untrustedReq, err := http.NewRequest(http.MethodPost, ts.URL+"/update/counter/untrustedCounter/100", nil)
	require.NoError(t, err)

	untrustedReq.Header.Add("X-Real-IP", "10.10.20.10")

	resp, err = http.DefaultClient.Do(untrustedReq)
	require.NoError(t, err)

	resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
