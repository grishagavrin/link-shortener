package handlers_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (int, string) {
	var req *http.Request
	var err error

	if method == "POST" {
		req, err = http.NewRequest(method, ts.URL+path, bytes.NewBufferString(body))
	} else if method == "GET" {
		req, err = http.NewRequest(method, ts.URL+path, nil)
	}

	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestWriteURL(t *testing.T) {
	r := routes.ServiceRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, body := testRequest(t, ts, "POST", "/", "http://yandex.ru")
	assert.Equal(t, http.StatusCreated, statusCode)
	assert.Equal(t, "http://localhost:8080/0", body)

	statusCode, body = testRequest(t, ts, "POST", "/", "")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, "Body is empty\n", body)
}

func TestGetURL(t *testing.T) {
	r := routes.ServiceRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, body := testRequest(t, ts, "GET", "/0", "")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.NotNil(t, body)

	statusCode, body = testRequest(t, ts, "GET", "/6", "")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, "The id parametr not found in DB\n", body)

	statusCode, body = testRequest(t, ts, "GET", "/aaa", "")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, "Enter a number type parameter\n", body)
}
