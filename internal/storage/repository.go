package storage

var MyDB = LinkDB{LinkList: make(map[string]string)}

func RepositoryAddLik(url string, key string) string {
	return MyDB.AddLinkDB(url, key)
}

func RepositoryGetLink(id string) string {
	return MyDB.GetLinkDB(id)
}
