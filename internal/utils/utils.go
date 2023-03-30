package utils

import (
	"math/rand"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/storage"
)

func RandStringBytes(n int) storage.URLKey {
	b := make([]byte, n)
	for i := range b {
		b[i] = config.HashSymbols[rand.Intn(len(config.HashSymbols))]
	}
	return storage.URLKey(b)
}
