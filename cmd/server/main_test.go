package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPostRouting(t *testing.T) {

	srv := httptest.NewServer(MyHandler())
	defer srv.Close()

	data := url.Values{}
	data.Set("url", "https://dzen.ru")
	res, err := http.Post(srv.URL, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))

	if err != nil {
		t.Errorf("could not send POST request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected status OK; got %v", res.Status)
	}

}
