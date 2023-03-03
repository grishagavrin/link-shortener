package storage

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/grishagavrin/link-shortener/internal/config"
)

const hashSymbols = "1234567890qwertyuiopasdfghjklzxcvbnm"

func AddLinkInDB(inputURL string) string {
	genKey := randStringBytes(config.LENHASH)
	urlString := RepositoryAddLik(inputURL, genKey)
	return fmt.Sprintf("http://localhost:8080/%s", urlString)
}

func GetLink(id string) (string, error) {
	url := RepositoryGetLink(id)
	if url == "" {
		return url, errors.New("DB doesn`t have value")
	}

	return url, nil
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = hashSymbols[rand.Intn(len(hashSymbols))]
	}
	return string(b)
}
