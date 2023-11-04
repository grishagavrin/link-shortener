package routes

import (
	"github.com/grishagavrin/link-shortener/internal/handlers"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
	"go.uber.org/zap"
)

// RouterFacade struct for return constructor
type RouterFacade struct {
	HTTPRoute HTTPRoute
}

// NewRouterFacade for return instance
func NewRouterFacade(
	h *handlers.Handler,
	l *zap.Logger,
	chBatch chan models.BatchDelete,
) *RouterFacade {
	return &RouterFacade{
		HTTPRoute: NewHTTPRouter(h, l, chBatch),
	}
}
