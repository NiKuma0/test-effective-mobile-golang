package http

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/nikuma0/test-effective-mobile-golang/internal/models"
)

// ListAllSongs godoc
//
//	@Summary		Show all songs
//	@Description	Paginate all songs filtered by song name or/and group name.
//	@Tags			Songs
//	@Produce		json
//	@Param			page	query		int					false	"Page (starts with 0)"
//	@Param			max		query		int					false	"Maximum elements (default 10)"
//	@Param			group	query		string				false	"Group name"
//	@Param			song	query		string				false	"Song name"
//	@Success		200		{object}	models.ListAllSongs	"List of songs with pagination details"
//	@Failure		400		{object}	models.Message		"Bad request, invalid parameters"
//	@Failure		404		{object}	models.Message		"Not found, no songs match the criteria or page is empty"
//	@Failure		500		{object}	models.Message		"Internal server error"
//	@Router			/songs [get]
func (h *Handler) ListAllSongs(c *gin.Context) {
	sq := models.NewSongsQuery()
	c.Bind(&sq)
	songs, amount, err := h.songsRepo.GetSongs(c.Request.Context(), &sq)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Message{
			Ok:  false,
			Msg: "something went wrong",
		})
		log.Panic(err.Error())
		return
	}

	c.JSON(http.StatusOK, models.ListAllSongs{
		Ok:     true,
		Data:   songs,
		Page:   sq.Page,
		Next:   sq.Max*(sq.Page+1) < amount,
		Amount: amount,
	})
}

// CreateSong godoc
//
//	@Summary		Create a new song
//	@Description	Creates a new song in the database with the provided details.
//	@Tags			Songs
//	@Accept			json
//	@Produce		json
//	@Param			body	body		models.SongCreate	true	"Song details"
//	@Success		200		{object}	models.Message		"OK response with success message"
//	@Failure		400		{object}	models.Message		"Bad request, invalid data"
//	@Failure		500		{object}	models.Message		"Internal server error"
//	@Router			/songs [post]
func (h *Handler) CreateSong(c *gin.Context) {
	var scq models.SongCreateQuery
	if err := c.ShouldBind(&scq); err != nil {
		c.JSON(http.StatusBadRequest, models.Message{
			Ok:  false,
			Msg: err.Error(),
		})
		return
	}
	if err := h.songsRepo.CreateSong(c.Request.Context(), &scq); err != nil {
		c.JSON(http.StatusBadGateway, models.Message{
			Ok:  false,
			Msg: "something went wrong",
		})
		log.Panic(err.Error())
		return
	}
	c.JSON(http.StatusCreated, models.Message{
		Ok:  true,
		Msg: "created",
	})
}

// GetSongDetail godoc
//
//	@Summary		Get details of a specific song
//	@Description	Retrieve detailed information about a song based on the provided query parameters.
//	@Tags			Songs
//	@Produce		json
//	@Param			group	query		string				true	"Group name"
//	@Param			song	query		string				true	"Song name"
//	@Success		200		{object}	models.SongDetail	"Song details"
//	@Failure		404		{object}	models.Message		"Song not found"
//	@Failure		500		{object}	models.Message		"Internal server error"
//	@Router			/songs/info [get]
func (h *Handler) GetSongDetail(c *gin.Context) {
	var sdq models.SongDetailQuery
	c.Bind(&sdq)
	sd, err := h.songsRepo.GetSong(c.Request.Context(), &sdq)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Message{Ok: false, Msg: "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Message{Ok: false, Msg: "something went wrong"})
		log.Panic(err.Error())
		return
	}
	c.JSON(http.StatusOK, sd)
}

// UpdateSong godoc
//
//	@Summary		Update a song
//	@Description	Update one or more fields of a specific song by its ID.
//	@Tags			Songs
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Song ID"
//	@Param			body	body		models.SongUpdate	true	"Fields to update"
//	@Success		200		{object}	models.Message		"Song successfully updated"
//	@Failure		400		{object}	models.Message		"Invalid song ID"
//	@Failure		404		{object}	models.Message		"Song not found"
//	@Failure		502		{object}	models.Message		"Internal server error"
//	@Router			/songs/{id} [patch]
func (h *Handler) UpdateSong(c *gin.Context) {
	var su models.SongUpdate

	songId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Message{Ok: false, Msg: err.Error()})
		return
	}
	c.Bind(&su)

	exists, err := h.songsRepo.CheckIfExists(c.Request.Context(), songId)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Message{Ok: false, Msg: "something went wrong"})
		log.Panic(err.Error())
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, models.Message{Ok: false, Msg: "not found"})
		return
	}

	err = h.songsRepo.UpdateSong(c.Request.Context(), &su, songId)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Message{Ok: false, Msg: "something went wrong"})
		log.Panic(err.Error())
		return
	}
	c.JSON(
		http.StatusOK,
		models.Message{
			Ok:  true,
			Msg: "updated",
		},
	)
	log.Debug("Songs updated", &su)
}

// GetSongText godoc
//
//	@Summary		Retrieve song text by ID
//	@Description	Fetches the text of a song given its ID, along with pagination details.
//	@Tags			Songs
//	@Produce		json
//	@Param			id		path		int					true	"Song ID"
//	@Param			page	query		int					false	"Page number for pagination"
//	@Param			max		query		int					false	"Maximum number of items per page"
//	@Success		200		{object}	models.SongsText	"Successful response containing song text"
//	@Failure		404		{object}	models.Message		"Song not found"
//	@Failure		502		{object}	models.Message		"Internal error or invalid input"
//	@Router			/songs/{id}/text [get]
func (h *Handler) GetSongText(c *gin.Context) {
	songId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Message{Ok: false, Msg: err.Error()})
		return
	}

	pmq := models.NewPageMaxQuery()
	c.Bind(&pmq)

	exists, err := h.songsRepo.CheckIfExists(c.Request.Context(), songId)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Message{Ok: false, Msg: "something went wrong"})
		log.Panic(err.Error())
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, models.Message{Ok: false, Msg: "not found"})
		return
	}

	songText, amount, err := h.songsRepo.GetSongText(c.Request.Context(), songId, &pmq)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.Message{Ok: false, Msg: "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, models.Message{Ok: false, Msg: "something went wrong"})
		log.Panic(err.Error())
		return
	}
	c.JSON(http.StatusOK, models.SongsText{
		Data:   songText,
		Page:   pmq.Page,
		Ok:     true,
		Amount: amount,
		Next:   pmq.Max*(pmq.Page+1) < amount,
	})
}
