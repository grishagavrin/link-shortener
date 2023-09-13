// Package delete implement handler for delete links for route /api/user/urls
package delete

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/handlers/middlewares"
	istorage "github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"go.uber.org/zap"
)

// Handler struct for delete batch
type Handler struct {
	l       *zap.Logger
	chBatch chan istorage.BatchDelete
}

// New instance of deleted handler
func New(l *zap.Logger, chBatch chan istorage.BatchDelete) *Handler {
	return &Handler{
		l,
		chBatch,
	}
}

// DeleteBatch godoc
// @Tags DeleteBatch
// @Summary Delete handler with fan in channel
// @Failure 500 {string} string "internal error"
// @Success 200 {array} array
// @Router /api/user/urls [delete]
// Delete handler with fan in channel
func (h Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var correlationIDs []string
	userID := middlewares.GetContextUserID(req)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &correlationIDs)
	if err != nil {
		http.Error(res, errs.ErrCorrectURL.Error(), http.StatusBadRequest)
		return
	}

	// Validate count
	if len(correlationIDs) == 0 {
		http.Error(res, errs.ErrCorrectURL.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusAccepted)

	chStruct := istorage.BatchDelete{
		UserID: string(userID),
		URLs:   correlationIDs,
	}

	go func() {
		h.l.Info("new chStruct")
		h.chBatch <- chStruct
	}()
}
