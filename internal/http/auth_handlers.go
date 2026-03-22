package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"blackjack/internal/auth"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAuthError(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Password) < 6 {
		writeAuthError(w, "ник и пароль (мин. 6 символов)", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user != nil {
		writeAuthError(w, "ник уже занят", http.StatusConflict)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	userID, err := h.userRepo.Create(req.Username, hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.playerRepo.GetOrCreateForUser(userID, req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := auth.CreateToken(userID)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	setAuthCookie(w, token)
	WriteJSON(w, map[string]any{"ok": true, "user_id": userID})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAuthError(w, "invalid json", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeAuthError(w, "введите ник и пароль", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		writeAuthError(w, "неверный ник или пароль", http.StatusUnauthorized)
		return
	}

	token, err := auth.CreateToken(user.ID)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	setAuthCookie(w, token)

	player, _ := h.playerRepo.GetByUserID(user.ID)
	resp := map[string]any{"ok": true, "user_id": user.ID}
	if player != nil {
		resp["balance"] = player.Balance
		resp["name"] = player.Name
		resp["premium_chips"] = player.PremiumChips
	}

	WriteJSON(w, resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	WriteJSON(w, map[string]string{"ok": "true"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	userID, ok := ctx.Value(userIDKey).(int)
	if !ok {
		writeAuthError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	player, err := h.playerRepo.GetByUserID(userID)
	if err != nil || player == nil {
		writeAuthError(w, "player not found", http.StatusNotFound)
		return
	}

	WriteJSON(w, map[string]any{
		"user_id":       userID,
		"name":          player.Name,
		"balance":       player.Balance,
		"premium_chips": player.PremiumChips,
	})
}

func writeAuthError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   7 * 24 * 3600, // 7 days
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}
