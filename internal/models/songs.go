package models

import "time"

type PageMaxQuery struct {
	Page int `form:"page" validate:"gte=0"`
	Max  int `form:"max" validate:"gte=1"`
}

func NewPageMaxQuery() PageMaxQuery {
	return PageMaxQuery{
		Page: 0,
		Max:  10,
	}
}

type SongsQuery struct {
	Page        int        `form:"page" validate:"gte=0"`
	Max         int        `form:"max" validate:"gte=1"`
	Group       *string    `form:"group"`
	Song        *string    `form:"song"`
	ReleaseDate *time.Time `form:"releaseDate" validate:"datetime"`
}

type SongDetailQuery struct {
	Group string `form:"group" validate:"required"`
	Song  string `form:"song" validate:"required"`
}

func NewSongsQuery() SongsQuery {
	return SongsQuery{
		Page: 0,
		Max:  10,
	}
}

type SongCreateQuery struct {
	Group       string     `json:"group" validate:"required"`
	Song        string     `json:"song" validate:"required"`
	Text        string     `json:"text" validate:"required"`
	ReleaseDate *time.Time `json:"releaseDate" validate:"required,datetime"`
}

type Song struct {
	Id          int       `json:"id"`
	GroupName   string    `json:"group"`
	Name        string    `json:"song"`
	ReleaseDate time.Time `json:"releaseDate"`
}

type SongDetail struct {
	Id          int       `json:"id"`
	GroupName   string    `json:"group"`
	Name        string    `json:"song"`
	Text        string    `json:"text"`
	ReleaseDate time.Time `json:"releaseDate"`
}

type SongUpdate struct {
	GroupName   *string    `json:"group"`
	Name        *string    `json:"song"`
	Text        *string    `json:"text"`
	ReleaseDate *time.Time `json:"releaseDate"`
}
