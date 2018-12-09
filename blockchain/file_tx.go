package blockchain

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
)

type Tx struct {
	id   string
	File packets_gossiper.File
}

func NewTx(publish *packets_gossiper.TxPublish) *Tx {
	//hash := publish.File.Hash()
	return &Tx{
		id:   publish.File.Name,
		File: publish.File,
	}
}

func (tx *Tx) IsValid(tx2 *Tx) bool {
	return tx.File.Name != tx2.File.Name //&&
	//utils.HashToHex(tx.File.MetafileHash) != utils.HashToHex(tx2.File.MetafileHash)
}

func (tx *Tx) ToTxPublish(hopLimit uint32) packets_gossiper.TxPublish {
	return packets_gossiper.TxPublish{
		File:     tx.File,
		HopLimit: hopLimit,
	}
}
