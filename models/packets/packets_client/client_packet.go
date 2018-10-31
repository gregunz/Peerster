package packets_client

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"strings"
)

type ClientPacket struct {
	GetId            *GetIdPacket            `json:"get-id"`
	PostMessage      *PostMessagePacket      `json:"post-message"`
	PostNode         *PostNodePacket         `json:"post-node"`
	SubscribeMessage *SubscribeMessagePacket `json:"subscribe-message"`
	SubscribeNode    *SubscribeNodePacket    `json:"subscribe-node"`
	SubscribeOrigin  *SubscribeOriginPacket  `json:"subscribe-origin"`
}

func (packet *ClientPacket) IsGetId() bool {
	return packet.GetId != nil
}

func (packet *ClientPacket) IsPostMessage() bool {
	return packet.PostMessage != nil
}

func (packet *ClientPacket) IsPostNode() bool {
	return packet.PostNode != nil
}

func (packet *ClientPacket) IsSubscribeMessage() bool {
	return packet.SubscribeMessage != nil
}

func (packet *ClientPacket) IsSubscribeNode() bool {
	return packet.SubscribeNode != nil
}

func (packet *ClientPacket) IsSubscribeOrigin() bool {
	return packet.SubscribeOrigin != nil
}

func (packet *ClientPacket) AckPrint() {

	if packet.IsGetId() {
		packet.GetId.AckPrint()
	}
	if packet.IsPostMessage() {
		packet.PostMessage.AckPrint()
	}
	if packet.IsPostNode() {
		packet.PostNode.AckPrint()
	}
	if packet.IsSubscribeMessage() {
		packet.SubscribeMessage.AckPrint()
	}
	if packet.IsSubscribeNode() {
		packet.SubscribeNode.AckPrint()
	}
	if packet.IsSubscribeOrigin() {
		packet.SubscribeOrigin.AckPrint()
	}
}

func (packet *ClientPacket) Check() error {
	var counter uint = 0
	if packet.IsGetId() {
		counter += 1
	}
	if packet.IsPostMessage() {
		counter += 1
	}
	if packet.IsPostNode() {
		counter += 1
	}
	if packet.IsSubscribeMessage() {
		counter += 1
	}
	if packet.IsSubscribeNode() {
		counter += 1
	}
	if packet.IsSubscribeOrigin() {
		counter += 1
	}
	if counter == 1 {
		return nil
	}
	return fmt.Errorf("ClientPacket should have at least and at most one entry not nil instead of %s", packet.String())
}

func (packet ClientPacket) String() string {
	ls := []string{}
	if packet.IsGetId() {
		ls = append(ls, packet.GetId.String())
	}
	if packet.IsPostMessage() {
		ls = append(ls, packet.PostMessage.String())
	}
	if packet.IsPostNode() {
		ls = append(ls, packet.PostNode.String())
	}
	if packet.IsSubscribeMessage() {
		ls = append(ls, packet.SubscribeMessage.String())
	}
	if packet.IsSubscribeNode() {
		ls = append(ls, packet.SubscribeNode.String())
	}
	if packet.IsSubscribeOrigin() {
		ls = append(ls, packet.SubscribeOrigin.String())
	}
	if len(ls) == 0 {
		common.HandleError(fmt.Errorf("empty client packet"))
		return "<empty>"
	}
	return strings.Join(ls, " + ")
}
