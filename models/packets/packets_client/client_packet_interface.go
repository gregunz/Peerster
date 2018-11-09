package packets_client

type ClientPacketI interface {
	ToClientPacket() *ClientPacket
}
