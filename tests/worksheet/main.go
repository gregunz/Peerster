package main

import (
	"github.com/gregunz/Peerster/blockchain"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/vector_clock"
	"github.com/gregunz/Peerster/utils"
)

func main() {
	bcf := blockchain.NewBCF()
	for !bcf.MineOnce() {
	}
	logger.Printlnf("%s", bcf.GetHead())
}

func testingDistributor() {
	d := utils.Distributor(32, 3)
	for i := 0; i < 10; i++ {
		logger.Printlnf("%d", d())
	}
}

func testingRandomGen() {
	logger.Printlnf("%x", utils.Random32Bytes())
	logger.Printlnf("%x", utils.Random32Bytes())
	logger.Printlnf("%x", utils.Random32Bytes())
}

func testingChannels() {
	c := vector_clock.NewRumorChan(true)
	c.Push(&packets_gossiper.RumorMessage{
		Origin: "hihihihi",
	})
	logger.Printlnf("%s", *c.Get())
}
