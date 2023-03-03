package storage

import (
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
