package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type DateFormat time.Time

func (df *DateFormat) UnmarshalJSON(data []byte) error {
	dateStr := string(data)
	dateStr = dateStr[1 : len(dateStr)-1]

	layout := "2006.01.02"

	t, err := time.Parse(layout, dateStr)
	if err != nil {
		return err
	}

	*df = DateFormat(t)
	return nil
}

func (cd DateFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(cd).Format("2006.01.02"))
}

func (cd *DateFormat) Scan(value interface{}) error {
	if value == nil {
		*cd = DateFormat(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*cd = DateFormat(v)
		return nil
	case string:
		parsedTime, err := time.Parse("2006-01-02", v)
		if err != nil {
			return err
		}
		*cd = DateFormat(parsedTime)
		return nil
	default:
		return errors.New("unsupported scan type")
	}
}

func (cd DateFormat) Value() (driver.Value, error) {
	if time.Time(cd).IsZero() {
		return nil, nil
	}
	return time.Time(cd), nil
}

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
	Page        int         `form:"page" validate:"gte=0"`
	Max         int         `form:"max" validate:"gte=1"`
	Group       *string     `form:"group"`
	Song        *string     `form:"song"`
	Link        *string     `form:"link"`
	ReleaseDate *DateFormat `form:"releaseDate" validate:"datetime"`
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
	Group       string      `json:"group" validate:"required"`
	Song        string      `json:"song" validate:"required"`
	Text        string      `json:"text" validate:"required"`
	Link        string      `json:"link" validate:"required"`
	ReleaseDate *DateFormat `json:"releaseDate" validate:"required,datetime"`
}

type Song struct {
	Id          int        `json:"id"`
	GroupName   string     `json:"group"`
	Name        string     `json:"song"`
	ReleaseDate DateFormat `json:"releaseDate"`
	Link        string     `json:"link"`
}

type SongDetail struct {
	Id          int        `json:"id"`
	GroupName   string     `json:"group"`
	Name        string     `json:"song"`
	Text        string     `json:"text"`
	ReleaseDate DateFormat `json:"releaseDate"`
	Link        string     `json:"link"`
}

type SongUpdate struct {
	GroupName   *string     `json:"group"`
	Name        *string     `json:"song"`
	Text        *string     `json:"text"`
	ReleaseDate *DateFormat `json:"releaseDate"`
	Link        *string     `json:"link"`
}
