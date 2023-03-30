package filestorage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/utils"
)

var errNotExistEnv = errors.New("unknown env or flag param")
var errNotOpenFile = errors.New("file doesn`t open")
var errNotFoundURL = errors.New("url in file doesn`t found")

type URLRecordInFile struct {
	Key storage.URLKey   `json:"key"`
	URL storage.ShortURL `json:"url"`
}
type FileStorage struct {
	MU      sync.Mutex
	file    *os.File
	writer  *bufio.Writer
	scanner *bufio.Scanner
}

func New() *FileStorage {
	return &FileStorage{}

}

func (f *FileStorage) SaveLinkDB(url storage.ShortURL) (storage.URLKey, error) {
	genKey := utils.RandStringBytes(config.LENHASH)

	filePath, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil {
		return "", errNotExistEnv
	}

	urlRec := URLRecordInFile{
		Key: genKey,
		URL: url,
	}

	f.MU.Lock()
	defer f.MU.Unlock()
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return "", err
	}

	f.file = file
	f.writer = bufio.NewWriter(file)
	defer f.Close()

	data, err := json.Marshal(&urlRec)
	if err != nil {
		return "", err
	}
	_ = data

	if _, err := f.writer.Write(data); err != nil {
		return "", err
	}

	if err := f.writer.WriteByte('\n'); err != nil {
		return "", err
	}

	f.writer.Flush()

	return genKey, nil
}

func (f *FileStorage) GetLinkDB(key storage.URLKey) (storage.ShortURL, error) {

	filePath, err := config.Instance().GetCfgValue(config.FileStoragePath)
	if err != nil {
		return "", errNotExistEnv
	}

	f.MU.Lock()
	defer f.MU.Unlock()
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", errNotOpenFile
	}

	f.file = file
	f.scanner = bufio.NewScanner(file)
	defer f.Close()

	urlRec := URLRecordInFile{}

	for {
		if !f.scanner.Scan() {
			return "", errNotFoundURL
		}

		data := f.scanner.Bytes()
		err := json.Unmarshal(data, &urlRec)
		if err != nil {
			return "", err
		}

		if urlRec.Key == key {
			return urlRec.URL, nil
		}
	}
}

func (f *FileStorage) Close() error {
	return f.file.Close()
}
