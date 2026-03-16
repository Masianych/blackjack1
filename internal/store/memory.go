package store

import (
	"sync"

	"blackjack/internal/game"
)

type MemoryStore struct {
	games map[string]*game.Game
	mu    sync.RWMutex
}

func NewMemoryStore() *MemoryStore {

	return &MemoryStore{
		games: make(map[string]*game.Game),
	}
}

func (s *MemoryStore) Get(id string) (*game.Game, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	g, ok := s.games[id]

	return g, ok
}

func (s *MemoryStore) Set(id string, g *game.Game) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.games[id] = g
}
