package packets_client

import "fmt"

type SubscribeNodePacket struct {
	Subscribe    bool `json:"subscribe"`
	WithPrevious bool `json:"with-previous"`
}

func (packet *SubscribeNodePacket) AckPrint() {
	fmt.Println(packet.String())
}

func (packet SubscribeNodePacket) String() string {
	text := "SUBSCRIBE NODE"
	if packet.Subscribe {
		with := "without"
		if packet.WithPrevious {
			with = "with"
		}
		return fmt.Sprintf("%s %s previous nodes", text, with)
	} else {
		return fmt.Sprintf("UN%s", text)
	}
}
