package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nikuma0/test-effective-mobile-golang/internal/repository/postgresql"
)

type Handler struct {
	songsRepo       postgresql.SongsRepositoryI
	songsRepoGetter func() postgresql.SongsRepositoryI
}

func New(songsRepoGetter func() postgresql.SongsRepositoryI) Handler {
	return Handler{songsRepoGetter: songsRepoGetter}
}

func NewTest(songsRepo postgresql.SongsRepositoryI) Handler {
	return Handler{songsRepo: songsRepo, songsRepoGetter: func() postgresql.SongsRepositoryI { return songsRepo }}
}

func (h *Handler) TransactionMiddleware(c *gin.Context) {
	h.songsRepo = h.songsRepoGetter()
	tr, _ := h.songsRepo.Begin()
	c.Next()
	statusCode := c.Writer.Status()
	if statusCode >= http.StatusOK && statusCode < http.StatusBadRequest {
		tr.Commit()
	}
	if statusCode >= http.StatusBadRequest {
		tr.Rollback()
	}
}
