package main

import (
	"log"
	"net/http"

	"blackjack/internal/database"
	apphttp "blackjack/internal/http"
	"blackjack/internal/repository"
	"blackjack/internal/store"
)

func main() {

	// подключение PostgreSQL
	db, err := database.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	// repositories
	playerRepo := repository.NewPlayerRepository(db)
	historyRepo := repository.NewHistoryRepository(db)

	// memory store для текущих игр
	gameStore := store.NewMemoryStore()

	// handler
	handler := apphttp.NewHandler(gameStore, playerRepo, historyRepo)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/leaderboard", handler.Leaderboard)
	mux.HandleFunc("/api/state", handler.State)
	mux.HandleFunc("/api/bet", handler.Bet)
	mux.HandleFunc("/api/hit", handler.Hit)
	mux.HandleFunc("/api/stand", handler.Stand)
	mux.HandleFunc("/api/history", handler.History)
	mux.HandleFunc("/api/setname", handler.SetName)
	// static frontend
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// middleware chain
	handlerWithMiddleware := apphttp.Chain(
		mux,
		apphttp.Recover,
		apphttp.Logging,
		apphttp.RequireJSON,
		apphttp.Session(gameStore),
	)

	log.Println("server running on :8080")

	log.Fatal(http.ListenAndServe(":8080", handlerWithMiddleware))
}
