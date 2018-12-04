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

func Min_uint64(a, b uint64) uint64 {
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

func Distributor(budget int, num int) func() int {
	distributed := 0
	if budget >= num {
		budgetPerPerson := budget / num
		numWithPlusOne := budget % num
		return func() int {
			defer func() { distributed += 1 }()
			if numWithPlusOne-distributed > 0 {
				return budgetPerPerson + 1
			} else if distributed < num {
				return budgetPerPerson
			} else {
				return 0
			}
		}
	} else {
		return func() int {
			defer func() { distributed += 1 }()
			if distributed < budget {
				return 1
			} else {
				return 0
			}
		}
	}
}
