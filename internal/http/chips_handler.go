package http

import (
	"net/http"

	"blackjack/internal/repository"
)

func (h *Handler) BuyChips(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	playerID, ok := r.Context().Value(playerIDKey).(int)
	if !ok {
		writeAuthError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	g := GetGame(r)
	g.Lock()
	defer g.Unlock()

	if g.Balance < repository.PremiumChipsPrice {
		writeAuthError(w, "недостаточно средств (нужно $1200)", http.StatusBadRequest)
		return
	}

	player, err := h.playerRepo.GetByUserID(r.Context().Value(userIDKey).(int))
	if err != nil || player == nil {
		writeAuthError(w, "player not found", http.StatusNotFound)
		return
	}
	if player.PremiumChips {
		writeAuthError(w, "у вас уже есть красивые фишки", http.StatusBadRequest)
		return
	}

	g.Balance -= repository.PremiumChipsPrice
	if err := h.playerRepo.BuyPremiumChips(playerID, player.Balance); err != nil {
		g.Balance += repository.PremiumChipsPrice // откат
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.playerRepo.UpdateBalance(playerID, g.Balance)

	WriteJSON(w, map[string]any{
		"ok":            true,
		"balance":       g.Balance,
		"premium_chips": true,
	})
}
