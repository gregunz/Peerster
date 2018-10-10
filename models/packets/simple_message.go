package packets

import "fmt"

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

func (msg *SimpleMessage) AckPrint() {
	fmt.Println(msg.String())
}

func (msg *SimpleMessage) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		Simple: msg,
	}
}

func (msg SimpleMessage) String() string {
	return fmt.Sprintf("SIMPLE MESSAGE origin %s from %s contents %s",
		msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
}
