package routes

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/grishagavrin/link-shortener/internal/handlers"
)

func ServiceRouter() chi.Router {
	r := chi.NewRouter()
	// cfg := config.GetENV()

	r.Use(middleware.Recoverer)
	r.Get("/{id}", handlers.GetLink)
	r.Post("/", handlers.AddLink)
	// r.Post(cfg.BaseURL, handlers.ShortenURL)
	r.Post("/api/shorten", handlers.ShortenURL)
	return r
}
