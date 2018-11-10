package packets_gossiper

type Transmittable interface {
	Hopped() Transmittable
	IsTransmittable() bool
	Dest() string

	GossipPacketI // transmittable should also be sendable (GossipPacketI)
}
