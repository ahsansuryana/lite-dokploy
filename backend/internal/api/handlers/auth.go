package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lite-dokploy/backend/internal/api/middleware"
	"github.com/lite-dokploy/backend/internal/auth"
)

type AuthHandler struct {
	manager *auth.Manager
}

func NewAuthHandler(manager *auth.Manager) *AuthHandler {
	return &AuthHandler{manager: manager}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if !h.manager.Authenticate(req.Username, req.Password) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.manager.GenerateToken(req.Username)
	if err != nil {
		http.Error(w, "generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	username, _ := r.Context().Value(middleware.CtxUsername).(string)
	json.NewEncoder(w).Encode(map[string]string{"username": username})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
