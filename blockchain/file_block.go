package blockchain

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"sync"
)

type FileBlock struct {
	length int
	id     string

	previous     *FileBlock
	hash         [32]byte
	nonce        [32]byte
	transactions map[string]*Tx

	sync.RWMutex
}

func (fb *FileBlock) ToBlock(hopLimit uint32) *packets_gossiper.Block {
	fb.RLock()
	defer fb.RUnlock()

	transactions := []packets_gossiper.TxPublish{}
	for _, tx := range fb.transactions {
		transactions = append(transactions, tx.ToTxPublish(hopLimit))
	}

	prevHash := [32]byte{}
	if fb.previous != nil {
		prevHash = fb.previous.hash
	}
	return &packets_gossiper.Block{
		PrevHash:     prevHash,
		Nonce:        fb.nonce,
		Transactions: transactions,
	}
}

func (fb *FileBlock) ToBlockPublish(hopLimit uint32) *packets_gossiper.BlockPublish {
	fb.RLock()
	defer fb.RUnlock()

	return &packets_gossiper.BlockPublish{
		Block:    *fb.ToBlock(hopLimit),
		HopLimit: hopLimit,
	}
}
