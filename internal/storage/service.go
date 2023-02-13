package storage

import "fmt"

func GetURLById(id int) (RedirectURL, error) {
	url, err := RepositoryGetURLById(id)

	if err != nil {
		fmt.Println(err)
		return url, err
	}

	return url, nil
}

func AddURL(inputURL string) RedirectURL {
	return RepositoryAddURL(inputURL)
}
