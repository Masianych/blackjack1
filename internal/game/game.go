package game

import (
	"strconv"
	"sync"
)

type Card struct {
	Rank string
	Suit string
}

type Game struct {
	mu sync.Mutex

	Deck    []Card
	Player  []Card
	Dealer  []Card
	Balance int
	Bet     int
	Reveal  bool
	Over    bool
}

func (g *Game) Lock() {
	g.mu.Lock()
}

func (g *Game) Unlock() {
	g.mu.Unlock()
}
func HandValue(hand []Card) int {

	total := 0
	aces := 0

	for _, c := range hand {

		switch c.Rank {

		case "A":
			total += 11
			aces++

		case "K", "Q", "J":
			total += 10

		default:
			v, err := strconv.Atoi(c.Rank)
			if err == nil {
				total += v
			}
		}
	}

	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}

	return total
}
