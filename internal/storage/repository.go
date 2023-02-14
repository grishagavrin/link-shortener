package storage

import (
	"errors"
	"strconv"
)

var databaseURL []RedirectURL = []RedirectURL{}

func RepositoryAddURL(inputURL string) RedirectURL {

	id := strconv.Itoa(len(databaseURL))

	var newURL RedirectURL = RedirectURL{
		Id:      id,
		Address: inputURL,
	}

	databaseURL = append(databaseURL, newURL)
	return newURL
}

func RepositoryGetURLById(id string) (RedirectURL, error) {
	var newURL RedirectURL

	for _, v := range databaseURL {
		if v.Id == id {
			newURL = v
		}
	}

	if newURL.Address == "" {
		return newURL, errors.New("DB doesn`t have value")
	}

	return newURL, nil
}
