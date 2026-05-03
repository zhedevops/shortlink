package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhedevops/shortlink/internal/audit"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/database"
	"github.com/zhedevops/shortlink/internal/middleware"
	"github.com/zhedevops/shortlink/internal/mocks"
	"github.com/zhedevops/shortlink/internal/model"
	"github.com/zhedevops/shortlink/internal/service"
	"github.com/zhedevops/shortlink/internal/storage"
)

func TestCreateShortLinkHandler(t *testing.T) {
	cnf := config.GetConfig()
	fileName := "../../data/files/defaultpath/test.json"
	defer func() {
		_ = os.Remove(fileName)
	}()
	fs := storage.NewFileStorage(fileName)
	srv := service.NewService(fs, cnf)
	var sinks []audit.AuditSink
	auditSrv := audit.NewAuditService(sinks)
	h := &Handler{audit: auditSrv, service: srv, Cfg: &cnf.Server}
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.GzipHandle)
	r.With(middleware.RequireContentType("text/plain")).HandleFunc("/", h.CreateShortLinkHandler)

	type want struct {
		code        int
		response    string
		contentType string
	}
	type args struct {
		method      string
		target      string
		body        string
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "MethodPost result success",
			args: args{
				method:      http.MethodPost,
				target:      "/",
				body:        "https://ria.ru/",
				contentType: "text/plain",
			},
			want: want{
				code:        http.StatusCreated,
				response:    "ZMFazWTA",
				contentType: "text/plain",
			},
		},
		{
			name: "unsupported content type",
			args: args{
				method:      http.MethodPost,
				target:      "/",
				body:        "https://ria.ru/",
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "unsupported content type",
				contentType: "text/plain",
			},
		},
		{
			name: "empty url",
			args: args{
				method:      http.MethodPost,
				target:      "/",
				body:        "",
				contentType: "text/plain",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "empty url",
				contentType: "text/plain",
			},
		},
		{
			name: "invalid url",
			args: args{
				method:      http.MethodPost,
				target:      "/",
				body:        "https://ria ru/",
				contentType: "text/plain",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "invalid url",
				contentType: "text/plain",
			},
		},
		{
			name: "unsupported scheme",
			args: args{
				method:      http.MethodPost,
				target:      "/",
				body:        "ftp://ria.ru/",
				contentType: "text/plain",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "unsupported scheme",
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.target, strings.NewReader(tt.args.body))
			request.Header.Add("Content-Type", tt.args.contentType)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer func() {
				_ = res.Body.Close()
			}()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Contains(t, string(resBody), tt.want.response)
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)
		})
	}
}

func TestGetLinkByIDHandler(t *testing.T) {
	cnf := config.GetConfig()
	fileName := "../../data/files/defaultpath/test.json"
	defer func() {
		_ = os.Remove(fileName)
	}()
	fs := storage.NewFileStorage(fileName)
	srv := service.NewService(fs, cnf)
	var sinks []audit.AuditSink
	auditSrv := audit.NewAuditService(sinks)
	h := &Handler{audit: auditSrv, service: srv, Cfg: &cnf.Server}
	shortID := "ZMFazWTA"
	originalURL := "https://ria.ru/"
	shortys := &model.Shorty{
		OriginalURL: "https://ria.ru/",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := srv.CreateShortLink(ctx, shortys)
	assert.Nil(t, err)

	r := chi.NewRouter()
	r.HandleFunc("/{id}", h.GetLinkByIDHandler)

	type want struct {
		code     int
		response string
		location string
	}
	type args struct {
		method      string
		target      string
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "MethodGet result success",
			args: args{
				method:      http.MethodGet,
				target:      "/" + shortID,
				contentType: "text/plain",
			},
			want: want{
				code:     http.StatusTemporaryRedirect,
				response: "",
				location: originalURL,
			},
		},
		{
			name: "unexpected length id",
			args: args{
				method:      http.MethodGet,
				target:      "/XXX",
				contentType: "text/plain",
			},
			want: want{
				code:     http.StatusBadRequest,
				response: "unexpected length id",
				location: "",
			},
		},
		{
			name: "url not found",
			args: args{
				method:      http.MethodGet,
				target:      "/XXXXXXXX",
				contentType: "text/plain",
			},
			want: want{
				code:     http.StatusBadRequest,
				response: "url not found",
				location: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.target, nil)
			request.Header.Add("Content-Type", tt.args.contentType)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer func() {
				_ = res.Body.Close()
			}()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Contains(t, string(resBody), tt.want.response)
			assert.Contains(t, res.Header.Get("Location"), tt.want.location)
		})
	}
}

func TestCreateShortLinkEncHandler(t *testing.T) {
	cnf := config.GetConfig()
	target := "/api/shorten"
	respLink := cnf.Server.ResponseAddr.ServerAddress + "/CSaEMooR"
	resp := model.Response{
		Result: respLink,
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(resp)
	require.NoError(t, err)
	fileName := "../../data/files/defaultpath/test.json"
	defer func() {
		_ = os.Remove(fileName)
	}()
	fs := storage.NewFileStorage(fileName)
	srv := service.NewService(fs, cnf)
	var sinks []audit.AuditSink
	auditSrv := audit.NewAuditService(sinks)
	h := &Handler{audit: auditSrv, service: srv, Cfg: &cnf.Server}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.With(middleware.RequireContentType("application/json")).HandleFunc(target, h.CreateShortLinkEncHandler)

	type want struct {
		code        int
		response    string
		err         string
		contentType string
	}
	type args struct {
		method      string
		target      string
		body        string
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "MethodPost result success",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `{"url": "https://practicum.yandex.ru"}`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusCreated,
				response:    buf.String(),
				err:         "",
				contentType: "application/json",
			},
		},
		{
			name: "unsupported content type",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `{"url": "https://practicum.yandex.ru"}`,
				contentType: "text/plain",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				err:         "unsupported content type",
				contentType: "text/plain",
			},
		},
		{
			name: "empty url",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `{"url": ""}`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "",
				err:         "empty url",
				contentType: "text/plain",
			},
		},
		{
			name: "invalid url",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `{"url": "https://practicum yandex ru"}`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "",
				err:         "invalid url",
				contentType: "text/plain",
			},
		},
		{
			name: "unsupported scheme",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `{"url": "ftp://practicum.yandex.ru"}`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "",
				err:         "unsupported scheme",
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.target, strings.NewReader(tt.args.body))
			request.Header.Add("Content-Type", tt.args.contentType)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer func() {
				_ = res.Body.Close()
			}()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)
			if tt.want.err != "" {
				assert.Contains(t, string(resBody), tt.want.err)
			} else {
				assert.JSONEq(t, tt.want.response, string(resBody))
			}
		})
	}
}

func TestHandler_PingHandler(t *testing.T) {
	a := assert.New(t)
	_ = godotenv.Load("../../.env")
	dsn, dsnErr := os.LookupEnv("DATABASE_DSN")
	if dsn == "" {
		t.Skip("dns is required")
	}
	a.True(dsnErr)
	cnf := config.GetConfig()
	// Открываем пул
	pool, err := database.ConnectDB(dsn)
	a.Nil(err)
	a.NotNil(pool)
	a.IsType(&pgxpool.Pool{}, pool)
	st := storage.NewDBStorage(pool)
	srv := service.NewService(st, cnf)
	var sinks []audit.AuditSink
	auditSrv := audit.NewAuditService(sinks)
	h := &Handler{audit: auditSrv, service: srv, Cfg: &cnf.Server}
	r := chi.NewRouter()
	r.HandleFunc("/ping", h.PingHandler)
	t.Run("Pool opened. Ping ok", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, request)

		res := w.Result()
		assert.Equal(t, http.StatusOK, res.StatusCode)
		defer func() {
			_ = res.Body.Close()
		}()
	})
	t.Run("Pool closed. Ping failure", func(t *testing.T) {
		// Удаляем пул
		database.CloseDB(pool)
		ctx := context.Background()
		err = pool.Ping(ctx)
		a.NotNil(err)

		request := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, request)

		res := w.Result()
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
		defer func() {
			_ = res.Body.Close()
		}()
	})
}

func TestHandler_CreateShortLinkBatchHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepository(ctrl)
	user := model.User{
		ID: 2,
	}
	value := &model.Shorty{}
	value2 := &model.Shorty{
		UUID:        "69cc5e9c-404e-47c3-b9cf-7222f0122e37",
		OriginalURL: "http://zgvx7h.ru",
		ShortURL:    "BGTHakFB",
		UserID:      user.ID,
	}
	m.EXPECT().CheckIDByURL("http://dlf82a5xunr.net/vmzsxxp").Return(value)
	m.EXPECT().CheckIDByURL("http://rk2trgcml.biz/rltva/sklvun/m2u0jhvdvv3epe").Return(value)
	m.EXPECT().CheckIDByURL("http://zgvx7h.ru").Return(value).Times(1)
	m.EXPECT().CheckIDByURL("http://zgvx7h.ru").Return(value2).Times(1)
	m.EXPECT().CheckIDByURL("http://qpsh6hy.biz").Return(value).Times(1)
	m.EXPECT().GetOriginalURL(gomock.Any(), "qknZDqRy").Return(value)
	m.EXPECT().GetOriginalURL(gomock.Any(), "HLYMhqfn").Return(value)
	m.EXPECT().GetOriginalURL(gomock.Any(), "BGTHakFB").Return(value)
	m.EXPECT().GetOriginalURL(gomock.Any(), "npDieteQ").Return(value)
	shortys := model.AddShortys("d51eae65-0408-4d2d-997d-989f77f26e71", "http://dlf82a5xunr.net/vmzsxxp", "qknZDqRy", user.ID)
	shortys2 := model.AddShortys("6200fd8b-a597-4167-97b9-7a6323117bc4", "http://rk2trgcml.biz/rltva/sklvun/m2u0jhvdvv3epe", "HLYMhqfn", user.ID)
	shortys3 := model.AddShortys("69cc5e9c-404e-47c3-b9cf-7222f0122e37", "http://zgvx7h.ru", "BGTHakFB", user.ID)
	shortys4 := model.AddShortys("8542f426-e340-45d7-b577-b36d5f08aee6", "http://qpsh6hy.biz", "npDieteQ", user.ID)
	m.EXPECT().SetShortURL(gomock.Any(), shortys).Return(nil)
	m.EXPECT().SetShortURL(gomock.Any(), shortys2).Return(nil)
	m.EXPECT().SetShortURL(gomock.Any(), shortys3).Return(nil).Times(1)
	m.EXPECT().SetShortURL(gomock.Any(), shortys4).Return(errors.New("db error")).Times(1)
	target := "/api/shorten/batch"
	cnf := config.GetConfig()
	srv := service.NewService(m, cnf)
	var sinks []audit.AuditSink
	auditSrv := audit.NewAuditService(sinks)
	h := &Handler{audit: auditSrv, service: srv, Cfg: &cnf.Server}
	ac := h.service.GetAuthCookie(user)
	cookie := &http.Cookie{
		Name:     "Authorization",
		Value:    ac,
		Path:     "/",
		HttpOnly: true,
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.With(middleware.RequireContentType("application/json")).HandleFunc(target, h.CreateShortLinkBatchHandler)

	type want struct {
		code        int
		response    string
		err         string
		contentType string
	}
	type args struct {
		method      string
		target      string
		body        string
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "MethodPost result success",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `[{"correlation_id":"d51eae65-0408-4d2d-997d-989f77f26e71","original_url":"http://dlf82a5xunr.net/vmzsxxp"},{"correlation_id":"6200fd8b-a597-4167-97b9-7a6323117bc4","original_url":"http://rk2trgcml.biz/rltva/sklvun/m2u0jhvdvv3epe"}]`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusCreated,
				response:    `[{"correlation_id":"d51eae65-0408-4d2d-997d-989f77f26e71","short_url":"http://localhost:8080/qknZDqRy"},{"correlation_id":"6200fd8b-a597-4167-97b9-7a6323117bc4","short_url":"http://localhost:8080/HLYMhqfn"}]`,
				err:         "",
				contentType: "application/json",
			},
		},
		{
			name: "unsupported content type",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `[{"correlation_id":"d51eae65-0408-4d2d-997d-989f77f26e71","original_url":"http://dlf82a5xunr.net/vmzsxxp"},{"correlation_id":"6200fd8b-a597-4167-97b9-7a6323117bc4","original_url":"http://rk2trgcml.biz/rltva/sklvun/m2u0jhvdvv3epe"}]`,
				contentType: "text/plain",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				err:         "unsupported content type",
				contentType: "text/plain",
			},
		},
		{
			name: "invalid json",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `{"correlation_id":"d51eae65-0408-4d2d-997d-989f77f26e71","original_url":"http://dlf82a5xunr.net/vmzsxxp"},{"correlation_id":"6200fd8b-a597-4167-97b9-7a6323117bc4","original_url":"http://rk2trgcml.biz/rltva/sklvun/m2u0jhvdvv3epe"}`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				err:         "cannot decode request JSON body",
				contentType: "text/plain",
			},
		},
		{
			name: "repeated",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `[{"correlation_id":"69cc5e9c-404e-47c3-b9cf-7222f0122e37","original_url":"http://zgvx7h.ru"},{"correlation_id":"69cc5e9c-404e-47c3-b9cf-7222f0122e37","original_url":"http://zgvx7h.ru"}]`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusCreated,
				response:    `[{"correlation_id":"69cc5e9c-404e-47c3-b9cf-7222f0122e37","short_url":"http://localhost:8080/BGTHakFB"}, {"correlation_id":"69cc5e9c-404e-47c3-b9cf-7222f0122e37","short_url":"http://localhost:8080/BGTHakFB"}]`,
				err:         "",
				contentType: "application/json",
			},
		},
		{
			name: "error creating shorty",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `[{"correlation_id":"8542f426-e340-45d7-b577-b36d5f08aee6","original_url":"http://qpsh6hy.biz"},{"correlation_id":"6200fd8b-a597-4167-97b9-7a6323117bc4","original_url":"http://rk2trgcml.biz/rltva/sklvun/m2u0jhvdvv3epe"}]`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "",
				err:         "failed set short link: db error",
				contentType: "text/plain",
			},
		},
		{
			name: "empty url",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `[{"correlation_id":"d51eae65-0408-4d2d-997d-989f77f26e71"}]`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusInternalServerError,
				response:    "",
				err:         "empty url",
				contentType: "text/plain",
			},
		},
		{
			name: "empty batch",
			args: args{
				method:      http.MethodPost,
				target:      target,
				body:        `[]`,
				contentType: "application/json",
			},
			want: want{
				code:        http.StatusCreated,
				response:    `[]`,
				err:         "",
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.target, strings.NewReader(tt.args.body))
			request.Header.Add("Content-Type", tt.args.contentType)
			request.AddCookie(cookie)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer func() {
				_ = res.Body.Close()
			}()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)
			if tt.want.err != "" {
				assert.Contains(t, string(resBody), tt.want.err)
			} else {
				assert.JSONEq(t, tt.want.response, string(resBody))
			}
		})
	}
}

type fakeRepo struct{}

func (f *fakeRepo) GetOriginalURL(ctx context.Context, id string) *model.Shorty {
	_ = ctx
	_ = id
	return &model.Shorty{}
}

func (f *fakeRepo) CheckIDByURL(url string) *model.Shorty {
	_ = url
	shortys := &model.Shorty{
		UUID:        "d51eae65-0408-4d2d-997d-989f77f26e71",
		OriginalURL: "http://dlf82a5xunr.net/vmzsxxp",
		ShortURL:    "qknZDqRy",
		UserID:      2,
	}
	return shortys
}

func (f *fakeRepo) SetShortURL(ctx context.Context, shortys *model.Shorty) error {
	_ = ctx
	_ = shortys
	return nil
}

func (f *fakeRepo) Ping(ctx context.Context) error {
	_ = ctx
	return nil
}

func (f *fakeRepo) CreateUser(ctx context.Context) (model.User, error) {
	_ = ctx
	return model.User{}, nil
}

func (f *fakeRepo) GetShortysByUser(ctx context.Context, userID uint32) ([]*model.Shorty, error) {
	_ = ctx
	_ = userID
	return []*model.Shorty{}, nil
}

func (f *fakeRepo) DeleteLinks(ctx context.Context, ids []string, userID uint32) error {
	_ = ctx
	_ = ids
	_ = userID
	return nil
}

// Пример создания короткой ссылки.
func ExampleHandler_CreateShortLinkHandler() {
	// Создаём пользователя
	user := model.User{ID: 2}
	// Создаём обработчик с зависимостями
	cnf := config.GetConfig()
	// Для примера используем фейковый репозиторий, метод которого CheckIDByURL будет возвращать такой ответ:
	//  model.Shorty{
	//		UUID:        "d51eae65-0408-4d2d-997d-989f77f26e71",
	//		OriginalURL: "http://dlf82a5xunr.net/vmzsxxp",
	//		ShortURL:    "qknZDqRy",
	//		UserID:      2,
	//	}
	srv := service.NewService(&fakeRepo{}, cnf)
	var sinks []audit.AuditSink
	auditSrv := audit.NewAuditService(sinks)
	h := &Handler{audit: auditSrv, service: srv, Cfg: &cnf.Server}
	// Создаём cookie для пользователя
	ac := h.service.GetAuthCookie(user)
	cookie := &http.Cookie{
		Name:     "Authorization",
		Value:    ac,
		Path:     "/",
		HttpOnly: true,
	}

	// Создаём запрос
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("http://dlf82a5xunr.net/vmzsxxp"))
	// Добавляем в запрос cookie
	req.AddCookie(cookie)
	w := httptest.NewRecorder()

	// Вызываем метод обработчика
	h.CreateShortLinkHandler(w, req)

	resp := w.Result()

	fmt.Println(resp.StatusCode)

	// Output:
	// 201
}
