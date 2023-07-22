package routes

import (
	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/handlers"
	"github.com/grishagavrin/link-shortener/internal/handlers/middlewares"
	"github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"go.uber.org/zap"
)

func ServiceRouter(stor iStorage.Repository, l *zap.Logger) chi.Router {
	r := chi.NewRouter()
	h := handlers.New(stor, l)

	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.CooksMiddleware)
	r.Get("/{id}", h.GetLink)
	r.Post("/", h.SaveTXT)
	r.Post("/api/shorten", h.SaveJSON)
	r.Get("/api/user/urls", h.GetLinks)
	r.Get("/ping", h.GetPing)
	r.Post("/api/shorten/batch", h.SaveBatch)
	r.Delete("/api/user/urls", h.DeleteBatch)

	return r
}
