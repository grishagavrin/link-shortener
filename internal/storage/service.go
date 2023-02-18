package storage

func GetURLById(id string) (RedirectURL, error) {
	url, err := RepositoryGetURLByID(id)

	if err != nil {
		return url, err
	}

	return url, nil
}

func AddURL(inputURL string) RedirectURL {
	return RepositoryAddURL(inputURL)
}
