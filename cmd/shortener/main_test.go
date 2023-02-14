package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostRouting(t *testing.T) {

	srv := httptest.NewServer(MyHandler())
	defer srv.Close()

	//POST
	res, err := http.Post(srv.URL, "text/plain", bytes.NewBufferString("http://yandex.ru"))

	if err != nil {
		t.Fatalf("could not send POST request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201 - Created; got %v", res.Status)
	}
}

func TestGetRouting(t *testing.T) {

	srv := httptest.NewServer(MyHandler())
	defer srv.Close()

	//POST
	res, err := http.Get(fmt.Sprintf("%s/0", srv.URL))

	if err != nil {
		t.Fatalf("could not send GET request: %v", err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)

	bString := string(bytes.TrimSpace(b))
	if bString == "" {
		t.Fatalf("could not read response: %v", bString)
	}

	if err != nil {
		t.Fatalf("could not read response: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200; got %v", res.Status)
	}
}
