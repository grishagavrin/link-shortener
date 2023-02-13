package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/grishagavrin/link-shortener/internal/storage"
)

var DB = map[int]string{}

func CommonHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Uncorrected route", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		q := r.URL.Query().Get("id")
		if q == "" {
			http.Error(w, "The id parameter is missing", http.StatusBadRequest)
			return
		}

		intID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		founded, err := storage.GetURLById(intID)
		if err != nil {
			http.Error(w, "id parametr not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, founded.Address, http.StatusTemporaryRedirect)
		return

	case "POST":
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		stringURL := r.FormValue("url")

		if stringURL == "" {
			http.Error(w, "The url parameter is missing", http.StatusBadRequest)
			return
		}

		dbURL := storage.AddURL(stringURL)

		resp, err := json.Marshal(dbURL)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write(resp)
	}
}
