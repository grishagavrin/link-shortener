package handlers_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (int, string) {
	var req *http.Request
	var err error

	if method == "POST" {
		req, err = http.NewRequest(method, ts.URL+path, bytes.NewBufferString(body))
	} else if method == "POSTJSON" {
		var jsonData = []byte(fmt.Sprintf("{%v}", body))
		req, err = http.NewRequest("POST", ts.URL+path, bytes.NewBuffer(jsonData))
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

func TestShortenURL(t *testing.T) {
	r := routes.ServiceRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, body := testRequest(t, ts, "POSTJSON", "/api/shorten", `"url1":"http://yandex.ru"`)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, "invalid fields in JSON\n", body)

	statusCode, _ = testRequest(t, ts, "POSTJSON", "/api/shorten", `"url":"http://yandex.ru"`)
	assert.Equal(t, http.StatusCreated, statusCode)

}

func TestWriteURL(t *testing.T) {
	r := routes.ServiceRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, body := testRequest(t, ts, "POST", "/", "http://yandex.ru")
	_ = body
	assert.Equal(t, http.StatusCreated, statusCode)

	statusCode, body = testRequest(t, ts, "POST", "/", "")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, "body is empty\n", body)
}

func TestGetURL(t *testing.T) {
	r := routes.ServiceRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, body := testRequest(t, ts, "POST", "/", "http://yandex.ru")
	assert.Equal(t, http.StatusCreated, statusCode)

	statusCode, body = testRequest(t, ts, "GET", "/"+body[len(body)-config.LENHASH:], "")
	assert.Equal(t, http.StatusOK, statusCode)
	assert.NotNil(t, body)

	statusCode, body = testRequest(t, ts, "GET", "/aaa", "")
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, fmt.Sprintf("enter correct url parameter - length: %v\n", config.LENHASH), body)
}
