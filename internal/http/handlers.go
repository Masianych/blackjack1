package http

import (
	"blackjack/internal/repository"
	"blackjack/internal/store"
	"encoding/json"
	"net/http"

	"blackjack/internal/game"
)

type Handler struct {
	store       *store.MemoryStore
	playerRepo  *repository.PlayerRepository
	historyRepo *repository.HistoryRepository
}
type Response struct {
	Player      []string `json:"player"`
	Dealer      []string `json:"dealer"`
	PlayerValue int      `json:"playerValue"`
	DealerValue int      `json:"dealerValue"`
	Balance     int      `json:"balance"`
	Bet         int      `json:"bet"`
	Result      string   `json:"result"`
	GameOver    bool     `json:"gameOver"`
}

func NewHandler(
	store *store.MemoryStore,
	playerRepo *repository.PlayerRepository,
	historyRepo *repository.HistoryRepository,
) *Handler {

	return &Handler{
		store:       store,
		playerRepo:  playerRepo,
		historyRepo: historyRepo,
	}
}

type BetRequest struct {
	Bet int `json:"bet"`
}

func cardsToStrings(cards []game.Card) []string {

	out := make([]string, len(cards))

	for i, c := range cards {
		out[i] = c.Rank + c.Suit
	}

	return out
}

func WriteGame(w http.ResponseWriter, g *game.Game, msg string) {

	dealerCards := cardsToStrings(g.Dealer)

	if !g.Reveal && len(dealerCards) > 0 {
		dealerCards[0] = "🂠"
	}

	resp := Response{
		Player:      cardsToStrings(g.Player),
		Dealer:      dealerCards,
		PlayerValue: game.HandValue(g.Player),
		DealerValue: game.HandValue(g.Dealer),
		Balance:     g.Balance,
		Bet:         g.Bet,
		Result:      msg,
		GameOver:    g.Over,
	}

	WriteJSON(w, resp)
}

func (h *Handler) State(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	g := GetGame(r)
	g.Lock()
	defer g.Unlock()

	WriteGame(w, g, "")
}

func (h *Handler) Bet(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	g := GetGame(r)
	g.Lock()
	defer g.Unlock()
	var req BetRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", 400)
		return
	}

	if req.Bet <= 0 || req.Bet > g.Balance {
		http.Error(w, "invalid bet", 400)
		return
	}

	g.Bet = req.Bet
	g.Balance -= req.Bet

	g.Deck = game.NewDeck()

	g.Player = []game.Card{
		game.Draw(&g.Deck),
		game.Draw(&g.Deck),
	}

	g.Dealer = []game.Card{
		game.Draw(&g.Deck),
		game.Draw(&g.Deck),
	}

	g.Over = false
	g.Reveal = false
	if game.HandValue(g.Player) == 21 {

		g.Reveal = true
		g.Over = true
		g.Balance += g.Bet * 2
		msg := "Blackjack! You win"
		WriteGame(w, g, msg)
		h.saveGame(r, g, msg)
		return
	}
	WriteGame(w, g, "")
}

func (h *Handler) Hit(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	g := GetGame(r)

	g.Lock()
	defer g.Unlock()

	if game.HandValue(g.Player) == 21 {

		g.Reveal = true
		g.Over = true

		WriteGame(w, g, "21! Остановись")
		return
	}
	// защита от нажатия до начала игры
	if g.Bet == 0 || len(g.Deck) == 0 {
		WriteGame(w, g, "Place a bet first")
		return
	}

	if g.Over {
		WriteGame(w, g, "")
		return
	}

	g.Player = append(g.Player, game.Draw(&g.Deck))
	value := game.HandValue(g.Player)

	if value == 21 {

		g.Reveal = true
		g.Over = true
		g.Balance += g.Bet * 2

		msg := "21! You win"

		h.saveGame(r, g, msg)

		WriteGame(w, g, msg)
		return
	}
	if value > 21 {

		g.Over = true
		g.Reveal = true

		msg := "Bust! You lose"

		h.saveGame(r, g, msg)

		WriteGame(w, g, msg)
		return
	}

	WriteGame(w, g, "")
}

func (h *Handler) Stand(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	g := GetGame(r)

	g.Lock()
	defer g.Unlock()

	if g.Bet == 0 || len(g.Deck) == 0 {
		WriteGame(w, g, "Place a bet first")
		return
	}

	if g.Over {
		WriteGame(w, g, "")
		return
	}

	g.Reveal = true

	// дилер тянет карты
	for game.HandValue(g.Dealer) < 17 && len(g.Deck) > 0 {
		g.Dealer = append(g.Dealer, game.Draw(&g.Deck))
	}

	p := game.HandValue(g.Player)
	d := game.HandValue(g.Dealer)

	g.Over = true

	var msg string

	switch {

	case d > 21 || p > d:
		g.Balance += g.Bet * 2
		msg = "You win!"

	case p == d:
		g.Balance += g.Bet
		msg = "Push"

	default:
		msg = "Dealer wins"

	}

	h.saveGame(r, g, msg)

	WriteGame(w, g, msg)
}
func (h *Handler) Leaderboard(w http.ResponseWriter, r *http.Request) {

	players, err := h.playerRepo.Leaderboard()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	WriteJSON(w, players)
}
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {

	cookie, _ := r.Cookie("session_id")

	playerID, _ := h.playerRepo.GetOrCreate(cookie.Value)

	history, err := h.historyRepo.GetHistory(playerID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	WriteJSON(w, history)
}
func (h *Handler) saveGame(r *http.Request, g *game.Game, msg string) {

	cookie, err := r.Cookie("session_id")
	if err != nil {
		return
	}

	playerID, err := h.playerRepo.GetOrCreate(cookie.Value)
	if err != nil {
		return
	}

	h.historyRepo.SaveGame(playerID, g, msg)

	// обновляем баланс
	h.playerRepo.UpdateBalance(playerID, g.Balance)

	// если выигрыш
	if msg == "You win!" || msg == "21! You win" || msg == "Blackjack! You win" {
		h.playerRepo.AddWin(playerID, g.Bet)
	}

}
func (h *Handler) SetName(w http.ResponseWriter, r *http.Request) {

	cookie, _ := r.Cookie("session_id")

	playerID, _ := h.playerRepo.GetOrCreate(cookie.Value)

	var req struct {
		Name string `json:"name"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	h.playerRepo.SetName(playerID, req.Name)

	WriteJSON(w, map[string]string{"status": "ok"})
}
