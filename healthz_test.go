package healthz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespond(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthz", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Respond)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")

	h := Healthz{}
	json.Unmarshal(rr.Body.Bytes(), &h)
	t.Logf("Healthz check completed: %+v", h)

	// TODO use reflection on Healthz struct to check for fields, not hard code
	expectedFields := []string{"uptime", "version", "cpu", "memory", "disk"}
	for _, field := range expectedFields {
		assert.Contains(t, rr.Body.String(), field, "Expected field %s in response", field)
	}
}

func TestPing(t *testing.T) {
	now := time.Now().Unix()
	health := Healthz{
		Time:    int(now),
		Status:  200,
		Uptime:  "1m",
		Version: "v1.2.3-abcd",
		CPU:     ".101",
		Memory:  ".2020",
		Disk:    ".303030",
	}
	// Create a test server that returns a predefined response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(health)
	}))
	defer ts.Close()

	h, err := Ping(ts.URL)
	require.NoError(t, err)
	require.Equal(t, h, health)
	t.Logf("Ping check completed: %+v", h)

	h, err = PingWithRetry(ts.URL, 3)
	require.NoError(t, err)
	require.Equal(t, h, health)
	t.Logf("Ping check with retries completed: %+v", h)

	// turn h.Time into a time.Time object
	newTime := time.Unix(int64(h.Time), 0)
	t.Logf("time: %v", newTime)
	require.Equal(t, now, newTime.Unix())
}
