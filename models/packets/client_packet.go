package packets

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"strings"
)

type ClientPacket struct {
	GetId       *GetIdPacket       `json:"get-id"`
	GetMessage  *GetMessagePacket  `json:"get-message"`
	PostMessage *PostMessagePacket `json:"post-message"`
	GetNode     *GetNodePacket     `json:"get-node"`
	PostNode    *PostNodePacket    `json:"post-node"`
}

func (packet *ClientPacket) IsGetId() bool {
	return packet.GetId != nil
}

func (packet *ClientPacket) IsGetMessage() bool {
	return packet.GetMessage != nil
}

func (packet *ClientPacket) IsPostMessage() bool {
	return packet.PostMessage != nil
}

func (packet *ClientPacket) IsGetNode() bool {
	return packet.GetNode != nil
}

func (packet *ClientPacket) IsPostNode() bool {
	return packet.PostNode != nil
}

func (packet *ClientPacket) AckPrint() {
	if packet.IsGetId() {
		packet.GetId.AckPrint()
	}
	if packet.IsGetMessage() {
		packet.GetMessage.AckPrint()
	}
	if packet.IsPostMessage() {
		packet.PostMessage.AckPrint()
	}
	if packet.IsGetNode() {
		packet.GetNode.AckPrint()
	}
	if packet.IsPostNode() {
		packet.PostNode.AckPrint()
	}
}

func (packet *ClientPacket) Check() error {
	var counter uint = 0
	if packet.IsGetId() {
		counter += 1
	}
	if packet.IsGetMessage() {
		counter += 1
	}
	if packet.IsPostMessage() {
		counter += 1
	}
	if packet.IsGetNode() {
		counter += 1
	}
	if packet.IsPostNode() {
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
	if packet.IsGetMessage() {
		ls = append(ls, packet.GetMessage.String())
	}
	if packet.IsPostMessage() {
		ls = append(ls, packet.PostMessage.String())
	}
	if packet.IsGetNode() {
		ls = append(ls, packet.GetNode.String())
	}
	if packet.IsPostNode() {
		ls = append(ls, packet.PostNode.String())
	}
	if len(ls) == 0 {
		common.HandleError(fmt.Errorf("empty gossip packet"))
		return ""
	}
	return strings.Join(ls, " + ")
}
