package storage

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/grishagavrin/link-shortener/internal/config"
)

func AddLinkInDB(inputURL string) (string, error) {
	baseURL, err := config.Instance().GetCfgValue(config.BaseURL)
	if err != nil {
		return "", fmt.Errorf("env tag is not created, %s", config.BaseURL)
	}

	genKey := randStringBytes(config.LENHASH)

	filePath, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil {
		return "", fmt.Errorf("env tag is not created, %s", config.FileStoragePath)
	}

	if filePath == "" {
		urlString := RepositoryAddLink(inputURL, genKey)
		return fmt.Sprintf("%s/%s", baseURL, urlString), nil
	}

	var urlRec = &URLRecordInFile{
		Key: genKey,
		URL: inputURL,
	}

	saved := RepositoryWriteFileDB(filePath, urlRec)
	if !saved {
		return "", fmt.Errorf("something went wrong with write file")
	}

	return fmt.Sprintf("%s/%s", baseURL, genKey), nil
}

func GetLink(id string) (string, error) {

	filePath, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil {
		return "", fmt.Errorf("env tag is not created, %s", config.FileStoragePath)
	}

	if filePath == "" {
		url := RepositoryGetLink(id)
		if url == "" {
			return url, errors.New("DB doesn`t have value")
		}
		return url, nil
	}

	foundedURL, err := RepositoryReadFileDB(filePath, id)
	if err != nil {
		return "", errors.New(err.Error())
	}

	return foundedURL, nil
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = config.HashSymbols[rand.Intn(len(config.HashSymbols))]
	}
	return string(b)
}
