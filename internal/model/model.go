package model

type Links struct {
	Url string
	Id  string
}

func NewLinks(url string, id string) *Links {
	return &Links{
		Url: url,
		Id:  id,
	}
}
