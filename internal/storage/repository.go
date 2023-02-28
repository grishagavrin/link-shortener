package storage

import "errors"

var localDB = DB{Links: make([]RedirectURL, 0)}

func RepositoryAddURL(url string) RedirectURL {
	return localDB.AddLink(url)
}

func RepositoryGetURLByID(id string) (RedirectURL, error) {
	newURL := localDB.GetLink(id)
	if newURL == (RedirectURL{}) {
		return newURL, errors.New("DB doesn`t have value")
	}

	return newURL, nil

}
