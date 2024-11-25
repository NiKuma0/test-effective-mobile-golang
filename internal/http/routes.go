package http

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) Routes(group *gin.RouterGroup) {
	songs := group.Group("/songs")
	songs.Use(h.TransactionMiddleware)
	{
		songs.GET("", h.ListAllSongs)
		songs.POST("", h.CreateSong)
		songs.GET("/:id/text", h.GetSongText)
		songs.PATCH("/:id", h.UpdateSong)
		songs.GET("/info", h.GetSongDetail)
	}
}
