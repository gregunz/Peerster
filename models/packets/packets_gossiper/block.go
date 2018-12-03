package packets_gossiper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/utils"
)

type Block struct {
	PrevHash     [32]byte    `json:"previous-hash"`
	Nonce        [32]byte    `json:"nonce"`
	Transactions []TxPublish `json:"transactions"`
}

func (block *Block) String() string {
	transactions := []string{}
	for _, t := range block.Transactions {
		transactions = append(transactions, fmt.Sprintf("<%s>", t.String()))
	}
	return fmt.Sprintf("BLOCK with previous hash %s nonce %s and transactions <%s>",
		utils.HashToHex(block.PrevHash[:]), utils.HashToHex(block.Nonce[:]), transactions)
}

func (block *Block) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(block.PrevHash[:])
	h.Write(block.Nonce[:])
	err := binary.Write(h, binary.LittleEndian, uint32(len(block.Transactions)))
	if err != nil {
		common.HandleAbort("unexpected error when computing hash of block", err)
		return
	}
	for _, t := range block.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}
