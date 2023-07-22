package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/handlers/middlewares"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"go.uber.org/zap"
)

type Handler struct {
	s iStorage.Repository
	l *zap.Logger
}

func New(stor iStorage.Repository, l *zap.Logger) *Handler {
	return &Handler{s: stor, l: l}
}

func (h *Handler) GetLink(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	q := chi.URLParam(req, "id")
	if len(q) != config.LENHASH {
		http.Error(res, errs.ErrCorrectURL.Error(), http.StatusBadRequest)
		return
	}

	h.l.Info("Get ID:", zap.String("id", q))
	foundedURL, err := h.s.GetLinkDB(ctx, iStorage.ShortURL(q))

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

func (h *Handler) SaveBatch(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrReadAll, err).Error(), http.StatusBadRequest)
		return
	}

	// Get url from json data
	var urls []iStorage.BatchReqURL
	err = json.Unmarshal(body, &urls)
	if err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrJsonUnMarshall, err).Error(), http.StatusBadRequest)
		return
	}

	shorts, err := h.s.SaveBatch(ctx, middlewares.GetContextUserID(req), urls)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
		return
	}

	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare results
	for k := range shorts {
		shorts[k].Short = fmt.Sprintf("%s/%s", baseURL, shorts[k].Short)
	}

	body, err = json.Marshal(shorts)
	if err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrJsonMarshall, err).Error(), http.StatusBadRequest)
		return
	}

	res.Header().Add("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(body)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
	}

}

func (h *Handler) SaveTXT(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrUnknownEnvOrFlag, err).Error(), http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(req.Body)
	body := string(b)
	if err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrReadAll, err).Error(), http.StatusInternalServerError)
		return
	}

	if body == "" {
		http.Error(res, errs.ErrEmptyBody.Error(), http.StatusBadRequest)
		return
	}

	userID := middlewares.GetContextUserID(req)

	origin, err := h.s.SaveLinkDB(ctx, user.UniqUser(userID), iStorage.Origin(body))

	status := http.StatusCreated
	if errors.Is(err, errs.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}

	response := fmt.Sprintf("%s/%s", baseURL, origin)
	res.WriteHeader(status)
	res.Write([]byte(response))
}

func (h *Handler) SaveJSON(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrUnknownEnvOrFlag, err).Error(), http.StatusInternalServerError)
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

	if err := decJSON.Decode(&reqBody); err != nil {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrFieldsJSON, err).Error(), http.StatusBadRequest)
		return
	}

	userID := middlewares.GetContextUserID(req)

	dbURL, err := h.s.SaveLinkDB(ctx, user.UniqUser(userID), iStorage.Origin(reqBody.URL))
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

func (h *Handler) GetPing(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	conn, err := storage.SQLDBConnection(h.l)
	if err == nil {
		if err := conn.Ping(ctx); err == nil {
			res.WriteHeader(http.StatusOK)
		}
	} else {
		h.l.Info("not connect to db", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) GetLinks(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()

	userID := middlewares.GetContextUserID(req)

	links, err := h.s.LinksByUser(ctx, user.UniqUser(userID))
	if err != nil {
		http.Error(res, errs.ErrNoContent.Error(), http.StatusNoContent)
		return
	}

	type coupleLinks struct {
		Short  string `json:"short_url"`
		Origin string `json:"original_url"`
	}

	lks := make([]coupleLinks, 0, len(links))
	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		http.Error(res, fmt.Errorf("%w: %v", errs.ErrUnknownEnvOrFlag, err).Error(), http.StatusInternalServerError)
		return
	}

	// Get all links
	for k, v := range links {
		lks = append(lks, coupleLinks{
			Short:  fmt.Sprintf("%s/%s", baseURL, string(k)),
			Origin: string(v),
		})
	}

	body, err := json.Marshal(lks)
	if err == nil {
		res.Header().Add("Content-Type", "application/json; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		_, err = res.Write(body)
		if err == nil {
			return
		}
	}
}

func (h *Handler) DeleteBatch(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
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

	inputCh := make(chan string)
	go func() {
		for _, id := range correlationIDs {
			inputCh <- id
		}
		close(inputCh)
	}()

	out := fanIn(ctx, string(userID), inputCh)

	var idS []string
	for value := range out {
		idS = append(idS, value)
	}
	err = h.s.BunchUpdateAsDeleted(ctx, idS, string(userID))
	if err != nil {
		fmt.Println(err)
	}

	res.WriteHeader(http.StatusAccepted)
}

func fanIn(ctx context.Context, userID string, inputs ...<-chan string) <-chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	wg.Add(len(inputs))

	for _, in := range inputs {
		go func(ch <-chan string) {
			for {
				value, ok := <-ch

				if !ok {
					wg.Done()
					break
				}

				out <- value
			}
		}(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
