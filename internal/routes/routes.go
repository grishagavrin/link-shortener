package routes

import (
	"log"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/grishagavrin/link-shortener/internal/handlers"
)

func ServiceRouter() chi.Router {
	r := chi.NewRouter()
	h, err := handlers.New()
	if err != nil {
		log.Fatal("get instance db error")
	}

	r.Use(middleware.Recoverer)
	r.Use(handlers.GzipMiddleware)
	r.Get("/{id}", h.GetLink)
	r.Post("/", h.SaveTXT)
	r.Post("/api/shorten", h.SaveJSON)
	return r
}
