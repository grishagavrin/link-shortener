package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/storage"
)

func SaveTXT(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	body := string(b)

	if body == "" {
		http.Error(w, "body is empty", http.StatusBadRequest)
		return
	}

	res, err := storage.AddLinkInDB(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(res))
}

func GetLink(w http.ResponseWriter, r *http.Request) {
	q := chi.URLParam(r, "id")
	if len(q) != config.LENHASH {
		http.Error(w, fmt.Sprintf("enter correct url parameter - length: %v", config.LENHASH), http.StatusBadRequest)
		return
	}

	foundedURL, err := storage.GetLink(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, foundedURL, http.StatusTemporaryRedirect)
}

func SaveJSON(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	reqBody := struct {
		URL string `json:"url"`
	}{}

	decJSON := json.NewDecoder(strings.NewReader(string(body)))
	decJSON.DisallowUnknownFields()

	if err := decJSON.Decode(&reqBody); err != nil {
		http.Error(w, "invalid fields in JSON", http.StatusBadRequest)
		return
	}

	dbURL, err := storage.AddLinkInDB(reqBody.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resBody := struct {
		ValueDB string `json:"result"`
	}{
		ValueDB: dbURL,
	}

	res, err := json.Marshal(resBody)
	if err != nil {
		http.Error(w, "Internal error on server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}
