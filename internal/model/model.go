package model

type Links struct {
	URL string
	ID  string
}

func NewLinks(url string, id string) *Links {
	return &Links{
		URL: url,
		ID:  id,
	}
}
