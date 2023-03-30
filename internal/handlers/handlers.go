package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/storage/filestorage"
	"github.com/grishagavrin/link-shortener/internal/storage/ramstorage"
)

type Handler struct {
	s storage.Repository
}

var errEmptyBody = errors.New("body is empty")
var errFieldsJSON = errors.New("invalid fields in json")
var errInternalSrv = errors.New("internal error on server")
var errCorrectURL = fmt.Errorf("enter correct url parameter - length: %v", config.LENHASH)

func New() (h *Handler, err error) {
	// If set file storage path
	fs, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil || fs == "" {
		return &Handler{s: ramstorage.New()}, nil
	} else {
		return &Handler{s: filestorage.New()}, nil
	}
}

func (h *Handler) GetLink(w http.ResponseWriter, r *http.Request) {
	q := chi.URLParam(r, "id")
	if len(q) != config.LENHASH {
		http.Error(w, errCorrectURL.Error(), http.StatusBadRequest)
		return
	}

	foundedURL, err := h.s.GetLinkDB(storage.URLKey(q))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, string(foundedURL), http.StatusTemporaryRedirect)
}

func (h *Handler) SaveTXT(w http.ResponseWriter, r *http.Request) {
	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	b, _ := io.ReadAll(r.Body)
	body := string(b)

	if body == "" {
		http.Error(w, errEmptyBody.Error(), http.StatusBadRequest)
		return
	}

	urlKey, err := h.s.SaveLinkDB(storage.ShortURL(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := fmt.Sprintf("%s/%s", baseURL, urlKey)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(res))
}

func (h *Handler) SaveJSON(w http.ResponseWriter, r *http.Request) {
	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	body, _ := io.ReadAll(r.Body)
	reqBody := struct {
		URL string `json:"url"`
	}{}

	decJSON := json.NewDecoder(strings.NewReader(string(body)))
	decJSON.DisallowUnknownFields()

	if err := decJSON.Decode(&reqBody); err != nil {
		http.Error(w, errFieldsJSON.Error(), http.StatusBadRequest)
		return
	}

	dbURL, err := h.s.SaveLinkDB(storage.ShortURL(reqBody.URL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resBody := struct {
		Result string `json:"result"`
	}{
		Result: fmt.Sprintf("%s/%s", baseURL, dbURL),
	}

	res, err := json.Marshal(resBody)
	if err != nil {
		http.Error(w, errInternalSrv.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}
