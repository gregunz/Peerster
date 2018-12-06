package packets_gossiper

import (
	"fmt"
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
