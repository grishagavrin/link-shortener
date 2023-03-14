package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type LinkDB struct {
	MU       sync.Mutex
	LinkList map[string]string
}

func (db *LinkDB) AddLinkDB(url string, key string) string {
	db.MU.Lock()
	defer db.MU.Unlock()

	if _, ok := db.LinkList[key]; !ok {
		db.LinkList[key] = url
	}
	return key
}

func (db *LinkDB) GetLinkDB(key string) string {
	db.MU.Lock()
	defer db.MU.Unlock()
	return db.LinkList[key]
}

type UrlRecordInFile struct {
	Key string `json:"key"`
	Url string `json:"url"`
}

type FileDB struct {
	MU      sync.Mutex
	file    *os.File
	writer  *bufio.Writer
	scanner *bufio.Scanner
}

func (f *FileDB) WriteEvent(urlRec *UrlRecordInFile) error {
	data, err := json.Marshal(&urlRec)
	if err != nil {
		return err
	}

	f.MU.Lock()
	defer f.MU.Unlock()
	if _, err := f.writer.Write(data); err != nil {
		return err
	}

	if err := f.writer.WriteByte('\n'); err != nil {
		return err
	}

	return f.writer.Flush()
}

func (f *FileDB) ReadEvent(key string) (string, error) {
	f.MU.Lock()
	defer f.MU.Unlock()

	urlRec := UrlRecordInFile{}

	for {
		if !f.scanner.Scan() {
			return "", errors.New("url not found")
		}
		data := f.scanner.Bytes()

		err := json.Unmarshal(data, &urlRec)
		if err != nil {
			return "", err
		}

		if urlRec.Key == key {
			return urlRec.Url, nil
		}
	}
}

func (p *FileDB) Close() error {
	return p.file.Close()
}
