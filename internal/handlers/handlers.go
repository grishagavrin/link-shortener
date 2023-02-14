package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/grishagavrin/link-shortener/internal/storage"
)

var DB = map[int]string{}

func CommonHandler(w http.ResponseWriter, r *http.Request) {
	// if r.URL.Path != "/" {
	// 	http.Error(w, "Uncorrected route", http.StatusBadRequest)
	// 	return
	// }

	switch r.Method {
	case "GET":
		q := r.URL.Path[1:]

		if q == "" {
			http.Error(w, "The id parameter is missing", http.StatusBadRequest)
			return
		}

		founded, err := storage.GetURLById(q)

		if err != nil {
			http.Error(w, "id parametr not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, founded.Address, http.StatusTemporaryRedirect)
		return

	case "POST":
		w.WriteHeader(http.StatusCreated)
		b, _ := io.ReadAll(r.Body)
		stringURL := string(b)

		if stringURL == "" {
			http.Error(w, fmt.Sprintf("The url parameter is missing, url is %s", string(b)), http.StatusBadRequest)
			return
		}

		dbURL := storage.AddURL(stringURL)

		myString := fmt.Sprintf("http://localhost:8080/%s", dbURL.Id)

		w.Write([]byte(myString))
	}
}
