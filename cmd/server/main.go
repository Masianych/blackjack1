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
	db, err := database.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal("migrate:", err)
	}

	playerRepo := repository.NewPlayerRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	userRepo := repository.NewUserRepository(db)
	gameStore := store.NewMemoryStore()

	handler := apphttp.NewHandler(gameStore, playerRepo, historyRepo, userRepo)

	mux := http.NewServeMux()

	// Public routes (no auth)
	mux.HandleFunc("/api/register", handler.Register)
	mux.HandleFunc("/api/login", handler.Login)
	mux.HandleFunc("/api/logout", handler.Logout)
	mux.HandleFunc("/api/leaderboard", handler.Leaderboard)

	// Protected routes: Auth runs first, then Session, then handler
		// Порядок важен: Auth первый (проверяет JWT, ставит userID), затем Session (грузит game по userID)
		protectedChain := apphttp.Chain(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/state":
					handler.State(w, r)
				case "/api/bet":
					handler.Bet(w, r)
				case "/api/hit":
					handler.Hit(w, r)
				case "/api/stand":
					handler.Stand(w, r)
				case "/api/history":
					handler.History(w, r)
				case "/api/setname":
					handler.SetName(w, r)
				case "/api/me":
					handler.Me(w, r)
				case "/api/buy-chips":
					handler.BuyChips(w, r)
				default:
					http.NotFound(w, r)
				}
			}),
			apphttp.Auth(userRepo, playerRepo),
			apphttp.Session(gameStore, playerRepo),
		)

	mux.Handle("/api/state", protectedChain)
	mux.Handle("/api/bet", protectedChain)
	mux.Handle("/api/hit", protectedChain)
	mux.Handle("/api/stand", protectedChain)
	mux.Handle("/api/history", protectedChain)
	mux.Handle("/api/setname", protectedChain)
	mux.Handle("/api/me", protectedChain)
	mux.Handle("/api/buy-chips", protectedChain)

	mux.Handle("/", http.FileServer(http.Dir("./web")))

	handlerWithMiddleware := apphttp.Chain(
		mux,
		apphttp.Recover,
		apphttp.Logging,
		apphttp.RequireJSON,
	)

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handlerWithMiddleware))
}
