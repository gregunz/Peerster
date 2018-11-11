package packets_client

import "fmt"

type GetIdPacket struct{}

func (packet *GetIdPacket) AckPrint() {
	fmt.Println(packet.String())
}

func (packet GetIdPacket) String() string {
	return fmt.Sprintf("GET ID")
}
