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
	// создаем handler
	h := handlers.New(stor.Repository, l)
	// создаем роутер
	r := routes.NewRouterFacade(h, l, chBatch)
	// create server
	ts := httptest.NewServer(r.HTTPRoute.Route)
	defer ts.Close()

	body := strings.NewReader("http://yandex.ru")
	request := httptest.NewRequest(http.MethodPost, "/", body)

	r.HTTPRoute.Route.HandleFunc("/", h.SaveTXT)
	r.HTTPRoute.Route.ServeHTTP(w, request)
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
	// создаем handler
	h := handlers.New(stor.Repository, l)
	// создаем роутер
	r := routes.NewRouterFacade(h, l, chBatch)
	// create server
	ts := httptest.NewServer(r.HTTPRoute.Route)
	defer ts.Close()

	body := strings.NewReader("{\"url\": \"http://yandex.ru\"}")
	request := httptest.NewRequest(http.MethodPost, "/", body)

	r.HTTPRoute.Route.HandleFunc("/", h.SaveJSON)
	r.HTTPRoute.Route.ServeHTTP(w, request)
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
	// создаем handler
	h := handlers.New(stor.Repository, l)
	// создаем роутер
	r := routes.NewRouterFacade(h, l, chBatch)
	// create server
	ts := httptest.NewServer(r.HTTPRoute.Route)
	defer ts.Close()

	body := strings.NewReader("")
	request := httptest.NewRequest(http.MethodGet, "/e409bf5aafc88512", body)

	r.HTTPRoute.Route.HandleFunc("/{id:.+}", h.GetLink)
	r.HTTPRoute.Route.ServeHTTP(w, request)
	res := w.Result()

	defer res.Body.Close()
}
