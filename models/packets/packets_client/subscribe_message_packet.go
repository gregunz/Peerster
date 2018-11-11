package packets_client

import "fmt"

type SubscribeMessagePacket struct {
	Subscribe    bool `json:"subscribe"`
	WithPrevious bool `json:"with-previous"`
}

func (packet *SubscribeMessagePacket) AckPrint() {
	fmt.Println(packet.String())
}

func (packet SubscribeMessagePacket) String() string {
	text := "SUBSCRIBE MESSAGE"
	if packet.Subscribe {
		with := "without"
		if packet.WithPrevious {
			with = "with"
		}
		return fmt.Sprintf("%s %s previous messages", text, with)
	} else {
		return fmt.Sprintf("UN%s", text)
	}
}
