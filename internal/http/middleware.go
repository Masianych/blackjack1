package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"blackjack/internal/game"
	"blackjack/internal/repository"
	"blackjack/internal/store"
)

type contextKey string

const (
	gameKey     contextKey = "game"
	userIDKey   contextKey = "user_id"
	playerIDKey contextKey = "player_id"
)

func GetUserID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(userIDKey).(int)
	return id, ok
}

func GetPlayerID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(playerIDKey).(int)
	return id, ok
}

func GetGame(r *http.Request) *game.Game {

	return r.Context().Value(gameKey).(*game.Game)
}

func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {

	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}

func Logging(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func Recover(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {

			if err := recover(); err != nil {

				http.Error(w, "internal error", 500)

			}

		}()

		next.ServeHTTP(w, r)
	})
}

func RequireJSON(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost &&
			r.URL.Path != "/api/hit" &&
			r.URL.Path != "/api/stand" &&
			r.URL.Path != "/api/logout" &&
			r.URL.Path != "/api/buy-chips" {

			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "content-type must be application/json", 415)
				return
			}

		}

		next.ServeHTTP(w, r)
	})
}

func Session(store *store.MemoryStore, playerRepo *repository.PlayerRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(userIDKey).(int)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			id := "user_" + strconv.Itoa(userID)
			g, ok := store.Get(id)
			if !ok {
				player, err := playerRepo.GetByUserID(userID)
				balance := 1000
				if err == nil && player != nil {
					balance = player.Balance
				}
				g = &game.Game{Balance: balance}
				store.Set(id, g)
			}

			ctx := context.WithValue(r.Context(), gameKey, g)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WriteJSON(w http.ResponseWriter, data any) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)
}
