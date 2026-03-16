package repository

import "database/sql"

type PlayerRepository struct {
	db *sql.DB
}

func NewPlayerRepository(db *sql.DB) *PlayerRepository {
	return &PlayerRepository{db}
}

func (r *PlayerRepository) GetOrCreate(sessionID string) (int, error) {

	var id int

	err := r.db.QueryRow(
		`SELECT id FROM players WHERE session_id=$1`,
		sessionID,
	).Scan(&id)

	if err == sql.ErrNoRows {

		err = r.db.QueryRow(
			`INSERT INTO players(session_id) VALUES($1) RETURNING id`,
			sessionID,
		).Scan(&id)
	}

	return id, err
}

type Player struct {
	ID        int
	SessionID string
	Balance   int
}

func (r *PlayerRepository) TopPlayers() ([]Player, error) {

	rows, err := r.db.Query(`
		SELECT id, session_id, balance
		FROM players
		ORDER BY balance DESC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var players []Player

	for rows.Next() {

		var p Player

		err := rows.Scan(&p.ID, &p.SessionID, &p.Balance)
		if err != nil {
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}
func (r *PlayerRepository) AddWin(playerID int, amount int) error {

	_, err := r.db.Exec(`
		UPDATE players
		SET balance = balance + $1,
		    total_won = total_won + $1
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

		err := rows.Scan(
			&p.Name,
			&p.Balance,
			&p.TotalWon,
		)
		if err != nil {
			return nil, err
		}

		players = append(players, p)
	}

	return players, nil
}
func (r *PlayerRepository) UpdateBalance(playerID int, balance int) error {

	_, err := r.db.Exec(`
		UPDATE players
		SET balance=$1
		WHERE id=$2
	`, balance, playerID)

	return err
}
func (r *PlayerRepository) SetName(playerID int, name string) error {

	_, err := r.db.Exec(`
		UPDATE players
		SET name=$1
		WHERE id=$2
	`, name, playerID)

	return err
}
