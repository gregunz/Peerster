package blockchain

import (
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/utils"
	"sync"
)

type BCF struct {
	forks     map[string]*FileBlock
	allBlocks map[string]*FileBlock
	//allTx     map[string]*Tx //todo: be sure we need to keep them when head changes

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

	return bcf.addBlock(block)
}

func (bcf *BCF) MineOnce(block *packets_gossiper.Block) (*FileBlock, error) {
	bcf.Lock()
	defer bcf.Unlock()

	nonce := utils.Random32Bytes()
	bcf.head.SetNonce(nonce)
	if fb, err := bcf.head.Build(); err != nil {

	}
}

func (bcf *BCF) addFileBlock(fb *FileBlock) bool {
	if fb.previous == nil {
		bcf.allBlocks[fb.id] = fb
		bcf.forks[fb.id] = fb
		bcf.headLength = fb.length
		bcf.head = NewFileBlockBuilder(fb)
	} else if _, ok := bcf.forks[fb.previous.id]; ok {
		bcf.allBlocks[fb.id] = fb
		delete(bcf.forks, fb.previous.id)
		bcf.forks[fb.id] = fb

		if fb.length > bcf.headLength { // even the longest fork now! changing head!
			// we need to keep the transactions that are not invalidated nor included in the new block
			newHead := NewFileBlockBuilder(fb)
			for _, tx := range bcf.head.transactions {
				newHead.AddTxIfValid(tx)
			}
			bcf.headLength = fb.length //not new head which is 1 greater
			bcf.head = newHead
		}
	} else if _, ok := bcf.allBlocks[fb.previous.id]; ok {
		bcf.allBlocks[fb.id] = fb
		bcf.forks[fb.id] = fb
	}
}

// private functions without locks
func (bcf *BCF) addBlock(block *packets_gossiper.Block) bool {
	previousId := utils.HashToHex(block.PrevHash[:])

	var previousBlock *FileBlock
	var ok bool
	if bcf.headLength == 0 { // first block, welcome and be our master!
		previousBlock = nil
		/*
			newFBB := NewFileBlockBuilder(nil)
			if fb, err := newFBB.SetBlockAndBuild(block); err != nil {
				bcf.allBlocks[fb.id] = fb
				bcf.forks[fb.id] = fb
				bcf.headLength = fb.length
				bcf.head = NewFileBlockBuilder(fb)
			} else {
				common.HandleAbort("AddBlock of new block failed", err)
				return false
			}
		*/
	} else if previousBlock, ok = bcf.forks[previousId]; ok { // no fork but one longer head
		/*
			newFBB := NewFileBlockBuilder(forkBlock)
			if fb, err := newFBB.SetBlockAndBuild(block); err != nil {
				bcf.allBlocks[fb.id] = fb
				delete(bcf.forks, previousId)
				bcf.forks[fb.id] = fb

				if fb.length > bcf.headLength { // even the longest fork now! changing head!
					// we need to keep the transactions that are not invalidated nor included in the new block
					newHead := NewFileBlockBuilder(fb)
					for _, tx := range bcf.head.transactions {
						newHead.AddTxIfValid(tx)
					}
					bcf.headLength = fb.length //not new head which is 1 greater
					bcf.head = newHead
				}
			} else {
				common.HandleAbort("AddBlock of over a fork-block failed", err)
				return false
			}
		*/
	} else if previousBlock, ok = bcf.allBlocks[previousId]; ok { // new fork, cannot be longest head
		/*
			newFBB := NewFileBlockBuilder(prevBlock)
			if fb, err := newFBB.SetBlockAndBuild(block); err != nil {
				bcf.allBlocks[fb.id] = fb
				bcf.forks[fb.id] = fb
			} else {
				common.HandleAbort("AddBlock of a new fork failed", err)
				return false
			}
		*/
	} else if utils.AllZero(block.PrevHash[:]) { // new fork but what is the tail ??? should I care ?
		//TODO
		return false
	} else {
		// same as above ?
		return false
	}

	newFBB := NewFileBlockBuilder(previousBlock)
	fb, err := newFBB.SetBlockAndBuild(block)
	if err != nil {
		common.HandleAbort("AddBlock failed", err)
		return false
	}
	return bcf.addFileBlock(fb)
}
