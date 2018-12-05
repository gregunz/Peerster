package packets_gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/utils"
	"strings"
)

type SearchRequest struct {
	Origin   string   `json:"origin"`
	Budget   uint64   `json:"budget"`
	Keywords []string `json:"keywords"`
}

func (packet *SearchRequest) AckPrint() {
	logger.Printlnf(packet.String())
}

func (packet *SearchRequest) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		SearchRequest: packet,
	}
}

func (packet *SearchRequest) String() string {
	return fmt.Sprintf("SEARCH REQUEST origin %s budget %d with keywords %s",
		packet.Origin, packet.Budget, strings.Join(packet.Keywords, " "))
}

func (packet *SearchRequest) GetBudget() uint64 {
	return packet.Budget
}

func (packet *SearchRequest) SetBudget(budget uint64) {
	packet.Budget = budget
}

func (packet *SearchRequest) DividePacket(num int) []BudgetPacket {
	packets := []BudgetPacket{}
	budgetDistributor := utils.Distributor(int(packet.Budget), num)
	for i := 0; i < num; i++ {
		newPacket := &SearchRequest{
			Origin:   packet.Origin,
			Budget:   uint64(budgetDistributor()),
			Keywords: packet.Keywords,
		}
		packets = append(packets, newPacket)
	}
	return packets
}
