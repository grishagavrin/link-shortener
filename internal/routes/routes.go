package routes

import (
	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/handlers"
	"github.com/grishagavrin/link-shortener/internal/handlers/middlewares"

	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

func ServiceRouter(l *zap.Logger) chi.Router {
	r := chi.NewRouter()
	h, err := handlers.New(l)
	if err != nil {
		l.Fatal("get instance ram/db error: ", zap.Error(err))
	}
	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.CooksMiddleware)

	r.Get("/{id}", h.GetLink)
	r.Post("/", h.SaveTXT)
	r.Post("/api/shorten", h.SaveJSON)
	r.Get("/api/user/urls", h.GetLinks)
	r.Get("/ping", h.GetPing)
	r.Post("/api/shorten/batch", h.SaveBatch)
	r.Delete("/api/user/urls", h.DeleteBatch)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://127.0.0.1:8080/swagger/doc.json"), //The url pointing to API definition
	))
	return r
}
