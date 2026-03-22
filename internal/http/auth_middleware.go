package http

import (
	"context"
	"net/http"

	"blackjack/internal/auth"
	"blackjack/internal/repository"
)

func Auth(userRepo *repository.UserRepository, playerRepo *repository.PlayerRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				writeAuthError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := auth.ParseToken(cookie.Value)
			if err != nil {
				clearAuthCookie(w)
				writeAuthError(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			player, err := playerRepo.GetOrCreateForUser(claims.UserID, "Player")
			if err != nil {
				http.Error(w, "failed to load player", http.StatusInternalServerError)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, playerIDKey, player.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

