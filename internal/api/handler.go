package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"bytestream/internal/domain"
)

type Handler struct {
	svc *domain.Service
	cat interface{ Len() int }
}

func NewHandler(svc *domain.Service, cat interface{ Len() int }) *Handler {
	return &Handler{svc: svc, cat: cat}
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.cat.Len() == 0 {
		WriteJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "not_ready", Details: "catalog is empty"})
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *Handler) Playback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method_not_allowed"})
		return
	}

	token, err := bearerToken(r.Header.Get("Authorization"))
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Details: err.Error()})
		return
	}

	videoID, err := videoIDFromPath(r.URL.Path)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Details: "invalid video_id"})
		return
	}

	resp, status, details := h.svc.Resolve(r.Context(), token, videoID)
	if status == 200 {
		WriteJSON(w, http.StatusOK, resp)
		return
	}

	switch status {
	case 401:
		WriteJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Details: details})
	case 403:
		WriteJSON(w, http.StatusForbidden, ErrorResponse{Error: "not_available", Details: details})
	case 404:
		WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Details: details})
	case 502:
		WriteJSON(w, http.StatusBadGateway, ErrorResponse{Error: "upstream_error", Details: details})
	default:
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error"})
	}
}

func videoIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[0] != "playback" {
		return 0, errors.New("bad path")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func bearerToken(h string) (string, error) {
	if h == "" {
		return "", errors.New("missing Authorization header")
	}
	fields := strings.Fields(h)
	if len(fields) != 2 || strings.ToLower(fields[0]) != "bearer" || fields[1] == "" {
		return "", errors.New("expected: Authorization: bearer <token>")
	}
	return fields[1], nil
}
