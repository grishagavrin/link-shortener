package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/storage"
)

func AddLink(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	body := string(b)

	if body == "" {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	dbURL := storage.AddLinkInDB(body)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(dbURL))
}

func GetLink(w http.ResponseWriter, r *http.Request) {
	q := chi.URLParam(r, "id")
	if len(q) != config.LEN_HASH {
		http.Error(w, "Enter a number type parameter", http.StatusBadRequest)
		return
	}

	foundedUrl, err := storage.GetLink(q)
	if err != nil {
		http.Error(w, "The id parametr not found in DB", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, foundedUrl, http.StatusTemporaryRedirect)
}

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	reqBody := struct {
		URL string `json:"url"`
	}{}

	decodeJSON := json.NewDecoder(strings.NewReader(string(body)))
	decodeJSON.DisallowUnknownFields()

	if err := decodeJSON.Decode(&reqBody); err != nil {
		http.Error(w, "Invalid fields in JSON", http.StatusBadRequest)
		return
	}

	dbURL := storage.AddLinkInDB(reqBody.URL)
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
