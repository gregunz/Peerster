package packets_client

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/www/subscription"
)

type SubscribePacket struct {
	Subscribe    bool `json:"subscribe"`
	WithPrevious bool `json:"with-previous"`
}

func (packet *SubscribePacket) AckPrint(sub subscription.Sub) {
	logger.Printlnf(packet.String(sub))
}

func (packet *SubscribePacket) String(sub subscription.Sub) string {
	text := "SUBSCRIBE"
	if packet.Subscribe {
		with := "without"
		if packet.WithPrevious {
			with = "with"
		}
		return fmt.Sprintf("%s %s %s previous", text, sub.Name(), with)
	} else {
		return fmt.Sprintf("UN%s %s", text, sub.Name())
	}
}
