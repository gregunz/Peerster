package packets_gossiper

import "fmt"

type PrivateMessage struct {
	Origin      string `json:"origin"`
	Id          uint32 `json:"id"`
	Text        string `json:"text"`
	Destination string `json:"dest"`
	HopLimit    uint32 `json:"hop-limit"`
}

func (msg *PrivateMessage) AckPrint(myOrigin string) {
	if myOrigin == msg.Origin {
		fmt.Println(msg.String())
	}
}

func (msg *PrivateMessage) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		Private: msg,
	}
}

func (msg PrivateMessage) String() string {
	return fmt.Sprintf("PRIVATE origin %s hop-limit %d contents %s", msg.Origin, msg.HopLimit, msg.Text)
}

func (msg *PrivateMessage) Hopped() *PrivateMessage {
	return &PrivateMessage{
		Origin:      msg.Origin,
		Id:          msg.Id,
		Text:        msg.Text,
		Destination: msg.Destination,
		HopLimit:    msg.HopLimit - 1,
	}
}
