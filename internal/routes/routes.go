package routes

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/grishagavrin/link-shortener/internal/handlers"
)

func ServiceRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(handlers.GzipMiddleware)
	r.Get("/{id}", handlers.GetLink)
	r.Post("/", handlers.SaveTXT)
	r.Post("/api/shorten", handlers.SaveJSON)

	return r
}
