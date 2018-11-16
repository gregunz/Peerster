package packets_client

import (
	"fmt"
	"github.com/gregunz/Peerster/www/subscription"
	"strings"
)

type ClientPacket struct {
	GetId *GetIdPacket `json:"get-id"`

	PostMessage *PostMessagePacket `json:"post-message"`
	PostNode    *PostNodePacket    `json:"post-node"`

	IndexFile   *IndexFilePacket   `json:"index-file"`
	RequestFile *RequestFilePacket `json:"request-file"`

	SubscribeMessage *SubscribePacket `json:"subscribe-message"`
	SubscribeNode    *SubscribePacket `json:"subscribe-node"`
	SubscribeOrigin  *SubscribePacket `json:"subscribe-origin"`
	SubscribeFile    *SubscribePacket `json:"subscribe-file"`
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

func (packet *ClientPacket) IsIndexFile() bool {
	return packet.IndexFile != nil
}

func (packet *ClientPacket) IsRequestFile() bool {
	return packet.RequestFile != nil
}

func (packet *ClientPacket) IsSubscribeFile() bool {
	return packet.SubscribeFile != nil
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

	if packet.IsRequestFile() {
		packet.RequestFile.AckPrint()
	}
	if packet.IsIndexFile() {
		packet.IndexFile.AckPrint()
	}

	if packet.IsSubscribeMessage() {
		packet.SubscribeMessage.AckPrint(subscription.Message)
	}
	if packet.IsSubscribeNode() {
		packet.SubscribeNode.AckPrint(subscription.Node)
	}
	if packet.IsSubscribeOrigin() {
		packet.SubscribeOrigin.AckPrint(subscription.Origin)
	}
	if packet.IsSubscribeFile() {
		packet.SubscribeFile.AckPrint(subscription.File)
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
	if packet.IsIndexFile() {
		counter += 1
	}
	if packet.IsRequestFile() {
		counter += 1
	}
	if packet.IsSubscribeFile() {
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
	if packet.IsIndexFile() {
		ls = append(ls, packet.IndexFile.String())
	}
	if packet.IsRequestFile() {
		ls = append(ls, packet.RequestFile.String())
	}
	if packet.IsSubscribeMessage() {
		ls = append(ls, packet.SubscribeMessage.String(subscription.Message))
	}
	if packet.IsSubscribeNode() {
		ls = append(ls, packet.SubscribeNode.String(subscription.Node))
	}
	if packet.IsSubscribeOrigin() {
		ls = append(ls, packet.SubscribeOrigin.String(subscription.Origin))
	}
	if packet.IsSubscribeFile() {
		ls = append(ls, packet.SubscribeFile.String(subscription.File))
	}
	if len(ls) == 0 {
		return "<empty>"
	}
	return strings.Join(ls, " + ")
}
