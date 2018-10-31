package packets_client

import "fmt"

type SubscribeOriginPacket struct {
	Subscribe    bool `json:"subscribe"`
	WithPrevious bool `json:"with-previous"`
}

func (packet *SubscribeOriginPacket) AckPrint() {
	fmt.Printf(packet.String())
}

func (packet SubscribeOriginPacket) String() string {
	text := "SUBSCRIBE ORIGIN"
	if packet.Subscribe {
		with := "without"
		if packet.WithPrevious {
			with = "with"
		}
		return fmt.Sprintf("%s %s previous origins\n", text, with)
	} else {
		return fmt.Sprintf("UN%s\n", text)
	}
}
