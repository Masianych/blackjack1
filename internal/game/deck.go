package game

import "math/rand"

var ranks = []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
var suits = []string{"♠", "♥", "♦", "♣"}

func NewDeck() []Card {

	deck := make([]Card, 0, 52)

	for _, s := range suits {
		for _, r := range ranks {
			deck = append(deck, Card{r, s})
		}
	}

	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	return deck
}

func Draw(deck *[]Card) Card {

	c := (*deck)[0]
	*deck = (*deck)[1:]

	return c
}
