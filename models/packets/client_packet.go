package packets

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"strings"
)

type ClientPacket struct {
	Text    *TextMessage   `json:"text"`
	AddNode *AddNodePacket `json:"add-node"`
}

func (packet *ClientPacket) AckPrint() {
	if packet.IsText() {
		packet.Text.AckPrint()
	}
}

func (packet *ClientPacket) IsText() bool {
	return packet.Text != nil
}

func (packet *ClientPacket) IsAddNode() bool {
	return packet.AddNode != nil
}

func (packet *ClientPacket) Check() error {
	var counter uint = 0
	if packet.IsText() {
		counter += 1
	}
	if packet.IsAddNode() {
		counter += 1
	}
	if counter == 1 {
		return nil
	}
	return fmt.Errorf("ClientPacket should have at least and at most one entry not nil instead of %s", packet.String())
}

func (packet ClientPacket) String() string {
	ls := []string{}
	if packet.IsText() {
		ls = append(ls, packet.Text.String())
	}
	if packet.IsAddNode() {
		ls = append(ls, packet.AddNode.String())
	}
	if len(ls) == 0 {
		common.HandleError(fmt.Errorf("empty gossip packet"))
		return ""
	}
	return strings.Join(ls, " + ")
}
