package postgresql

import (
	"context"

	"github.com/nikuma0/test-effective-mobile-golang/internal/models"
)

type SongsRepositoryI interface {
	GetSong(ctx context.Context, sdq *models.SongDetailQuery) (models.SongDetail, error)
	GetSongs(ctx context.Context, sq *models.SongsQuery) ([]models.Song, int, error)
	CreateSong(ctx context.Context, scq *models.SongCreateQuery) error
	UpdateSong(ctx context.Context, su *models.SongUpdate, songId int) error
	CheckIfExists(ctx context.Context, songId int) (bool, error)
	GetSongText(ctx context.Context, songId int, pmq *models.PageMaxQuery) ([]string, int, error)
}
