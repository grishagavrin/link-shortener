package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/grishagavrin/link-shortener/internal/handlers"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
)

func ExampleHandler_SaveTXT() {
	chBatch := make(chan models.BatchDelete)
	defer close(chBatch)
	w := httptest.NewRecorder()
	// create logger
	l, _ := logger.Instance()
	// create storage
	stor, _ := storage.Instance(l, chBatch)
	// create router
	r := routes.ServiceRouter(stor.Repository, l, chBatch)
	// handlers
	h := handlers.New(stor.Repository, l)
	// create server
	ts := httptest.NewServer(r)
	defer ts.Close()

	body := strings.NewReader("http://yandex.ru")
	request := httptest.NewRequest(http.MethodPost, "/", body)

	r.HandleFunc("/", h.SaveTXT)
	r.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()
}

func ExampleHandler_SaveJSON() {
	chBatch := make(chan models.BatchDelete)
	defer close(chBatch)
	w := httptest.NewRecorder()
	// create logger
	l, _ := logger.Instance()
	// create storage
	stor, _ := storage.Instance(l, chBatch)
	// create router
	r := routes.ServiceRouter(stor.Repository, l, chBatch)
	// handlers
	h := handlers.New(stor.Repository, l)
	// create server
	ts := httptest.NewServer(r)
	defer ts.Close()

	body := strings.NewReader("{\"url\": \"http://yandex.ru\"}")
	request := httptest.NewRequest(http.MethodPost, "/", body)

	r.HandleFunc("/", h.SaveJSON)
	r.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()
}

func ExampleHandler_GetLink() {
	chBatch := make(chan models.BatchDelete)
	defer close(chBatch)
	w := httptest.NewRecorder()
	// create logger
	l, _ := logger.Instance()
	// create storage
	stor, _ := storage.Instance(l, chBatch)
	// create router
	r := routes.ServiceRouter(stor.Repository, l, chBatch)
	// handlers
	h := handlers.New(stor.Repository, l)
	// create server
	ts := httptest.NewServer(r)
	defer ts.Close()

	body := strings.NewReader("")
	request := httptest.NewRequest(http.MethodGet, "/e409bf5aafc88512", body)

	r.HandleFunc("/{id:.+}", h.GetLink)
	r.ServeHTTP(w, request)
	res := w.Result()

	defer res.Body.Close()
}
