package repository

import "database/sql"

type PlayerRepository struct {
	db *sql.DB
}

func NewPlayerRepository(db *sql.DB) *PlayerRepository {
	return &PlayerRepository{db}
}

func (r *PlayerRepository) GetOrCreateForUser(userID int, name string) (*Player, error) {
	var p Player
	err := r.db.QueryRow(
		`SELECT id, user_id, name, balance, total_won, premium_chips FROM players WHERE user_id = $1`,
		userID,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.Balance, &p.TotalWon, &p.PremiumChips)

	if err == sql.ErrNoRows {
		err = r.db.QueryRow(
			`INSERT INTO players(user_id, name) VALUES($1, $2) RETURNING id, user_id, name, balance, total_won, premium_chips`,
			userID,
			name,
		).Scan(&p.ID, &p.UserID, &p.Name, &p.Balance, &p.TotalWon, &p.PremiumChips)
	}

	return &p, err
}

func (r *PlayerRepository) GetByUserID(userID int) (*Player, error) {
	var p Player
	err := r.db.QueryRow(
		`SELECT id, user_id, name, balance, total_won, COALESCE(premium_chips, false) FROM players WHERE user_id = $1`,
		userID,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.Balance, &p.TotalWon, &p.PremiumChips)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

type Player struct {
	ID           int
	UserID       int
	Name         string
	Balance      int
	TotalWon     int
	PremiumChips bool
}

func (r *PlayerRepository) AddWin(playerID int, amount int) error {
	_, err := r.db.Exec(`
		UPDATE players
		SET total_won = total_won + $1
		WHERE id = $2
	`, amount, playerID)
	return err
}

type LeaderboardPlayer struct {
	Name     string `json:"name"`
	Balance  int    `json:"balance"`
	TotalWon int    `json:"total_won"`
}

func (r *PlayerRepository) Leaderboard() ([]LeaderboardPlayer, error) {
	rows, err := r.db.Query(`
		SELECT 
			COALESCE(name,'Player'),
			balance,
			total_won
		FROM players
		ORDER BY total_won DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []LeaderboardPlayer
	for rows.Next() {
		var p LeaderboardPlayer
		if err := rows.Scan(&p.Name, &p.Balance, &p.TotalWon); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, nil
}

func (r *PlayerRepository) UpdateBalance(playerID int, balance int) error {
	_, err := r.db.Exec(`
		UPDATE players SET balance = $1 WHERE id = $2
	`, balance, playerID)
	return err
}

func (r *PlayerRepository) SetName(playerID int, name string) error {
	_, err := r.db.Exec(`
		UPDATE players SET name = $1 WHERE id = $2
	`, name, playerID)
	return err
}

const PremiumChipsPrice = 1200

func (r *PlayerRepository) BuyPremiumChips(playerID int, balance int) error {
	if balance < PremiumChipsPrice {
		return nil // caller should check
	}
	_, err := r.db.Exec(`
		UPDATE players SET balance = balance - $1, premium_chips = true WHERE id = $2 AND balance >= $1
	`, PremiumChipsPrice, playerID)
	return err
}
