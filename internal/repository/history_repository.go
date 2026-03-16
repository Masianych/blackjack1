package repository

import (
	"database/sql"
	"strings"

	"blackjack/internal/game"
)

type HistoryRepository struct {
	db *sql.DB
}

func NewHistoryRepository(db *sql.DB) *HistoryRepository {
	return &HistoryRepository{db}
}

func (r *HistoryRepository) SaveGame(playerID int, g *game.Game, result string) error {

	playerCards := ""
	dealerCards := ""

	for i, c := range g.Player {
		if i > 0 {
			playerCards += ","
		}
		playerCards += c.Rank + c.Suit
	}

	for i, c := range g.Dealer {
		if i > 0 {
			dealerCards += ","
		}
		dealerCards += c.Rank + c.Suit
	}

	_, err := r.db.Exec(`
        INSERT INTO game_history
        (player_id, player_cards, dealer_cards, bet, result)
        VALUES ($1,$2,$3,$4,$5)
    `,
		playerID,
		playerCards,
		dealerCards,
		g.Bet,
		result,
	)

	return err
}

func cardsToString(cards []game.Card) string {

	var out []string

	for _, c := range cards {
		out = append(out, c.Rank+c.Suit)
	}

	return strings.Join(out, ",")
}

type GameHistory struct {
	PlayerCards string `json:"player_cards"`
	DealerCards string `json:"dealer_cards"`
	Bet         int    `json:"bet"`
	Result      string `json:"result"`
	CreatedAt   string `json:"created_at"`
}

func (r *HistoryRepository) GetHistory(playerID int) ([]GameHistory, error) {

	rows, err := r.db.Query(`
        SELECT player_cards, dealer_cards, bet, result
        FROM game_history
        WHERE player_id=$1
        ORDER BY created_at DESC
        LIMIT 20
    `, playerID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	history := []GameHistory{}

	for rows.Next() {

		var g GameHistory

		err := rows.Scan(
			&g.PlayerCards,
			&g.DealerCards,
			&g.Bet,
			&g.Result,
		)

		if err != nil {
			return nil, err
		}

		history = append(history, g)
	}

	return history, nil
}
