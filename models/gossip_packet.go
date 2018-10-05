package models

import (
	"fmt"
)

type GossipPacket struct {
	Simple *SimpleMessage
	Rumor  *RumorMessage
	Status *StatusPacket
}

func (packet *GossipPacket) Check() error {
	if (packet.Simple != nil && packet.Rumor == nil && packet.Status == nil) ||
		(packet.Simple == nil && packet.Rumor != nil && packet.Status == nil) ||
		(packet.Simple == nil && packet.Rumor == nil && packet.Status != nil) {
		return nil
	}
	return fmt.Errorf("GossipPacket should have at least and at most one entry not nil instead of %s", packet)
}

/*
func (packet *GossipPacket) AckPrint(fromClient bool){
	var msg AckPrintable
	if packet.Simple != nil {
		msg = packet.Simple
	}
	if packet.Rumor != nil {

	}
	msg.AckPrint(fromClient)
}
*/
