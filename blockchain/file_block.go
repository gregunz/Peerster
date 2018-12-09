package blockchain

import (
	"fmt"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/utils"
	"strings"
)

type FileBlock struct {
	length int
	id     string

	previous     *FileBlock
	hash         [32]byte
	nonce        [32]byte
	transactions map[string]*Tx
}

func (fb *FileBlock) txIsValidWithThisBlock(newTx *Tx) bool {
	_, ok := fb.transactions[newTx.id]
	return !ok //is valid if not yet present
}

func (fb *FileBlock) ToBlock(hopLimit uint32) *packets_gossiper.Block {
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
	return &packets_gossiper.BlockPublish{
		Block:    *fb.ToBlock(hopLimit),
		HopLimit: hopLimit,
	}
}

func (fb *FileBlock) String() string {
	prevHash := [32]byte{}
	if fb.previous != nil {
		prevHash = fb.previous.hash
	}
	txStrings := []string{}
	for _, tx := range fb.transactions {
		txStrings = append(txStrings, tx.File.Name)
	}
	return fmt.Sprintf("%s:%s:%s", fb.hash, utils.HashToHex(prevHash[:]), strings.Join(txStrings, ","))
}

func (fb *FileBlock) ChainString() string {
	blockStrings := []string{}
	block := fb
	for block != nil {
		blockStrings = append(blockStrings, block.String())
		block = block.previous
	}
	return fmt.Sprintf("CHAIN %s", strings.Join(blockStrings, " "))
}
