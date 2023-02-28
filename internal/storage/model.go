package storage

import (
	"strconv"
	"sync"
)

type DB struct {
	MU    sync.Mutex
	Links []RedirectURL
}

type RedirectURL struct {
	ID      string
	Address string
}

func (db *DB) AddLink(url string) RedirectURL {
	db.MU.Lock()
	defer db.MU.Unlock()
	id := strconv.Itoa(len(db.Links))

	newURL := RedirectURL{
		ID:      id,
		Address: url,
	}

	db.Links = append(db.Links, newURL)
	return newURL
}

func (db *DB) GetLink(id string) RedirectURL {
	db.MU.Lock()
	defer db.MU.Unlock()
	var newURL RedirectURL

	for _, v := range db.Links {
		if v.ID == id {
			newURL = RedirectURL{
				v.ID,
				v.Address,
			}
			break
		}
	}

	return newURL
}
