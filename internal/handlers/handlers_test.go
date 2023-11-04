package handlers_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grishagavrin/link-shortener/internal/handlers"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHandler_GetLink(t *testing.T) {
	chBatch := make(chan models.BatchDelete)
	defer close(chBatch)
	// создаём новый Recorder
	w := httptest.NewRecorder()
	// создаем логер
	l, _ := logger.Instance()
	// создаем хранение
	stor, _ := storage.Instance(l, chBatch)
	// создаем handler
	h := handlers.New(stor.Repository, l)
	// создаем роутер
	r := routes.NewRouterFacade(h, l, chBatch)
	// создаем сервер
	ts := httptest.NewServer(r.HTTPRoute.Route)
	defer ts.Close()

	// определяем структуру теста
	type want struct {
		code        int
		response    string
		contentType string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name        string
		want        want
		queryString string
	}{
		// определяем все тесты
		{
			name:        "negative test #1",
			queryString: "/e409bf5aafc88512",
			want: want{
				code:        400,
				response:    "bad request\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:        "negative test #2",
			queryString: "/e409bf5aafc",
			want: want{
				code:        400,
				response:    "enter correct url parameter\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			// создаем запрос
			req, _ := http.NewRequest(http.MethodGet, ts.URL+tt.queryString, nil)
			// делаем запрос
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				l.Fatal("TestGetLinkHandler", zap.Error(err))
			}
			defer res.Body.Close()

			respBody, _ := io.ReadAll(res.Body)

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			// заголовок ответа
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}

			// тело ответа
			if eqBody := assert.Equal(t, tt.want.response, string(respBody)); !eqBody {
				t.Errorf("Expected Body %s, got %s", tt.want.response, string(respBody))
			}

		})
	}
}

func TestHandler_SaveTXT(t *testing.T) {
	chBatch := make(chan models.BatchDelete)
	defer close(chBatch)
	// создаем логер
	l, _ := logger.Instance()
	// создаем хранение
	stor, _ := storage.Instance(l, chBatch)
	// создаем handler
	h := handlers.New(stor.Repository, l)
	// создаем роутер
	r := routes.NewRouterFacade(h, l, chBatch)
	// создаем сервер
	ts := httptest.NewServer(r.HTTPRoute.Route)
	defer ts.Close()

	// определяем структуру теста
	type want struct {
		code        int
		contentType string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name        string
		want        want
		queryParam  string
		requestBody string
		queryString string
	}{
		// определяем все тесты
		{
			name: "positive test #1",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain; charset=utf-8",
			},
			requestBody: "http://yandex.ru",
			queryString: "/",
		},
		{
			name: "positive test #2",
			want: want{
				code:        http.StatusConflict,
				contentType: "text/plain; charset=utf-8",
			},
			requestBody: "http://yandex.ru",
			queryString: "/",
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// создаем запрос
			req, _ := http.NewRequest(http.MethodPost, ts.URL+tt.queryString, bytes.NewBufferString(tt.requestBody))
			// делаем запрос
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				l.Fatal("TestSaveTXTHandler", zap.Error(err))
			}
			defer res.Body.Close()

			resBody, _ := io.ReadAll(res.Body)

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				if res.StatusCode != http.StatusCreated {
					if res.StatusCode != http.StatusConflict {
						t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
					}
				}
			}

			// заголовок ответа
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}

			// тело ответа
			if body := assert.NotNil(t, resBody); !body {
				t.Error("Expected not nil Body, got nil")
			}

		})
	}
}

func TestHandler_SaveJSON(t *testing.T) {
	chBatch := make(chan models.BatchDelete)
	// создаем логер
	l, _ := logger.Instance()
	// создаем хранение
	stor, _ := storage.Instance(l, chBatch)
	// создаем handler
	h := handlers.New(stor.Repository, l)
	// создаем роутер
	r := routes.NewRouterFacade(h, l, chBatch)
	// создаем сервер
	ts := httptest.NewServer(r.HTTPRoute.Route)
	defer ts.Close()
	defer close(chBatch)

	// определяем структуру теста
	type want struct {
		code        int
		contentType string
		response    string
	}

	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name        string
		want        want
		requestBody string
		queryString string
	}{
		// определяем все тесты
		{
			name: "negative test #1",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				response:    "invalid fields in json: json: unknown field \"url1\"\n",
			},
			requestBody: `"url1":"http://yandex.ru"`,
			queryString: "/api/shorten",
		},
		{
			name: "positive test #1",
			want: want{
				code:        http.StatusConflict,
				contentType: "application/json",
				response:    "localhost:8080",
			},
			requestBody: `"url":"http://yandex.ru"`,
			queryString: "/api/shorten",
		},
	}
	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			// создаём новый Recorder
			w := httptest.NewRecorder()
			// создаем тело
			var jsonData = []byte(fmt.Sprintf("{%v}", tt.requestBody))
			// создаем запрос
			req, _ := http.NewRequest(http.MethodPost, ts.URL+tt.queryString, bytes.NewBuffer(jsonData))
			// делаем запрос
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				l.Fatal("TestSaveJSONHandler", zap.Error(err))
			}
			defer res.Body.Close()

			resBody, _ := io.ReadAll(res.Body)

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
				if res.StatusCode != http.StatusCreated {
					if res.StatusCode != http.StatusConflict {
						t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
					}
				}
			}

			// заголовок ответа
			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}

			// тело ответа
			body := assert.NotNil(t, resBody)
			if !body {
				t.Error("Expected not nil Body, got nil")
			}

			// содержание тела ответа
			containsBody := strings.Contains(string(resBody), tt.want.response)
			if !containsBody {
				t.Errorf("Expected Body %s, got %s", tt.want.response, string(resBody))
			}
		})
	}
}
