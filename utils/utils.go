package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/gregunz/Peerster/common"
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
	return hex.EncodeToString(hash)
}

func HexToHash(hexHash string) []byte {
	hash, err := hex.DecodeString(hexHash)
	if err != nil {
		common.HandleAbort(fmt.Sprint("could not decode hexadecimal string '%s'", hexHash), err)
		return nil
	}
	return hash
}
