package blockchain

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/utils"
	"sync"
	"time"
)

type BCF struct {
	forks       map[string]*FileBlock
	allBlocks   map[string]*FileBlock
	chainLength int
	head        *FileBlockBuilder // the block we will be mining over (not yet on the blockchain, hence *Builder)

	MineChan MineChan

	sync.RWMutex
}

func NewBCF() *BCF {
	return &BCF{
		forks:       map[string]*FileBlock{},
		allBlocks:   map[string]*FileBlock{},
		chainLength: 0,
		head:        NewFileBlockBuilder(nil),
		MineChan:    NewMineChan(true),
	}
}

func (bcf *BCF) AddTx(tx *Tx) {
	bcf.RLock()
	defer bcf.RUnlock()

	bcf.head.AddTxIfValid(tx)
}

func (bcf *BCF) GetHead() *FileBlockBuilder {
	bcf.RLock()
	defer bcf.RUnlock()

	return bcf.head
}

func (bcf *BCF) AddBlock(block *packets_gossiper.Block) bool {
	bcf.Lock()
	defer bcf.Unlock()

	previousId := utils.HashToHex(block.PrevHash[:])
	var previousBlock *FileBlock
	if bcf.chainLength == 0 {
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

func (bcf *BCF) MineOnce() bool {
	bcf.Lock()
	defer bcf.Unlock()

	nonce := utils.Random32Bytes()
	bcf.head.SetNonce(nonce)
	fb, err := bcf.head.Build()
	if err != nil {
		//common.HandleError(err)
		return false
	}
	logger.Printlnf("FOUND-BLOCK %s", utils.HashToHex(fb.hash[:])) //hw03 print
	if bcf.addFileBlock(fb) {
		bcf.MineChan.Push(fb)
		return true
	}
	common.HandleError(fmt.Errorf("block mined not added to chain, this should not happen"))
	return false
}

// public functions without locks

func (bcf *BCF) MiningRoutine(group *sync.WaitGroup) {
	defer group.Done()
	for {
		if len(bcf.head.newTransactions) > 0 { // only mine if new transactions
			bcf.MineOnce()
		}
		// give time to other functions to access locks between mining
		// and allows cpu not to be overused when no transactions
		time.Sleep(10 * time.Millisecond)
	}
}

// private functions without locks

func (bcf *BCF) addFileBlock(fb *FileBlock) bool {
	if fb.previous == nil {
		bcf.allBlocks[fb.id] = fb
		bcf.forks[fb.id] = fb
		logger.Printlnf(fb.ChainString()) // hw03 print
		bcf.chainLength = fb.length
		bcf.head = NewFileBlockBuilder(fb)
		return true
	} else if _, ok := bcf.forks[fb.previous.id]; ok {
		bcf.allBlocks[fb.id] = fb
		delete(bcf.forks, fb.previous.id)
		bcf.forks[fb.id] = fb

		if fb.length > bcf.chainLength {
			// even the longest fork now! changing head!
			// we need to keep the transactions that are not invalidated nor included in the new block

			newHead := NewFileBlockBuilder(fb)
			for _, tx := range bcf.head.newTransactions {
				newHead.AddTxIfValid(tx)
			}

			if rewind, _ := findMergure(fb, bcf.head.previous); rewind > 0 {
				logger.Printlnf("FORK-LONGER rewind %d blocks", rewind)
			}
			logger.Printlnf(fb.ChainString()) // hw03 print
			bcf.chainLength = fb.length       //not new head which is 1 greater
			bcf.head = newHead
		} else {
			_, hashString := findMergure(fb, bcf.head.previous)
			logger.Printlnf("FORK-SHORTER %s", hashString)
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

func findMergure(newBlock, oldBlock *FileBlock) (int, string) {
	rewind := 0
	newChainBlocks := map[string]bool{}
	newChainBlock := newBlock
	for newChainBlock != nil {
		newChainBlocks[newChainBlock.id] = true
		newChainBlock = newChainBlock.previous
	}

	oldChainBlock := oldBlock
	for oldChainBlock != nil {
		if _, ok := newChainBlocks[oldChainBlock.id]; ok {
			break
		}
		rewind += 1
		oldChainBlock = oldChainBlock.previous
	}
	return rewind, oldBlock.id
}
