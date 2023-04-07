package routes

import (
	"fmt"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/handlers"
)

func ServiceRouter() chi.Router {
	r := chi.NewRouter()
	h, err := handlers.New()
	if err != nil {
		fmt.Printf("get instance db error: %s", err.Error())
	}

	// r.Use(middleware.Recoverer)
	// r.Use(middlewares.GzipMiddleware)
	// r.Use(middlewares.CooksMiddleware)
	r.Get("/{id}", h.GetLink)
	r.Get("/user/urls", h.GetLinks)
	r.Post("/", h.SaveTXT)
	r.Post("/api/shorten", h.SaveJSON)
	return r
}
