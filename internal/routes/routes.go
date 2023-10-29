// Package routes contains general routes for service link-shortener
package routes

import (
	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/handlers"
	"github.com/grishagavrin/link-shortener/internal/handlers/delete"
	"github.com/grishagavrin/link-shortener/internal/handlers/middlewares"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
	"go.uber.org/zap"
)

// ServiceRouter define routes in server
func ServiceRouter(stor handlers.Repository, l *zap.Logger, chBatch chan models.BatchDelete) chi.Router {
	r := chi.NewRouter()
	h := handlers.New(stor, l)

	// Middlewares
	r.Use(middlewares.GzipMiddleware)
	r.Use(middlewares.CooksMiddleware)
	// Handlers
	r.Get("/{id}", h.GetLink)
	r.Post("/", h.SaveTXT)
	r.Post("/api/shorten", h.SaveJSON)
	r.Get("/api/user/urls", h.GetLinks)
	r.Get("/ping", h.GetPing)
	r.Post("/api/shorten/batch", h.SaveBatch)
	r.Delete("/api/user/urls", delete.New(l, chBatch).ServeHTTP)
	r.Get("/api/internal/stats", h.GetStats)

	return r
}
