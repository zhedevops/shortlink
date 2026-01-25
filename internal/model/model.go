package model

type Links struct {
	URL string
	ID  string
}

type Request struct {
	Url string `json:"url,required"`
}

type Response struct {
	Result string `json:"result"`
}

func NewLinks(url string, id string) *Links {
	return &Links{
		URL: url,
		ID:  id,
	}
}
