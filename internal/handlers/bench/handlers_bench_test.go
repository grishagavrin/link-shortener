package bench

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/handlers"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
)

func BenchmarkHandler_SaveTXT(b *testing.B) {
	var r io.Reader

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
	routes := routes.NewRouterFacade(h, l, chBatch)

	rtr := chi.NewRouter()

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		st := "http://yandex" + strconv.Itoa(i) + ".ru"
		r = strings.NewReader(st)
		request := httptest.NewRequest(http.MethodPost, "/", r)

		b.StartTimer()
		routes.HTTPRoute.Route.HandleFunc("/", h.SaveTXT)

		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()
		b.StopTimer() // останавливаем таймер
		res.Body.Close()
	}
}

func BenchmarkHandler_SaveJSON(b *testing.B) {
	var r io.Reader

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
	routes := routes.NewRouterFacade(h, l, chBatch)

	rtr := chi.NewRouter()

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		st := "{\"url\": \"http://yandex.ru" + strconv.Itoa(i) + ".ru\"}"
		r = strings.NewReader(st)
		request := httptest.NewRequest(http.MethodPost, "/", r)

		b.StartTimer() //
		routes.HTTPRoute.Route.HandleFunc("/", h.SaveJSON)
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		b.StopTimer() // останавливаем таймер

		res.Body.Close()
	}
}

func BenchmarkHandler_SaveBatch(b *testing.B) {
	var r io.Reader

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
	routes := routes.NewRouterFacade(h, l, chBatch)

	rtr := chi.NewRouter()

	b.ResetTimer() // reset all timers
	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		st := "{\"original_url\": \"http://yandex" + strconv.Itoa(i) + ".ru\",\"correlation_id\": \"" + strconv.Itoa(i) + "\"}"
		r = strings.NewReader(st)
		request := httptest.NewRequest(http.MethodPost, "/", r)

		b.StartTimer() //
		routes.HTTPRoute.Route.HandleFunc("/", h.SaveJSON)
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		b.StopTimer() // останавливаем таймер

		res.Body.Close()
	}
}

func BenchmarkHandler_GetUrls(b *testing.B) {
	var r io.Reader

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
	routes := routes.NewRouterFacade(h, l, chBatch)

	rtr := chi.NewRouter()

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers

		request := httptest.NewRequest(http.MethodGet, "/user/urls", r)

		b.StartTimer() //
		routes.HTTPRoute.Route.HandleFunc("/user/urls", h.GetLinks)
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		res.Body.Close()
	}
}
