package repository

import (
	"database/sql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}

type User struct {
	ID           int
	Username     string
	PasswordHash string
}

func (r *UserRepository) Create(username, passwordHash string) (int, error) {
	var id int
	err := r.db.QueryRow(
		`INSERT INTO users(username, password_hash) VALUES($1, $2) RETURNING id`,
		username,
		passwordHash,
	).Scan(&id)
	return id, err
}

func (r *UserRepository) GetByUsername(username string) (*User, error) {
	var u User
	err := r.db.QueryRow(
		`SELECT id, username, password_hash FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepository) GetByID(id int) (*User, error) {
	var u User
	err := r.db.QueryRow(
		`SELECT id, username, password_hash FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}
