package http

import "github.com/nikuma0/test-effective-mobile-golang/internal/repository/postgresql"

type Handler struct {
	songsRepo postgresql.SongsRepositoryI
}

func New(songsRepo postgresql.SongsRepositoryI) Handler {
	return Handler{songsRepo: songsRepo}
}
