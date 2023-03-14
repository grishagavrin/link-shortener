package storage

import (
	"bufio"
	"errors"
	"os"
)

var MyDB = LinkDB{LinkList: make(map[string]string)}

func RepositoryReadFileDB(filePath, key string) (string, error) {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", errors.New("file doesn`t read")
	}

	readFile := &FileDB{
		file:    file,
		scanner: bufio.NewScanner(file),
	}
	defer readFile.Close()

	str, err := readFile.ReadEvent(key)
	if err != nil {
		return "", errors.New(err.Error())
	}

	return str, nil
}

func RepositoryWriteFileDB(filePath string, urlRec *UrlRecordInFile) bool {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return false
	}

	wrFile := &FileDB{
		file:   file,
		writer: bufio.NewWriter(file),
	}
	defer wrFile.Close()

	if err := wrFile.WriteEvent(urlRec); err != nil {
		return false
	}
	return true
}

func RepositoryAddLink(url string, key string) string {
	return MyDB.AddLinkDB(url, key)
}

func RepositoryGetLink(id string) string {
	return MyDB.GetLinkDB(id)
}
