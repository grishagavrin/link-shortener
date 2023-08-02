package handlers_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"log"
// 	"math/rand"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/grishagavrin/link-shortener/internal/config"
// 	"github.com/grishagavrin/link-shortener/internal/logger"
// 	"github.com/grishagavrin/link-shortener/internal/routes"
// 	"github.com/grishagavrin/link-shortener/internal/storage"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func testRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (int, string) {
// 	var req *http.Request
// 	var err error

// 	if method == "POST" {
// 		req, err = http.NewRequest(method, ts.URL+path, bytes.NewBufferString(body))
// 	} else if method == "POSTJSON" {
// 		var jsonData = []byte(fmt.Sprintf("{%v}", body))
// 		req, err = http.NewRequest("POST", ts.URL+path, bytes.NewBuffer(jsonData))
// 	} else if method == "GET" {
// 		req, err = http.NewRequest(method, ts.URL+path, nil)
// 	}

// 	require.NoError(t, err)

// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(t, err)

// 	respBody, err := io.ReadAll(resp.Body)
// 	require.NoError(t, err)

// 	defer resp.Body.Close()

// 	return resp.StatusCode, string(respBody)
// }

// func TestShortenURL(t *testing.T) {
// 	l, err := logger.Instance()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	stor, _, _ := storage.Instance(l)
// 	r := routes.ServiceRouter(stor, l)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()

// 	statusCode, body := testRequest(t, ts, "POSTJSON", "/api/shorten", `"url1":"http://yandex.ru"`)
// 	assert.Equal(t, http.StatusBadRequest, statusCode)
// 	assert.Equal(t, "invalid fields in json: json: unknown field \"url1\"\n", body)

// 	randomInt := rand.Intn(100000-0) + 44

// 	statusCode, _ = testRequest(t, ts, "POSTJSON", "/api/shorten", fmt.Sprintf(`"url": "http://yandex%d.ru"`, randomInt))
// 	assert.Equal(t, http.StatusCreated, statusCode)

// }

// func TestWriteURL(t *testing.T) {
// 	l, err := logger.Instance()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	stor, _, _ := storage.Instance(l)
// 	r := routes.ServiceRouter(stor, l)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()

// 	randomInt := rand.Intn(100000-0) + 44

// 	statusCode, body := testRequest(t, ts, "POST", "/", fmt.Sprintf("http://yandex%d.ru", randomInt))
// 	_ = body
// 	assert.Equal(t, http.StatusCreated, statusCode)

// 	statusCode, body = testRequest(t, ts, "POST", "/", "")
// 	assert.Equal(t, http.StatusBadRequest, statusCode)
// 	assert.Equal(t, "body is empty\n", body)
// }

// func TestGetURL(t *testing.T) {
// 	l, err := logger.Instance()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	stor, _, _ := storage.Instance(l)
// 	r := routes.ServiceRouter(stor, l)
// 	ts := httptest.NewServer(r)
// 	defer ts.Close()

// 	statusCode, body := testRequest(t, ts, "POST", "/", "http://yandex.ru")
// 	assert.Equal(t, http.StatusCreated, statusCode)

// 	statusCode, body = testRequest(t, ts, "GET", "/"+body[len(body)-config.LENHASH:], "")
// assert.Equal(t, http.StatusOK, statusCode)
// assert.NotNil(t, body)

// 	statusCode, body = testRequest(t, ts, "GET", "/aaa", "")
// 	assert.Equal(t, http.StatusBadRequest, statusCode)
// 	assert.Equal(t, "enter correct url parameter\n", body)
// }

func TestGetLinkHandler(t *testing.T) {
	// создаём новый Recorder
	w := httptest.NewRecorder()
	// создаем логер
	l, _ := logger.Instance()
	// создаем хранение
	stor, _, _ := storage.Instance(l)
	// создаем роутер
	r := routes.ServiceRouter(stor, l)
	// создаем сервер
	ts := httptest.NewServer(r)
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
			//делаем запрос
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

func TestSaveTXTHandler(t *testing.T) {
	// создаем логер
	l, _ := logger.Instance()
	// создаем хранение
	stor, _, _ := storage.Instance(l)
	// создаем роутер
	r := routes.ServiceRouter(stor, l)
	// создаем сервер
	ts := httptest.NewServer(r)
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
			//делаем запрос
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

func TestSaveJSONHandler(t *testing.T) {
	// создаем логер
	l, _ := logger.Instance()
	// создаем хранение
	stor, _, _ := storage.Instance(l)
	// создаем роутер
	r := routes.ServiceRouter(stor, l)
	// создаем сервер
	ts := httptest.NewServer(r)
	defer ts.Close()

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
			//делаем запрос
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
