package routes

import (
	"log"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/handlers"
	"github.com/grishagavrin/link-shortener/internal/handlers/middlewares"
)

func ServiceRouter() chi.Router {
	r := chi.NewRouter()
	h, err := handlers.New()
	if err != nil {
		log.Fatal("get instance db error")
	}

	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.CooksMiddleware)
	r.Get("/{id}", h.GetLink)
	r.Post("/", h.SaveTXT)
	r.Post("/api/shorten", h.SaveJSON)
	r.Get("/api/user/urls", h.GetLinks)
	r.Get("/ping", h.GetPing)
	return r
}
