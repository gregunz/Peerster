package packets_gossiper

import "fmt"

type PrivateMessage struct {
	Origin      string `json:"origin"`
	ID          uint32 `json:"id"`
	Text        string `json:"text"`
	Destination string `json:"destination"`
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
	return fmt.Sprintf("PRIVATE origin %s hop-limit %d contents %s to %s", msg.Origin, msg.HopLimit, msg.Text, msg.Destination)
}

func (msg *PrivateMessage) Hopped() *PrivateMessage {
	return &PrivateMessage{
		Origin:      msg.Origin,
		ID:          msg.ID,
		Text:        msg.Text,
		Destination: msg.Destination,
		HopLimit:    msg.HopLimit - 1,
	}
}
