// go test -bench=BenchmarkHandler_SaveTXT -benchmem -benchtime=2500x -memprofile base.pprof // Run one bench
// go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof // See difference
// go test -bench=. -memprofile=base.pprof // Run all benchs in handlers
// go tool pprof -http=":9090" handlers.test base.pprof // See profile
// go tool pprof -http=":9090" handlers.test result.pprof // See result profile
package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/logger"
)

func BenchmarkHandler_SaveTXT(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	l, _ := logger.Instance()
	rtr := chi.NewRouter()
	h, _ := New(l)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		st := "http://testlink" + strconv.Itoa(i) + ".ru"
		r = strings.NewReader(st)
		request := httptest.NewRequest(http.MethodPost, "/", r)

		b.StartTimer()
		rtr.HandleFunc("/", h.SaveTXT)
		rtr.ServeHTTP(w, request)
		res := w.Result()
		b.StopTimer()

		res.Body.Close()
	}
}

func BenchmarkHandler_SaveJSON(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	l, _ := logger.Instance()
	rtr := chi.NewRouter()
	h, _ := New(l)

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		st := "{\"url\": \"http://testlink" + strconv.Itoa(i) + ".ru\"}"
		r = strings.NewReader(st)
		request := httptest.NewRequest(http.MethodPost, "/", r)

		b.StartTimer() //
		rtr.HandleFunc("/", h.SaveJSON)
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()

		b.StopTimer() // останавливаем таймер

		res.Body.Close()
	}
}

func BenchmarkHandler_GetLinks(b *testing.B) {
	var r io.Reader
	w := httptest.NewRecorder()
	l, _ := logger.Instance()
	rtr := chi.NewRouter()
	h, _ := New(l)

	b.ResetTimer() // reset all timers

	for i := 0; i < b.N; i++ {
		b.StopTimer() // stop all timers
		request := httptest.NewRequest(http.MethodGet, "/user/urls", r)
		b.StartTimer() //
		rtr.HandleFunc("/user/urls", h.GetLinks)
		// запускаем сервер
		rtr.ServeHTTP(w, request)
		res := w.Result()
		res.Body.Close()
	}
}
