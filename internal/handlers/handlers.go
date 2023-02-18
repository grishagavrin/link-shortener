package handlers

import (
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/go-chi/chi"
	"github.com/grishagavrin/link-shortener/internal/storage"
)

func GetURL(w http.ResponseWriter, r *http.Request) {
	if !regexp.MustCompile(`^/[0-9]+$`).MatchString(r.URL.Path) {
		http.Error(w, "Enter a number type parameter", http.StatusBadRequest)
		return
	}

	q := chi.URLParam(r, "id")

	founded, err := storage.GetURLById(q)

	if err != nil {
		http.Error(w, "The id parametr not found in DB", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, founded.Address, http.StatusTemporaryRedirect)
}

func WriteURL(w http.ResponseWriter, r *http.Request) {

	b, _ := io.ReadAll(r.Body)
	body := string(b)

	if body == "" {
		http.Error(w, "Body is empty", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	dbURL := storage.AddURL(body)
	myString := fmt.Sprintf("http://localhost:8080/%s", dbURL.ID)

	w.Write([]byte(myString))
}
