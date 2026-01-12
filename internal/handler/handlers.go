package handler

import (
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zhedevops/shortlink/internal/config"
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	link, err := h.service.CreateShortLink(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := h.Cfg.ResponseAddr.ServerAddress + "/" + link.ID
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(resp))
	if err != nil {
		log.Printf("failed to write response: %v", err)
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
