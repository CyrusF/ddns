package server

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
)

type ReportRequest struct {
	Token string `json:"token"`
	IP    string `json:"ip"`
}

type ReportResponse struct {
	OK bool `json:"ok"`
}

type IPResponse struct {
	IP        string `json:"ip"`
	UpdatedAt string `json:"updated_at"`
}

type APIHandler struct {
	store *IPStore
	token string
}

func NewAPIHandler(store *IPStore, token string) *APIHandler {
	return &APIHandler{store: store, token: token}
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/report":
		h.handleReport(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/ip":
		h.handleGetIP(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *APIHandler) handleReport(w http.ResponseWriter, r *http.Request) {
	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Token != h.token {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "invalid token"})
		return
	}

	ip := req.IP
	if ip == "" {
		ip = extractIP(r)
	}
	if ip == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot determine client IP"})
		return
	}

	h.store.Set(ip)
	slog.Info("ip updated", "ip", ip)
	writeJSON(w, http.StatusOK, ReportResponse{OK: true})
}

func (h *APIHandler) handleGetIP(w http.ResponseWriter, r *http.Request) {
	ip, updatedAt, ok := h.store.Get()
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no ip registered"})
		return
	}
	writeJSON(w, http.StatusOK, IPResponse{
		IP:        ip,
		UpdatedAt: updatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
