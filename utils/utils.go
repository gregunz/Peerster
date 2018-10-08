package utils

import "math/rand"

func FlipCoin() bool {
	return (rand.Int() % 2) == 0
}
