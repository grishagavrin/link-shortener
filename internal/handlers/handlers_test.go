package handlers_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grishagavrin/link-shortener/internal/handlers"
)

func TestCommonHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
		value       string
	}

	postTests := []struct {
		want want
	}{
		{
			want: want{
				code:        201,
				response:    "http://localhost:8080/0",
				contentType: "",
				value:       "http://yandex.ru",
			},
		},
	}

	for _, tt := range postTests {
		t.Run("POST create short URL in DB (positive)", func(t *testing.T) {

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.want.value))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(handlers.CommonHandler)

			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatal(err)
			}

			if string(resBody) != tt.want.response {
				t.Errorf("Expected body %s, got:  %s", tt.want.response, w.Body.String())
			}

			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})

		t.Run("GET short URL by id (positive)", func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/0", nil)
			request.Header.Add("Content-Type", tt.want.contentType)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(handlers.CommonHandler)

			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != 307 {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(string(resBody), tt.want.value) {
				t.Errorf("Expected body %s, got:  %s", tt.want.value, w.Body.String())
			}

			if !strings.Contains(res.Header.Get("Content-Type"), "text/html") {
				t.Errorf("Expected Content-Type %s, got %s", "text/html", res.Header.Get("Content-Type"))
			}
		})
	}
}
