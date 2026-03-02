package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/model"
	"github.com/zhedevops/shortlink/internal/service"
)

type Handler struct {
	service *service.Service
	Cfg     *config.Config
}

func NewHandler(s *service.Service, cnf *config.Config) *Handler {
	return &Handler{
		service: s,
		Cfg:     cnf,
	}
}

func (h *Handler) CreateShortLinkHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "handleCookie_failure", err.Error())
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var shortys = model.NewShortys("", string(body), "", user.ID)
	link, err := h.service.CreateShortLink(shortys)
	if err != nil {
		if errors.Is(err, model.ErrConflict) {
			h.setErrorResponseOnConflict(w, link)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "service_CreateShortLink_failure", err.Error())
		return
	}
	resp := h.Cfg.ResponseAddr.ServerAddress + "/" + link.ShortURL
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(resp))
	if err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}

func (h *Handler) CreateShortLinkEncHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "handleCookie_failure", err.Error())
		return
	}
	var req model.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "decode_body_failure", "cannot decode request JSON body")
		return
	}
	defer r.Body.Close()
	var shortys = model.NewShortys("", req.URL, "", user.ID)
	link, err := h.service.CreateShortLink(shortys)
	if err != nil {
		if errors.Is(err, model.ErrConflict) {
			h.setShortenErrorResponseOnConflict(w, link)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "service_CreateShortLink_failure", err.Error())
		return
	}
	respLink := h.Cfg.ResponseAddr.ServerAddress + "/" + link.ShortURL
	var resp = model.Response{
		Result: respLink,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	if err = encoder.Encode(resp); err != nil {
		log.Error().Err(err).Msg("error encoding response")
	}
}

func (h *Handler) CreateShortLinkBatchHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "handleCookie_failure", err.Error())
		return
	}
	var req []model.RequestBatch
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	resp := []model.ResponseBatch{}
	for _, rb := range req {
		var shortys = model.NewShortys(rb.CorrelationID, rb.OriginalURL, "", user.ID)
		link, err := h.service.CreateShortLink(shortys)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "service_CreateShortLink_failure", err.Error())
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

func (h *Handler) GetLinkByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	urlStr, err := h.service.GetOriginalURL(id)
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
}

func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.service.Ping(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "service_Ping_failure", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) setErrorResponseOnConflict(w http.ResponseWriter, link *model.Shorty) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusConflict)

	resp := h.Cfg.ResponseAddr.ServerAddress + "/" + link.ShortURL
	_, err := w.Write([]byte(resp))
	if err != nil {
		log.Error().Err(err).Msg("failed to write response")
	}
}

func (h *Handler) setShortenErrorResponseOnConflict(w http.ResponseWriter, link *model.Shorty) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)
	var resp = model.Response{
		Result: h.Cfg.ResponseAddr.ServerAddress + "/" + link.ShortURL,
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resp); err != nil {
		log.Error().Err(err).Msg("error encoding response")
	}
}

func (h *Handler) UserLinksHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "handleCookie_failure", err.Error())
		return
	}
	links, err := h.service.GetUserLinks(user.ID)
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
		_ = h.service.DeleteLinks(shortURLs, user.ID)
	}()
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) handleCookie(w http.ResponseWriter, r *http.Request) (model.User, error) {
	cookieAuth, err := r.Cookie("Authorization")
	var user = model.User{}
	needCreate := err != nil

	if !needCreate {
		user, err = h.service.CheckAuthCookie(cookieAuth)
		if err != nil {
			needCreate = true
		}
	}
	if needCreate {
		user, err = h.service.GetNewUser()
		if err != nil {
			return user, err
		}
		ac := h.service.GetAuthCookie(user)
		http.SetCookie(w, &http.Cookie{
			Name:     "Authorization",
			Value:    ac,
			Path:     "/",
			HttpOnly: true,
		})
	}
	return user, nil
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
