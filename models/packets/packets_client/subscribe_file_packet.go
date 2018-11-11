package packets_client

import "fmt"

type SubscribeFilePacket struct {
	Subscribe    bool `json:"subscribe"`
	WithPrevious bool `json:"with-previous"`
}

func (packet *SubscribeFilePacket) AckPrint() {
	fmt.Printf(packet.String())
}

func (packet SubscribeFilePacket) String() string {
	text := "SUBSCRIBE FILE"
	if packet.Subscribe {
		with := "without"
		if packet.WithPrevious {
			with = "with"
		}
		return fmt.Sprintf("%s %s previously indexed files\n", text, with)
	} else {
		return fmt.Sprintf("UN%s\n", text)
	}
}
