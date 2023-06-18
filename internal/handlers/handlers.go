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
	"github.com/grishagavrin/link-shortener/internal/storage/dbstorage"
	"github.com/grishagavrin/link-shortener/internal/storage/ramstorage"
	"github.com/grishagavrin/link-shortener/internal/user"
	"github.com/grishagavrin/link-shortener/internal/utils/db"
	"go.uber.org/zap"
)

type Handler struct {
	s storage.Repository
	l *zap.Logger
}

func New(l *zap.Logger) (*Handler, error) {
	_, err := db.Instance(l)
	if err == nil {
		l.Info("Set DB handler")
		storage, err := dbstorage.New(l)
		if err != nil {
			return nil, err
		}

		return &Handler{s: storage, l: l}, nil
	} else {
		storage, err := ramstorage.New(l)
		if err != nil {
			return nil, err
		}

		l.Info("Set RAM handler")
		return &Handler{s: storage, l: l}, nil
	}
}

func (h *Handler) GetLink(res http.ResponseWriter, req *http.Request) {
	q := chi.URLParam(req, "id")
	if len(q) != config.LENHASH {
		http.Error(res, errs.ErrCorrectURL.Error(), http.StatusBadRequest)
		return
	}

	h.l.Info("Get ID:", zap.String("id", q))
	foundedURL, err := h.s.GetLinkDB(storage.Origin(q))

	if err != nil {
		if errors.Is(err, errs.ErrURLIsGone) {
			h.l.Info("Get error is gone", zap.Error(err))
			http.Error(res, errs.ErrURLIsGone.Error(), http.StatusGone)
			return
		}

		h.l.Info("Get error is bad request", zap.Error(err))
		http.Error(res, errs.ErrBadRequest.Error(), http.StatusBadRequest)
		return
	}

	h.l.Info("redirect")
	http.Redirect(res, req, string(foundedURL), http.StatusTemporaryRedirect)
}

func (h *Handler) SaveBatch(res http.ResponseWriter, req *http.Request) {
	h.l.Info("BunchSaveJSON run")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
		return
	}
	// Get url from json data
	var urls []storage.BatchURL
	err = json.Unmarshal(body, &urls)
	if err != nil {
		http.Error(res, errs.ErrCorrectURL.Error(), http.StatusBadRequest)
		return
	}

	userID := middlewares.GetContextUserID(req)

	shorts, err := h.s.SaveBatch(userID, urls)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
		return
	}
	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
		return
	}

	// Prepare results
	for k := range shorts {
		shorts[k].Short = fmt.Sprintf("%s/%s", baseURL, shorts[k].Short)
	}

	body, err = json.Marshal(shorts)
	if err == nil {
		// Prepare response
		res.Header().Add("Content-Type", "application/json; charset=utf-8")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write(body)
		if err == nil {
			return
		}
	}

	http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
}

func (h *Handler) SaveTXT(res http.ResponseWriter, req *http.Request) {
	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	b, _ := io.ReadAll(req.Body)
	body := string(b)

	if body == "" {
		http.Error(res, errs.ErrEmptyBody.Error(), http.StatusBadRequest)
		return
	}

	userID := middlewares.GetContextUserID(req)

	urlKey, err := h.s.SaveLinkDB(user.UniqUser(userID), storage.ShortURL(body))
	status := http.StatusCreated

	if errors.Is(err, errs.ErrAlreadyHasShort) {
		status = http.StatusConflict
	}

	response := fmt.Sprintf("%s/%s", baseURL, urlKey)
	res.WriteHeader(status)
	res.Write([]byte(response))
}

func (h *Handler) SaveJSON(res http.ResponseWriter, req *http.Request) {
	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	body, _ := io.ReadAll(req.Body)
	reqBody := struct {
		URL string `json:"url"`
	}{}

	decJSON := json.NewDecoder(strings.NewReader(string(body)))
	decJSON.DisallowUnknownFields()

	if err := decJSON.Decode(&reqBody); err != nil {
		http.Error(res, errs.ErrFieldsJSON.Error(), http.StatusBadRequest)
		return
	}

	userID := middlewares.GetContextUserID(req)

	dbURL, err := h.s.SaveLinkDB(user.UniqUser(userID), storage.ShortURL(reqBody.URL))
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
	conn, err := db.Instance(h.l)
	if err == nil {
		if err := conn.Ping(req.Context()); err == nil {
			res.WriteHeader(http.StatusOK)
		}
	} else {
		h.l.Info("not connect to db", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) GetLinks(res http.ResponseWriter, req *http.Request) {
	userIDCtx := req.Context().Value(middlewares.UserIDCtxName)
	// Convert interface type to user.UniqUser
	userID := userIDCtx.(string)

	links, err := h.s.LinksByUser(user.UniqUser(userID))
	if err != nil {
		http.Error(res, errs.ErrNoContent.Error(), http.StatusNoContent)
		return
	}

	type coupleLinks struct {
		Short  string `json:"short_url"`
		Origin string `json:"original_url"`
	}
	var lks []coupleLinks
	baseURL, _ := config.Instance().GetCfgValue(config.BaseURL)

	// Get all links
	for k, v := range links {
		lks = append(lks, coupleLinks{
			Short:  fmt.Sprintf("%s/%s", baseURL, string(k)),
			Origin: string(v),
		})
	}

	body, err := json.Marshal(lks)
	if err == nil {
		// Prepare response
		res.Header().Add("Content-Type", "application/json; charset=utf-8")
		res.WriteHeader(http.StatusOK)
		_, err = res.Write(body)
		if err == nil {
			return
		}
	}
}

func (h *Handler) DeleteBatch(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, errs.ErrInternalSrv.Error(), http.StatusBadRequest)
		return
	}

	var correlationIDs []string
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

	userID := middlewares.GetContextUserID(req)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
