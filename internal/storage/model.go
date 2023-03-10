package storage

import (
	"bufio"
	"encoding/json"
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

type Event struct {
	Key string `json:"key"`
	Url string `json:"url"`
}

type Producer struct {
	file    *os.File
	writer  *bufio.Writer
	scanner *bufio.Scanner
}

func (f *Producer) WriteEvent(event *Event) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := f.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := f.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return f.writer.Flush()

	// return f.encoder.Encode(&event)
}

func (c *Producer) ReadEvent() (*Event, error) {
	// одиночное сканирование до следующей строки
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()

	event := Event{}
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (p *Producer) Close() error {
	return p.file.Close()
}
