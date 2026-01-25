package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/middleware"
	"github.com/zhedevops/shortlink/internal/model"
	"github.com/zhedevops/shortlink/internal/service"
	"github.com/zhedevops/shortlink/internal/storage"
)

func TestCreateShortLinkHandler(t *testing.T) {
	cnf := config.GetConfig()
	ms := storage.NewMemoryStorage()
	srv := service.NewService(ms)
	h := &Handler{service: srv, Cfg: cnf}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
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
	ms := storage.NewMemoryStorage()
	srv := service.NewService(ms)
	h := &Handler{service: srv}
	shortID := "ZMFazWTA"
	originalURL := "https://ria.ru/"
	ms.Store[shortID] = originalURL

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
	var target = "/api/shorten"
	respLink := cnf.ResponseAddr.ServerAddress + "/CSaEMooR"
	var resp = model.Response{
		Result: respLink,
	}
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	err := encoder.Encode(resp)
	require.NoError(t, err)
	ms := storage.NewMemoryStorage()
	srv := service.NewService(ms)
	h := &Handler{service: srv, Cfg: cnf}
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
