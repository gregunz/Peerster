package blockchain

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/utils"
	"sync"
)

type BCF struct {
	forks     map[string]*FileBlock
	allBlocks map[string]*FileBlock

	headLength int
	head       *FileBlockBuilder // the block we will be mining over (not yet on the blockchain, hence *Builder)

	sync.RWMutex
}

func NewBCF(block *packets_gossiper.Block) *BCF {
	bcf := &BCF{
		forks:     map[string]*FileBlock{},
		allBlocks: map[string]*FileBlock{},
	}
	bcf.AddBlock(block)
	return bcf
}

func (bcf *BCF) AddTx(tx *Tx) {
	bcf.RLock()
	defer bcf.RUnlock()

	bcf.head.AddTxIfValid(tx)
}

func (bcf *BCF) AddBlock(block *packets_gossiper.Block) bool {
	bcf.Lock()
	defer bcf.Unlock()

	previousId := utils.HashToHex(block.PrevHash[:])
	var previousBlock *FileBlock
	if bcf.headLength == 0 {
		// first block, welcome and be our master! (previous set to nil)
		previousBlock = nil
	} else if forkBlock, ok := bcf.forks[previousId]; ok {
		// no new fork but one longer head (previous is a fork)
		previousBlock = forkBlock
	} else if singleBlock, ok := bcf.allBlocks[previousId]; ok {
		// new fork, cannot be longest head (previous is part of the chain)
		previousBlock = singleBlock
	} else if utils.AllZero(block.PrevHash[:]) {
		// new fork but without tail ??? should I care ?
		//Todo: do better but for now we don't take it
		return false
	} else {
		// same as above ?
		return false
	}

	newFBB := NewFileBlockBuilder(previousBlock)
	if fb, err := newFBB.SetBlockAndBuild(block); err != nil {
		common.HandleAbort("AddBlock failed when building", err)
		return false
	} else {
		return bcf.addFileBlock(fb)
	}
}

func (bcf *BCF) MineOnce(block *packets_gossiper.Block) (*FileBlock, error) {
	bcf.Lock()
	defer bcf.Unlock()

	nonce := utils.Random32Bytes()
	bcf.head.SetNonce(nonce)
	return bcf.head.Build()
}

// private functions without locks

func (bcf *BCF) addFileBlock(fb *FileBlock) bool {
	if fb.previous == nil {
		bcf.allBlocks[fb.id] = fb
		bcf.forks[fb.id] = fb
		bcf.headLength = fb.length
		bcf.head = NewFileBlockBuilder(fb)
		return true
	} else if _, ok := bcf.forks[fb.previous.id]; ok {
		bcf.allBlocks[fb.id] = fb
		delete(bcf.forks, fb.previous.id)
		bcf.forks[fb.id] = fb

		if fb.length > bcf.headLength { // even the longest fork now! changing head!
			// we need to keep the transactions that are not invalidated nor included in the new block
			newHead := NewFileBlockBuilder(fb)
			for _, tx := range bcf.head.newTransactions {
				newHead.AddTxIfValid(tx)
			}
			bcf.headLength = fb.length //not new head which is 1 greater
			bcf.head = newHead
		}
		return true
	} else if _, ok := bcf.allBlocks[fb.previous.id]; ok {
		bcf.allBlocks[fb.id] = fb
		bcf.forks[fb.id] = fb
		return true
	}
	common.HandleError(fmt.Errorf("file-block comes out of nowhere"))
	return false
}
