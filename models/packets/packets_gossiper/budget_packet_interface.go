package packets_gossiper

type BudgetPacket interface {
	GetBudget() uint64
	SetBudget(uint64)
	DividePacket(int) []BudgetPacket

	GossipPacketI // BudgetPacket should also be sendable (GossipPacketI)
}
