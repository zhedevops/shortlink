package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()
	var shortys = model.NewShortys("", string(body), "", user.ID)
	link, err := h.service.CreateShortLink(shortys)
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
	_, err = w.Write([]byte(resp))
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func (h *Handler) CreateShortLinkEncHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req model.Request
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()
	var shortys = model.NewShortys("", req.URL, "", user.ID)
	link, err := h.service.CreateShortLink(shortys)
	if err != nil {
		if errors.Is(err, model.ErrConflict) {
			h.setShortenErrorResponseOnConflict(w, link)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		log.Printf("error encoding response: %v", err)
	}
}

func (h *Handler) CreateShortLinkBatchHandler(w http.ResponseWriter, r *http.Request) {
	user, err := h.handleCookie(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req []model.RequestBatch
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()
	resp := []model.ResponseBatch{}
	for _, rb := range req {
		var shortys = model.NewShortys(rb.CorrelationID, rb.OriginalURL, "", user.ID)
		link, err := h.service.CreateShortLink(shortys)
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
		log.Printf("error encoding response: %v", err)
	}
}

func (h *Handler) GetLinkByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	urlStr, err := h.service.GetOriginalURL(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", urlStr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.service.Ping(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		log.Printf("failed to write response: %v", err)
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
		log.Printf("error encoding response: %v", err)
	}
}

func (h *Handler) UserLinksHandler(w http.ResponseWriter, r *http.Request) {
	cookieAuth, _ := r.Cookie("Authorization")
	user, err := h.service.CheckAuthCookie(cookieAuth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	links, err := h.service.GetUserLinks(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	w.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(resp); err != nil {
		log.Printf("error encoding response: %v", err)
	}
}

func (h *Handler) handleCookie(w http.ResponseWriter, r *http.Request) (model.User, error) {
	cookieAuth, err := r.Cookie("Authorization")
	var user = model.User{}
	needCreate := false
	if err != nil {
		needCreate = true
	}
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
