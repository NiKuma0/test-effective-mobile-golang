package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	handlers "github.com/nikuma0/test-effective-mobile-golang/internal/http"
	"github.com/nikuma0/test-effective-mobile-golang/internal/models"
	"github.com/nikuma0/test-effective-mobile-golang/internal/repository/postgresql"
	"github.com/nikuma0/test-effective-mobile-golang/internal/utils"
)

type MockSongsRepository struct {
	mock.Mock
}

func (m *MockSongsRepository) Begin() (*postgresql.Transaction, error) {
	return &postgresql.Transaction{}, nil
}

func (m *MockSongsRepository) CheckIfExists(ctx context.Context, songId int) (bool, error) {
	args := m.Called(ctx, songId)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockSongsRepository) GetSongText(ctx context.Context, songId int, pmq *models.PageMaxQuery) ([]string, int, error) {
	args := m.Called(ctx, songId, pmq)
	return args.Get(0).([]string), args.Get(1).(int), args.Error(1)
}

func (m *MockSongsRepository) GetSong(ctx context.Context, sdq *models.SongDetailQuery) (models.SongDetail, error) {
	args := m.Called(ctx, sdq)
	return args.Get(0).(models.SongDetail), args.Error(1)
}

func (m *MockSongsRepository) GetSongsAmount(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockSongsRepository) GetSongs(ctx context.Context, sq *models.SongsQuery) ([]models.Song, int, error) {
	args := m.Called(ctx, sq)
	return args.Get(0).([]models.Song), args.Int(1), args.Error(2)
}

func (m *MockSongsRepository) CreateSong(ctx context.Context, scq *models.SongCreateQuery) error {
	args := m.Called(ctx, scq)
	return args.Error(0)
}

func (m *MockSongsRepository) UpdateSong(ctx context.Context, su *models.SongUpdate, songId int) error {
	args := m.Called(ctx, su)
	return args.Error(0)
}

func initHelper() (*gin.Engine, handlers.Handler, *MockSongsRepository) {
	mockRepo := new(MockSongsRepository)
	handler := handlers.NewTest(mockRepo)
	r := gin.Default()
	handler.Routes(r.Group(""))
	return r, handler, mockRepo
}

func TestCreateSong(t *testing.T) {
	simpleSongCreateQuery := models.SongCreateQuery{
		Group: "Group",
		Song:  "Song",
		Text:  "Text",
	}
	cases := []struct {
		name           string
		body           interface{}
		createSongData interface{}
		exceptedStatus int
		exceptedBody   string
		isRepoCalled   bool
	}{
		{
			name:           "Simple",
			body:           &simpleSongCreateQuery,
			createSongData: &simpleSongCreateQuery,
			exceptedStatus: http.StatusCreated,
			exceptedBody:   `{"ok":true,"msg":"created"}`,
			isRepoCalled:   true,
		},
		{
			name:           "InvalidJson",
			body:           `InvalidJson`,
			exceptedStatus: http.StatusBadRequest,
			exceptedBody:   `{"ok":false,"msg":"json: cannot unmarshal string into Go value of type models.SongCreateQuery"}`,
			isRepoCalled:   false,
		},
		{
			name: "NotAllRequiredFieldsProvided",
			body: `{
				"text": "text",
			}`,
			exceptedStatus: http.StatusBadRequest,
			exceptedBody:   `{"ok":false,"msg":"json: cannot unmarshal string into Go value of type models.SongCreateQuery"}`,
			isRepoCalled:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, _, mockRepo := initHelper()
			if tc.isRepoCalled {
				mockRepo.On("CreateSong", mock.Anything, tc.createSongData).Return(nil)
			}
			w := performRequestWithBody(r, "POST", "/songs", tc.body)
			assert.Equal(t, tc.exceptedStatus, w.Code)
			assert.Equal(t, tc.exceptedBody, w.Body.String())
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSongs(t *testing.T) {
	t.Run("Get all songs", func(t *testing.T) {
		r, _, mockRepo := initHelper()
		date := models.DateFormat(time.Date(2024, 11, 23, 0, 0, 0, 0, time.UTC))
		expectedSongs := []models.Song{
			{Id: 1, Name: "Song 1", GroupName: "Group 1", ReleaseDate: date},
			{Id: 2, Name: "Song 2", GroupName: "Group 1", ReleaseDate: date},
		}
		expectedAmount := 2

		mockRepo.On("GetSongs", mock.Anything, &models.SongsQuery{Page: 0, Max: 10, Group: utils.Ptr("Group 1")}).Return(expectedSongs, expectedAmount, nil)
		w := performRequest(r, "GET", "/songs?page=0&max=10&group=Group 1")

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Song 1")
		assert.Contains(t, w.Body.String(), "Song 2")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Create song", func(t *testing.T) {
		r, _, mockRepo := initHelper()
		newSong := &models.SongCreateQuery{
			Group:       "Group",
			Song:        "Song",
			Text:        "Text",
			ReleaseDate: utils.Ptr(models.DateFormat(time.Date(2022, 11, 8, 0, 0, 0, 0, time.UTC))),
		}
		mockRepo.On("CreateSong", mock.Anything, newSong).Return(nil)
		w := performRequestWithBody(r, "POST", "/songs", newSong)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), `"ok":true`)
		assert.Contains(t, w.Body.String(), `"msg":"created"`)
		mockRepo.AssertExpectations(t)
	})
}

func performRequest(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performRequestWithBody(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
