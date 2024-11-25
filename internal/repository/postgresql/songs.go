package postgresql

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/nikuma0/test-effective-mobile-golang/internal/models"
)

type executor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type SongsRepository struct {
	pool executor
	db   *sql.DB
}

type Transaction struct {
	repo *SongsRepository
}

func NewSongsRepository(pool *sql.DB) *SongsRepository {
	return &SongsRepository{
		pool: pool,
		db:   pool,
	}
}

func (sr *SongsRepository) Begin() (tr *Transaction, err error) {
	tx, err := sr.db.Begin()
	if err != nil {
		return
	}
	sr.pool = tx
	tr = &Transaction{
		repo: sr,
	}
	return
}

func (tr *Transaction) Commit() error {
	err := tr.repo.pool.(*sql.Tx).Commit()
	tr.repo.pool = tr.repo.db
	return err
}

func (tr *Transaction) Rollback() error {
	err := tr.repo.pool.(*sql.Tx).Rollback()
	tr.repo.pool = tr.repo.db
	return err
}

func (sr *SongsRepository) GetSong(ctx context.Context, sdq *models.SongDetailQuery) (models.SongDetail, error) {
	var sm models.SongDetail
	var textLen int
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := sr.pool.QueryRowContext(
		ctx,
		`
		SELECT s.id, s.name, s.group_name, substring(s.text for 1024), release_date, char_length(s.text) FROM songs s
		WHERE s.name = $1 AND s.group_name = $2
		`,
		sdq.Song,
		sdq.Group,
	)
	err := row.Scan(&sm.Id, &sm.Name, &sm.GroupName, &sm.Text, &sm.ReleaseDate, &textLen)
	if textLen > len(sm.Text) {
		sm.Text += "..."
	}
	return sm, err
}

func (sr *SongsRepository) GetSongs(ctx context.Context, sq *models.SongsQuery) (res []models.Song, amount int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var releaseDate any
	if sq.ReleaseDate != nil {
		releaseDate = sq.ReleaseDate
	} else {
		releaseDate = sql.NullTime{}
	}

	row := sr.pool.QueryRowContext(
		ctx,
		`
		SELECT count(*) FROM songs s
		WHERE (s.name = $1 OR $1 IS NULL) AND (s.group_name = $2 OR $2 IS NULL) AND (s.release_date = $3 OR $3 IS NULL)
		`,
		sq.Song, sq.Group, releaseDate,
	)
	if err = row.Scan(&amount); err != nil || amount == 0 {
		return
	}

	rows, err := sr.pool.QueryContext(
		ctx,
		`
		SELECT s.id, s.name, s.group_name, s.release_date FROM songs s
		WHERE (s.name = $1 OR $1 IS NULL) AND (s.group_name = $2 OR $2 IS NULL) AND (s.release_date = $3 OR $3 IS NULL)
		LIMIT $4
		OFFSET $5
		`,
		sq.Song, sq.Group, releaseDate, sq.Max, sq.Max*sq.Page,
	)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var song models.Song
		if err = rows.Scan(&song.Id, &song.Name, &song.GroupName, &song.ReleaseDate); err != nil {
			return
		}
		res = append(res, song)
	}
	return
}

func (sr *SongsRepository) CreateSong(ctx context.Context, scq *models.SongCreateQuery) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	stmt := `
	INSERT INTO songs (name, group_name, text, release_date) VALUES ($1, $2, $3, $4);
	`
	args := []any{scq.Song, scq.Group, scq.Text, scq.ReleaseDate}
	if scq.ReleaseDate == nil {
		stmt = `
		INSERT INTO songs (name, group_name, text) VALUES ($1, $2, $3);
		`
		args = args[:len(args)-1]
	}

	_, err := sr.pool.ExecContext(
		ctx,
		stmt,
		args...,
	)
	return err
}

func (sr *SongsRepository) CheckIfExists(ctx context.Context, songId int) (exists bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := sr.pool.QueryRowContext(
		ctx, `SELECT EXISTS(SELECT 1 FROM songs s WHERE s.id = $1)`,
		songId,
	)
	err = row.Scan(&exists)
	return
}

func (sr *SongsRepository) UpdateSong(ctx context.Context, su *models.SongUpdate, songId int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fields := make(map[string]any)
	var setStmt string

	if su.GroupName != nil {
		fields["group_name"] = su.GroupName
	}
	if su.Name != nil {
		fields["name"] = su.Name
	}
	if su.Text != nil {
		fields["text"] = su.Text
	}
	if su.ReleaseDate != nil {
		fields["release_date"] = su.ReleaseDate
	}
	if len(fields) == 0 {
		return nil
	}

	counter := 0
	args := make([]any, len(fields)+1)
	for k, v := range fields {
		args[counter] = v
		counter++
		if setStmt != "" {
			setStmt += ", " + k + " = $" + strconv.Itoa(counter)
		} else {
			setStmt += k + " = $" + strconv.Itoa(counter)
		}
	}
	args[len(fields)] = songId
	_, err := sr.pool.ExecContext(
		ctx,
		`UPDATE songs SET `+setStmt+` WHERE id=$`+strconv.Itoa(len(fields)+1),
		args...,
	)
	return err
}

func (sr *SongsRepository) GetSongText(ctx context.Context, songId int, pmq *models.PageMaxQuery) (res []string, amount int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	row := sr.pool.QueryRowContext(
		ctx,
		`
		WITH split_text AS (
			SELECT
				UNNEST(STRING_TO_ARRAY(text, '\n')) AS line
			FROM songs
			WHERE id = $1
		)
		SELECT count(line)
		FROM split_text
		`,
		songId,
	)
	if err = row.Scan(&amount); err != nil {
		return
	}

	rows, err := sr.pool.QueryContext(
		ctx,
		`
		WITH split_text AS (
			SELECT
				UNNEST(STRING_TO_ARRAY(text, '\n')) AS line,
				generate_subscripts(STRING_TO_ARRAY(text, '\n'), 1) AS line_number
			FROM songs
			WHERE id = $1
		)
		SELECT line
		FROM split_text
		ORDER BY line_number
		OFFSET $2
		LIMIT $3
		`,
		songId,
		pmq.Page*pmq.Max,
		pmq.Max,
	)
	if err != nil {
		return
	}
	for rows.Next() {
		var line string
		if err = rows.Scan(&line); err != nil {
			return
		}
		res = append(res, line)
	}
	return
}
