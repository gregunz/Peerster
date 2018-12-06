package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/utils"
	"sync"
)

const (
	NumOfZeroBytes = 2
)

type FileBlockBuilder struct {
	length int

	previous     *FileBlock
	prevHash     [32]byte // needed only if previous is nil
	nonce        [32]byte
	transactions map[string]*Tx

	sync.RWMutex
}

func NewFileBlockBuilder(previousBlock *FileBlock) *FileBlockBuilder {
	fbb := &FileBlockBuilder{
		length: 1,

		previous: previousBlock,
		prevHash: [32]byte{},
		//nonce:            [32]byte{},
		transactions: map[string]*Tx{},
	}
	if previousBlock != nil {
		fbb.length = previousBlock.length + 1
	}
	return fbb
}

func (fbb *FileBlockBuilder) SetNonce(nonce [32]byte) {
	fbb.Lock()
	defer fbb.Unlock()

	fbb.nonce = nonce
}

func (fbb *FileBlockBuilder) AddTxIfValid(newTx *Tx) bool {
	fbb.Lock()
	defer fbb.Unlock()

	return fbb.addTxIfValid(newTx)
}

func (fbb *FileBlockBuilder) SetBlockAndBuild(block *packets_gossiper.Block) (*FileBlock, error) {
	fbb.Lock()
	defer fbb.Unlock()

	if fbb.previous != nil && fbb.previous.hash != block.PrevHash {
		return nil, fmt.Errorf("trying to add a block over a mismatching previous file-block")
	}

	fbb.transactions = map[string]*Tx{} // clear previous entries in transactions if they were some
	for _, txPublish := range block.Transactions {
		tx := NewTx(txPublish)
		if !fbb.addTxIfValid(tx) { // one tx contradicts another
			return nil, fmt.Errorf("one tx (%s) contradicts another previous tx", tx.File.String())
		}
	}
	fbb.prevHash = block.PrevHash // in case previous is nil when computing hash (prevHash needed)
	fbb.nonce = block.Nonce

	return fbb.Build()
}

func (fbb *FileBlockBuilder) Build() (*FileBlock, error) {
	fbb.RLock()
	defer fbb.RUnlock()

	hash := fbb.Hash()
	if !utils.FirstNZero(NumOfZeroBytes, hash[:]) { // checking if hash is truly starting with `NumOfZeroBytes` bytes
		return nil, fmt.Errorf("hash needs to have %d leading bits set to zeros (%d bytes)", NumOfZeroBytes*8, NumOfZeroBytes)
	}

	return &FileBlock{
		length:       fbb.length,
		id:           utils.HashToHex(hash[:]),
		previous:     fbb.previous,
		hash:         hash,
		nonce:        fbb.nonce,
		transactions: fbb.transactions,
	}, nil
}

func (fbb *FileBlockBuilder) Hash() (out [32]byte) {
	fbb.RLock()
	defer fbb.RUnlock()

	previousHash := fbb.prevHash
	if fbb.previous != nil {
		previousHash = fbb.previous.hash
	}

	h := sha256.New()
	h.Write(previousHash[:])
	h.Write(fbb.nonce[:])
	err := binary.Write(h, binary.LittleEndian, uint32(len(fbb.transactions)))
	if err != nil {
		common.HandleAbort("unexpected error when computing hash of block", err)
		return
	}
	for _, t := range fbb.transactions {
		th := t.File.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}

// private functions without locks

func (fbb *FileBlockBuilder) addTxIfValid(newTx *Tx) bool {
	if _, ok := fbb.transactions[newTx.id]; ok {
		return false
	}
	prevBlock := fbb.previous
	for prevBlock != nil {
		if !prevBlock.txIsValidWithThisBlock(newTx) {
			return false
		}
	}
	fbb.transactions[newTx.id] = newTx
	return true
}
