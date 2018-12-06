package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"math/rand"
	"strings"
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

func Match(filename string, keywords []string) bool {
	for _, k := range keywords {
		if strings.Contains(filename, k) { // here is where we check for keyword match
			return true
		}
	}
	return false
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

func FirstNZero(n int, bytes []byte) bool {
	if n < 0 || n > len(bytes) {
		common.HandleError(fmt.Errorf("FirstNZero failed with n=%d and len(bytes)=%d", n, len(bytes)))
		return false
	}
	return AllZero(bytes[:n])
}

func AllZero(bytes []byte) bool {
	for _, v := range bytes {
		if v != 0 {
			return false
		}
	}
	return true
}

func Random32Bytes() [32]byte {
	bytes := [32]byte{}
	rand.Read(bytes[:])
	return bytes
}
