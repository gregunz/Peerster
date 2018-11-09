package utils

import (
	"fmt"
	"math/rand"
)

func FlipCoin() bool {
	return (rand.Int() % 2) == 0
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func HashToHex(hash []byte) string {
	return fmt.Sprintf("%x", hash)
}
