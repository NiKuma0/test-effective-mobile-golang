package models

type Data[D any] struct {
	Ok   bool `json:"ok"`
	Data D    `json:"data"`
}

type Paginator[D any] struct {
	Data   D    `json:"data"`
	Page   int  `json:"page"`
	Amount int  `json:"amount"`
	Next   bool `json:"next"`
	Ok     bool `json:"ok"`
}

type Message struct {
	Ok  bool   `json:"ok"`
	Msg string `json:"msg"`
}

type ListAllSongs = Paginator[[]Song]
type SongsText = Paginator[[]string]
