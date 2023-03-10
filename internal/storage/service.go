package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/grishagavrin/link-shortener/internal/config"
)

func AddLinkInDB(inputURL string) string {
	cfg := config.ConfigENV{}
	baseURL, exists := cfg.GetEnvValue(config.BaseURL)
	if !exists {
		fmt.Printf("env tag is not created, %s", config.BaseURL)
	}

	genKey := randStringBytes(config.LENHASH)
	// При отсутствии переменной окружения или при её пустом значении вернитесь к хранению сокращённых URL в памяти.

	filePath, exists := cfg.GetEnvValue(config.FileStoragePath)

	if !exists {
		//Если env не содержить путь до файла
		//То пишем в переменную
		urlString := RepositoryAddLik(inputURL, genKey)
		return fmt.Sprintf("%s/%s", baseURL, urlString)
	}

	AddLinkInFile(filePath, inputURL, genKey)
	return fmt.Sprintf("%s/%s", baseURL, genKey)
}

func AddLinkInFile(filePath, url, key string) {
	// producer, err := NewProducer(filePath)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		// return nil, err
	}

	writeFile := &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}

	defer writeFile.Close()

	var event = &Event{
		Key: key,
		Url: url,
	}

	if err := writeFile.WriteEvent(event); err != nil {
		log.Fatal(err)
	}

}

func NewProducer(fileName string) (*Producer, error) {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func GetLink(id string) (string, error) {
	cfg := config.ConfigENV{}
	filePath, exists := cfg.GetEnvValue(config.FileStoragePath)

	if !exists {
		//Если env не содержить путь до файла
		//То читаем из переменной
		url := RepositoryGetLink(id)
		if url == "" {
			return url, errors.New("DB doesn`t have value")
		}
		return url, nil
	}

	linkFile := GetLinkFromFile(filePath, id)

	return linkFile, nil

}

func GetLinkFromFile(filePath, key string) string {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		// return nil, err
	}

	readFile := &Producer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}
	defer readFile.Close()

	for readFile.scanner.Scan() {
		// fmt.Println(readFile.scanner.Text())
		// читаем данные из scanner
		data := readFile.scanner.Bytes()
		event := Event{}
		err := json.Unmarshal(data, &event)
		if err != nil {
			// return nil, err
		}
		if event.Key == key {
			return event.Url
		}
	}

	return ""

}
func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = config.HashSymbols[rand.Intn(len(config.HashSymbols))]
	}
	return string(b)
}
