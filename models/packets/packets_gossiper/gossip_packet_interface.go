package packets_gossiper

type GossipPacketI interface {
	ToGossipPacket() *GossipPacket
}
