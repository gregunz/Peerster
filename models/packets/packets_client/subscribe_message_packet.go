package packets_client

import "fmt"

type SubscribeMessagePacket struct {
	WithPrevious bool `json:"with-previous"`
}

func (packet *SubscribeMessagePacket) AckPrint() {
	fmt.Printf(packet.String())
}

func (packet SubscribeMessagePacket) String() string {
	s := "without"
	if packet.WithPrevious {
		s = "with"
	}
	return fmt.Sprintf("SUBSCRIBE MESSAGE %s previous messages\n", s)
}
