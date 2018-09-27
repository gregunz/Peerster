package communication

import (
	"fmt"
	"github.com/gregunz/Peerster/models"
)

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

func ClientBroadcast(message *SimpleMessage, senderName string, senderAddr models.Address) *SimpleMessage {
	message.OriginalName = senderName
	message.RelayPeerAddr = senderAddr.String()
	return message
}

func PeerBroadcast(message *SimpleMessage, senderAddr models.Address) (*SimpleMessage, models.Address) {
	relayPeerAddr := models.StringToAddress(message.RelayPeerAddr)
	message.RelayPeerAddr = senderAddr.String()
	return message, relayPeerAddr
}

func AckClientMessage(message *SimpleMessage) {
	fmt.Printf("CLIENT MESSAGE %s\n", message.Contents)
}

func AckPeerMessage(message *SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		message.OriginalName,
		message.RelayPeerAddr,
		message.Contents)
}
