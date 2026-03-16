package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"blackjack/internal/game"
	"blackjack/internal/session"
	"blackjack/internal/store"
)

type contextKey string

const gameKey contextKey = "game"

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
			r.URL.Path != "/api/stand" {

			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "content-type must be application/json", 415)
				return
			}

		}

		next.ServeHTTP(w, r)
	})
}

func Session(store *store.MemoryStore) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("session_id")

			var id string
			var g *game.Game

			if err != nil {

				id = session.NewID()

				http.SetCookie(w, &http.Cookie{
					Name:     "session_id",
					Value:    id,
					Path:     "/",
					HttpOnly: true,
				})

				g = &game.Game{Balance: 1000}

				store.Set(id, g)

			} else {

				id = cookie.Value

				gameStored, ok := store.Get(id)

				if !ok {
					g = &game.Game{Balance: 1000}
					store.Set(id, g)
				} else {
					g = gameStored
				}
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
