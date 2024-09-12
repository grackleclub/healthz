package healthz

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
