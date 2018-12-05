package main

import (
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/utils"
)

func main() {
	d := utils.Distributor(32, 3)
	for i := 0; i < 10; i++ {
		logger.Printlnf("%d", d())
	}
}
