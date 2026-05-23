// Package handler Обработчик API запросов
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	guid "github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/zhedevops/shortlink/internal/audit"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/model"
	"github.com/zhedevops/shortlink/internal/service"
)

// Handler Тип обработчика
// generate:reset
type Handler struct {
	// audit Сервис аудита.
	audit *audit.AuditService
	// service Сервис, отвечающий за обработку запросов обработчика.
	service *service.Service
	// Cfg Конфигурация.
	Cfg *config.ServerConfig
}

// NewHandler Создаёт новый обработчик
func NewHandler(audit *audit.AuditService, srv *service.Service, cnf *config.ServerConfig) *Handler {
	return &Handler{
		audit:   audit,
		service: srv,
		Cfg:     cnf,
	}
}

// CreateShortLinkHandler Создаёт короткую ссылку для адреса
func (h *Handler) CreateShortLinkHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	shortys := model.NewShortys(guid.New().String(), string(body), user.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	link, err := h.service.CreateShortLink(ctx, shortys)
	if err != nil {
		if errors.Is(err, model.ErrConflict) {
			h.setErrorResponseOnConflict(w, link)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := h.Cfg.ResponseAddr.ServerAddress + "/" + link.ShortURL
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte(resp)); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
	h.audit.Send(audit.AuditEvent{
		UserID:    strconv.Itoa(int(user.ID)),
		Action:    audit.ActionShorten,
		URL:       link.OriginalURL,
		Timestamp: time.Now(),
	})
}

// CreateShortLinkEncHandler Создаёт короткую ссылку из запроса с json-телом
func (h *Handler) CreateShortLinkEncHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req model.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shortys := model.NewShortys(guid.New().String(), req.URL, user.ID)
	link, err := h.service.CreateShortLink(ctx, shortys)
	if err != nil {
		if errors.Is(err, model.ErrConflict) {
			h.setShortenErrorResponseOnConflict(w, link)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respLink := h.Cfg.ResponseAddr.ServerAddress + "/" + link.ShortURL
	resp := model.Response{
		Result: respLink,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	if err = encoder.Encode(resp); err != nil {
		log.Error().Err(err).Msg("error encoding response")
	}
	h.audit.Send(audit.AuditEvent{
		UserID:    strconv.Itoa(int(user.ID)),
		Action:    audit.ActionShorten,
		URL:       link.OriginalURL,
		Timestamp: time.Now(),
	})
}

// CreateShortLinkBatchHandler Осущестляет пакетную обработку запроса, принимая в теле запроса множество ссылок
func (h *Handler) CreateShortLinkBatchHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req []model.RequestBatch
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp := []model.ResponseBatch{}
	for _, rb := range req {
		shortys := model.NewShortys(rb.CorrelationID, rb.OriginalURL, user.ID)
		link, err := h.service.CreateShortLink(ctx, shortys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp = append(resp, model.ResponseBatch{
			CorrelationID: link.UUID,
			ShortURL:      h.Cfg.ResponseAddr.ServerAddress + "/" + link.ShortURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resp); err != nil {
		log.Error().Err(err).Msg("error encoding response")
	}
}

// GetLinkByIDHandler Получает оригинальную ссылку по короткой
func (h *Handler) GetLinkByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	id := chi.URLParam(r, "id")
	urlStr, err := h.service.GetOriginalURL(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrURLDeleted) {
			writeJSONError(w, http.StatusGone, "url_is_deleted", err.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, "service_GetOriginalURL_failure", err.Error())
		return
	}
	w.Header().Set("Location", urlStr)
	w.WriteHeader(http.StatusTemporaryRedirect)
	h.audit.Send(audit.AuditEvent{
		UserID:    "",
		Action:    audit.ActionFollow,
		URL:       urlStr,
		Timestamp: time.Now(),
	})
}

// PingHandler Осуществляет пинг сервера
func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.service.Ping(ctx); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "service_Ping_failure", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UserLinksHandler Получает все ссылки, сгенерированные пользователем
func (h *Handler) UserLinksHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "handleCookie_failure", err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	links, err := h.service.GetUserLinks(ctx, user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "service_GetUserLinks_failure", err.Error())
	}
	if links == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	resp := []model.ResponseUserLinks{}
	for _, s := range links {
		resp = append(resp, model.ResponseUserLinks{
			ShortURL:    h.Cfg.ResponseAddr.ServerAddress + "/" + s.ShortURL,
			OriginalURL: s.OriginalURL,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resp); err != nil {
		log.Error().Err(err).Msg("error encoding response")
	}
}

// DeleteLinkBatchHandler Осуществляет пакетное удаление оригинальных ссылок по полученным коротким ссылкам
func (h *Handler) DeleteLinkBatchHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "handleCookie_failure", err.Error())
		return
	}
	var shortURLs []string
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&shortURLs); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "decode_failure", "cannot decode request JSON body")
		return
	}
	defer r.Body.Close()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.service.DeleteLinks(ctx, shortURLs, user.ID)
	}()
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.service.GetStats(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "service_StatsHandler_failure", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resp); err != nil {
		log.Error().Err(err).Msg("error encoding response")
	}
}

func (h *Handler) handleCookie(w http.ResponseWriter, r *http.Request) (model.User, error) {
	cookieAuth, err := r.Cookie("Authorization")
	user := model.User{}
	needCreate := err != nil

	if !needCreate {
		user, err = h.service.CheckAuthCookie(cookieAuth)
		if err != nil {
			needCreate = true
		}
	}
	if needCreate {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		user, err = h.service.GetNewUser(ctx)
		if err != nil {
			return user, err
		}
		ac := h.service.GetAuthToken(user)
		http.SetCookie(w, &http.Cookie{
			Name:     "Authorization",
			Value:    ac,
			Path:     "/",
			HttpOnly: true,
		})
	}
	return user, nil
}

func (h *Handler) setErrorResponseOnConflict(w http.ResponseWriter, link *model.Shorty) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusConflict)

	resp := h.Cfg.ResponseAddr.ServerAddress + "/" + link.ShortURL
	if _, err := w.Write([]byte(resp)); err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}

func (h *Handler) setShortenErrorResponseOnConflict(w http.ResponseWriter, link *model.Shorty) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	resp := model.Response{
		Result: h.Cfg.ResponseAddr.ServerAddress + "/" + link.ShortURL,
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resp); err != nil {
		log.Error().Err(err).Msg("error encoding response")
	}
}

func writeJSONError(w http.ResponseWriter, status int, errCode, msg string) {
	err := model.ErrorResponse{
		Error:   errCode,
		Message: msg,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(err)
}
