package storage

import (
	"errors"
	"fmt"
	"log"
	"math/rand"

	"github.com/grishagavrin/link-shortener/internal/config"
)

const hashSymbols = "1234567890qwertyuiopasdfghjklzxcvbnm"

func AddLinkInDB(inputURL string) string {
	cfg := config.ConfigENV{}
	baseUrl, exists := cfg.GetEnvValue(config.BaseURL)
	if !exists {
		log.Fatalf("env tag is not created, %s", config.BaseURL)
	}

	genKey := randStringBytes(config.LENHASH)
	urlString := RepositoryAddLik(inputURL, genKey)
	return fmt.Sprintf("%s/%s", baseUrl, urlString)
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
