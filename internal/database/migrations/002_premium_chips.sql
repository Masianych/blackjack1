-- Добавить premium_chips для существующих БД
ALTER TABLE players ADD COLUMN IF NOT EXISTS premium_chips BOOLEAN DEFAULT FALSE;
