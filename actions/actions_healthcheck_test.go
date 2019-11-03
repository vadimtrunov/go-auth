package actions

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthcheck(t *testing.T) {
	assert := assert.New(t)
	status, response := Healthcheck(nil)

	assert.Equal(status, http.StatusOK, "Invalid response status code")

	data := response.(healthCheckResponse)

	assert.Equal(data.Status, "Ok", "Invalid response body. Invalid field 'status'")
	assert.Equal(data.Code, http.StatusOK, "Invalid response body. Invalid field 'code'")
}
