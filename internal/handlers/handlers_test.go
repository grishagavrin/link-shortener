package handlers_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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
				response:    `{"id":0,"address":"https://dzen.ru"}`,
				contentType: "application/json",
				value:       "",
			},
		},
	}

	for _, tt := range postTests {
		t.Run("POST create short URL in DB (positive)", func(t *testing.T) {
			data := url.Values{}
			data.Set("url", "https://dzen.ru")
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(data.Encode()))
			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

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
			request := httptest.NewRequest(http.MethodGet, "/?id=0", nil)
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

			if !strings.Contains(string(resBody), "https://dzen.ru") {
				t.Errorf("Expected body %s, got:  %s", "https://dzen.ru", w.Body.String())
			}

			if !strings.Contains(res.Header.Get("Content-Type"), "text/html") {
				t.Errorf("Expected Content-Type %s, got %s", "text/html", res.Header.Get("Content-Type"))
			}
		})
	}
}
