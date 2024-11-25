package postgresql_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path"
	"testing"
	"time"

	"github.com/DATA-DOG/go-txdb"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nikuma0/test-effective-mobile-golang/internal/models"
	"github.com/nikuma0/test-effective-mobile-golang/internal/repository/postgresql"
	"github.com/nikuma0/test-effective-mobile-golang/internal/utils"
)

const (
	testDSN  = "host=localhost user=postgres password=postgres dbname=testdb sslmode=disable"
	basePath = "../../../"
)

func runMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+path.Join(basePath, "migrations"),
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("failed to initialize migrations: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to apply migrations: %v", err)
	}
}

func initFixtures(t *testing.T, db *sql.DB) {
	r, err := os.ReadFile(path.Join(basePath, "tests/fixtures/songs.sql"))
	if err != nil {
		log.Fatalf("failed to open fixture file: %v", err)
	}
	if _, err = db.Exec(string(r)); err != nil {
		log.Fatalf("failed to apply fixture file: %v", err)
	}
	t.Cleanup(func() { cleanFixtures(db) })
}

func cleanFixtures(db *sql.DB) {
	if _, err := db.Exec(`DELETE FROM songs`); err != nil {
		log.Fatalf("failed to apply fixture file: %v", err)
	}
}

func TestMain(m *testing.M) {
	txdb.Register("txdb", "postgres", testDSN)

	code := m.Run()

	os.Exit(code)
}

func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("txdb", "unique-connection-id")
	runMigrations(db)
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

func initHelper(t *testing.T, load bool) (db *sql.DB) {
	db = newTestDB(t)
	if load {
		initFixtures(t, db)
	}
	return
}

func initRepo(t *testing.T, db *sql.DB) postgresql.SongsRepository {
	repo := *postgresql.NewSongsRepository(db)
	tr, err := repo.Begin()
	if err != nil {
		log.Fatal(err)
	}
	t.Cleanup(func() { tr.Rollback() })
	return repo
}

func compareDates(t *testing.T, excepted, actual time.Time) {
	assert.Equal(t, excepted.Format("2006-01-02"), actual.Format("2006-01-02"))
}

func TestCreateSong(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	createSongCases := []struct {
		song          *models.SongCreateQuery
		returnErr     bool
		errMsgAndArgs []interface{}
		name          string
		getSong       *models.SongDetailQuery
	}{
		{
			name: "Simple",
			song: &models.SongCreateQuery{
				Song:        "Test Song",
				Group:       "Test Group",
				Text:        "Some lyrics",
				ReleaseDate: utils.Ptr(models.DateFormat(time.Now())),
				Link:        "https://example.com",
			},
			getSong: &models.SongDetailQuery{
				Group: "Test Group",
				Song:  "Test Song",
			},
		},
		{
			name: "CreateSongWithNoReleaseDate",
			song: &models.SongCreateQuery{
				Group: "Test Group",
				Song:  "Test Song",
				Text:  "Some lyrics",
				Link:  "https://example.com",
			},
			getSong: &models.SongDetailQuery{
				Group: "Test Group",
				Song:  "Test Song",
			},
		},
	}

	for _, tc := range createSongCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := initRepo(t, db)
			err := repo.CreateSong(ctx, tc.song)
			if tc.returnErr {
				require.Error(t, err, tc.errMsgAndArgs...)
				return
			} else {
				require.NoError(t, err)
			}
			song, err := repo.GetSong(ctx, tc.getSong)
			require.NoError(t, err)
			assert.Equal(t, tc.song.Song, song.Name)
			assert.Equal(t, tc.song.Group, song.GroupName)
			if tc.song.ReleaseDate != nil {
				compareDates(t, time.Time(*tc.song.ReleaseDate), time.Time(song.ReleaseDate))
			} else {
				compareDates(t, time.Now(), time.Time(song.ReleaseDate))
			}
		})
	}

}

func TestGetSong(t *testing.T) {
	db := initHelper(t, true)

	t.Run("GetSong", func(t *testing.T) {
		repo := initRepo(t, db)
		ctx := context.Background()
		song, err := repo.GetSong(ctx, &models.SongDetailQuery{Group: "Group 2", Song: "Song 1"})
		require.NoError(t, err)

		assert.Equal(t, "Group 2", song.GroupName)
		assert.Equal(t, "Song 1", song.Name)
	})

	t.Run("GetNotExistsSong", func(t *testing.T) {
		repo := initRepo(t, db)
		ctx := context.Background()
		_, err := repo.GetSong(ctx, &models.SongDetailQuery{Group: "Never existed group", Song: "Never existed song"})
		require.Error(t, err, sql.ErrNoRows)
	})

	t.Run("Pagination", func(t *testing.T) {
		repo := initRepo(t, db)
		ctx := context.Background()
		songs, _, err := repo.GetSongs(ctx, &models.SongsQuery{Max: 5, Page: 0})
		require.NoError(t, err)
		assert.Len(t, songs, 5)
	})
}

func TestCheckIfExists(t *testing.T) {
	db := initHelper(t, true)
	t.Run("ExistingSong", func(t *testing.T) {
		repo := initRepo(t, db)
		ctx := context.Background()
		exists, err := repo.CheckIfExists(ctx, 2)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("NonExistingSong", func(t *testing.T) {
		repo := initRepo(t, db)
		ctx := context.Background()
		exists, err := repo.CheckIfExists(ctx, 999)
		require.NoError(t, err)
		assert.False(t, exists)
	})

}

func TestUpdateSong(t *testing.T) {
	db := initHelper(t, true)
	t.Run("UpdateAllFields", func(t *testing.T) {
		repo := initRepo(t, db)
		ctx := context.Background()
		songUpdate := &models.SongUpdate{
			Name:        utils.Ptr("Updated Song"),
			GroupName:   utils.Ptr("Updated Group"),
			Text:        utils.Ptr("Updated lyrics"),
			ReleaseDate: utils.Ptr(models.DateFormat(time.Now().AddDate(0, 0, -1))),
		}
		err := repo.UpdateSong(ctx, songUpdate, 1)
		require.NoError(t, err)

		updatedSong, err := repo.GetSong(ctx, &models.SongDetailQuery{Group: "Updated Group", Song: "Updated Song"})
		require.NoError(t, err)

		assert.Equal(t, "Updated Song", updatedSong.Name)
		assert.Equal(t, "Updated Group", updatedSong.GroupName)
		assert.Equal(t, "Updated lyrics", updatedSong.Text)
		compareDates(t, time.Time(*songUpdate.ReleaseDate), time.Time(updatedSong.ReleaseDate))
	})

	t.Run("UpdatePartialFields", func(t *testing.T) {
		repo := initRepo(t, db)
		ctx := context.Background()
		songUpdate := &models.SongUpdate{
			Name: utils.Ptr("Partially Updated Song"),
		}
		err := repo.UpdateSong(ctx, songUpdate, 1)
		require.NoError(t, err)

		updatedSong, err := repo.GetSong(ctx, &models.SongDetailQuery{Group: "Group 2", Song: "Partially Updated Song"})
		require.NoError(t, err)

		assert.Equal(t, "Partially Updated Song", updatedSong.Name)
		assert.Equal(t, "Group 2", updatedSong.GroupName)
	})
}

func TestGetSongText(t *testing.T) {
	db := initHelper(t, true)

	t.Run("GetSongTextWithPagination", func(t *testing.T) {
		repo := initRepo(t, db)
		ctx := context.Background()
		pageQuery := &models.PageMaxQuery{
			Page: 0,
			Max:  2,
		}
		lines, total, err := repo.GetSongText(ctx, 1, pageQuery)
		require.NoError(t, err)

		assert.Equal(t, total, 1)
		assert.Len(t, lines, 1)
		assert.Equal(t, lines[0], "Lyrics for song 1")
	})
}
