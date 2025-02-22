package healthz

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// tests the handler and the ping with retry together
func TestPingAndRespond(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Respond))
	defer ts.Close()
	h, err := PingWithRetry(ts.URL, 3)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, h.Status, "Expected status OK")
}
