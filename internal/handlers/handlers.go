// Package handlers includes general handlers for service shortener
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/handlers/middlewares"
	"github.com/grishagavrin/link-shortener/internal/storage"
	istorage "github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"go.uber.org/zap"
)

// Handler general type fo handler
type Handler struct {
	s istorage.Repository
	l *zap.Logger
}

// New allocation new handler
func New(stor istorage.Repository, l *zap.Logger) *Handler {
	return &Handler{
		s: stor,
		l: l,
	}
}

// GetLink godoc
// @Tags GetLink
// @Summary Request to get the original link
// @Param id path string true "2dace3f162eb9f0d"
// @Failure 400 {string} string "bad request"
// @Success 200 {string} string
// @Router /{id} [get]
// GetLink get original link
func (h *Handler) GetLink(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	q := chi.URLParam(req, "id")

	if len(q) != config.LENHASH {
		http.Error(res, errs.ErrCorrectURL.Error(), http.StatusBadRequest)
		return
	}

	h.l.Info("Get ID:", zap.String("id", q))
	foundedURL, err := h.s.GetLinkDB(ctx, istorage.ShortURL(q))

	if err != nil {
		if errors.Is(err, errs.ErrURLIsGone) {
			h.l.Info(errs.ErrURLIsGone.Error(), zap.Error(err))
			http.Error(res, errs.ErrURLIsGone.Error(), http.StatusGone)
			return
		}

		h.l.Info(errs.ErrBadRequest.Error(), zap.Error(err))
		http.Error(res, errs.ErrBadRequest.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(res, req, string(foundedURL), http.StatusTemporaryRedirect)
}

// SaveBatch godoc
// @Tags SaveBatch
// @Summary Request to save data and return multiply
// @Failure 400 {string} string "bad request"
// @Success 200 {object} object
// @Router /api/shorten/batch [post]
// SaveBatch save data and return multiply
func (h *Handler) SaveBatch(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrReadAll, err).Error(), http.StatusBadRequest)
		return
	}

	// get url from json data
	var urls []istorage.BatchReqURL
	err = json.Unmarshal(body, &urls)
	if err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrJSONUnMarshall, err).Error(), http.StatusBadRequest)
		return
	}

	shorts, err := h.s.SaveBatch(ctx, middlewares.GetContextUserID(req), urls)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
		return
	}

	// config instance
	cfg, err := config.Instance()
	if errors.Is(err, errs.ErrENVLoading) {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	// config value
	baseURL, err := cfg.GetCfgValue(config.BaseURL)
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	// prepare results
	for k := range shorts {
		shorts[k].Short = fmt.Sprintf("%s/%s", baseURL, shorts[k].Short)
	}

	body, err = json.Marshal(shorts)
	if err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrJSONMarshall, err).Error(), http.StatusBadRequest)
		return
	}

	res.Header().Add("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(body)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
	}

}

// SaveTXT godoc
// @Tags SaveTXT
// @Summary Convert link to shorting and store in database
// @Failure 400 {string} string "bad request"
// @Success 200 {string} string
// @Router / [post]
// SaveTXT convert link to shorting and store in database
func (h *Handler) SaveTXT(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	// Config instance
	cfg, err := config.Instance()
	if errors.Is(err, errs.ErrENVLoading) {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	// Config value
	baseURL, err := cfg.GetCfgValue(config.BaseURL)
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrReadAll, err).Error(), http.StatusInternalServerError)
		return
	}

	body := string(b)

	if body == "" {
		http.Error(res, errs.ErrEmptyBody.Error(), http.StatusBadRequest)
		return
	}

	userID := middlewares.GetContextUserID(req)

	origin, err := h.s.SaveLinkDB(ctx, istorage.UniqUser(userID), istorage.Origin(body))

	status := http.StatusCreated
	if errors.Is(err, errs.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}

	response := fmt.Sprintf("%s/%s", baseURL, origin)
	res.Header().Set("content-type", "text/plain; charset=utf-8")
	res.WriteHeader(status)
	res.Write([]byte(response))
}

// SaveJSON godoc
// @Tags SaveJSON
// @Summary Convert link to shorting and store in database
// @Failure 400 {string} string "bad request"
// @Success 200 {object} object
// @Router /api/shorten [post]
// SaveJSON convert link to shorting and store in database
func (h *Handler) SaveJSON(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	// config instance
	cfg, err := config.Instance()
	if errors.Is(err, errs.ErrENVLoading) {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	// config value
	baseURL, err := cfg.GetCfgValue(config.BaseURL)
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrReadAll, err).Error(), http.StatusInternalServerError)
		return
	}

	reqBody := struct {
		URL string `json:"url"`
	}{}

	decJSON := json.NewDecoder(strings.NewReader(string(body)))
	decJSON.DisallowUnknownFields()

	if err = decJSON.Decode(&reqBody); err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrFieldsJSON, err).Error(), http.StatusBadRequest)
		return
	}

	userID := middlewares.GetContextUserID(req)

	dbURL, err := h.s.SaveLinkDB(ctx, istorage.UniqUser(userID), istorage.Origin(reqBody.URL))
	status := http.StatusCreated
	if errors.Is(err, errs.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}

	resBody := struct {
		Result string `json:"result"`
	}{
		Result: fmt.Sprintf("%s/%s", baseURL, dbURL),
	}

	js, err := json.Marshal(resBody)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("content-type", "application/json")
	res.WriteHeader(status)
	res.Write(js)
}

// GetPing godoc
// @Tags GetPing
// @Summary Implement ping connection for sql database storage
// @Failure 500 {string} string "internal error"
// @Success 200 {string} string
// @Router /ping [get]
// GetPing implement ping connection for sql database storage
func (h *Handler) GetPing(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	conn, err := storage.SQLDBConnection(h.l)
	if err == nil {
		err = conn.Ping(ctx)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	} else {
		h.l.Info("not connect to db", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
	}
}

// GetLinks godoc
// @Tags GetLinks
// @Summary Get all urls by user
// @Failure 500 {string} string "internal error"
// @Success 200 {object} object
// @Router /api/user/urls [get]
// GetLinks get all urls by user
func (h *Handler) GetLinks(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	userID := middlewares.GetContextUserID(req)

	links, err := h.s.LinksByUser(ctx, istorage.UniqUser(userID))
	if err != nil {
		http.Error(res, errs.ErrNoContent.Error(), http.StatusNoContent)
		return
	}

	type coupleLinks struct {
		Short  string `json:"short_url"`
		Origin string `json:"original_url"`
	}

	lks := make([]coupleLinks, 0, len(links))

	// config instance
	cfg, err := config.Instance()
	if errors.Is(err, errs.ErrENVLoading) {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	// config value
	baseURL, err := cfg.GetCfgValue(config.BaseURL)
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	// get all links
	for k, v := range links {
		lks = append(lks, coupleLinks{
			Short:  fmt.Sprintf("%s/%s", baseURL, string(k)),
			Origin: string(v),
		})
	}

	body, err := json.Marshal(lks)
	if err != nil {
		http.Error(res, errs.ErrJSONMarshall.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Add("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write(body)
}

// GetStats godoc
// @Tags GetStats
// @Summary Request to get statistics quantity urls and users
// @Failure 400 {string} string "bad request"
// @Success 200 {string} string
// @Router /api/internal/stats [get]
// GetStats get statistics quantity urls and users
func (h *Handler) GetStats(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	userID := middlewares.GetContextUserID(req)

	foundedStat, err := h.s.GetStats(ctx, istorage.UniqUser(userID))
	if errors.Is(err, errs.ErrInternalSrv) {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(foundedStat)
	if err != nil {
		http.Error(res, errs.ErrJSONMarshall.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Add("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write(body)
}
