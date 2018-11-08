package utils

import "math/rand"

func FlipCoin() bool {
	return (rand.Int() % 2) == 0
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
