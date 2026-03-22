-- Схема: логин по нику (username)
-- Если на Render старая схема (session_id) — пересоздаём таблицы
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='players' AND column_name='session_id') THEN
    DROP TABLE IF EXISTS game_history;
    DROP TABLE IF EXISTS players;
    DROP TABLE IF EXISTS users;
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) DEFAULT 'Player',
    balance INTEGER DEFAULT 1000,
    total_won INTEGER DEFAULT 0,
    premium_chips BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS game_history (
    id SERIAL PRIMARY KEY,
    player_id INTEGER NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    player_cards TEXT,
    dealer_cards TEXT,
    bet INTEGER NOT NULL,
    result VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_game_history_player_id ON game_history(player_id);
-- Индекс только если колонка user_id есть (новая схема)
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='players' AND column_name='user_id') THEN
    CREATE INDEX IF NOT EXISTS idx_players_user_id ON players(user_id);
  END IF;
END $$;
