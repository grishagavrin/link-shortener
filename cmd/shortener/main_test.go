package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBufferString(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}

func TestServerRun(t *testing.T) {
	l, err := logger.Instance()
	if err != nil {
		log.Fatal(err)
	}
	r := routes.ServiceRouter(l)
	ts := httptest.NewServer(r)
	defer ts.Close()

	statusCode, body := testRequest(t, ts, "POST", "/", "http://yandex.ru")
	assert.Equal(t, http.StatusCreated, statusCode)

	statusCode, _ = testRequest(t, ts, "GET", "/"+body[len(body)-config.LENHASH:], "")
	assert.Equal(t, http.StatusOK, statusCode)

	statusCode, body = testRequest(t, ts, "POST", "/api/shorten", "")
	assert.Equal(t, "invalid fields in json\n", body)
	assert.Equal(t, http.StatusBadRequest, statusCode)
}
