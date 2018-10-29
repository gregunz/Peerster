package gossiper

import "github.com/gregunz/Peerster/models/packets/packets_gossiper"

type PrivateMsgChan interface {
	AddPrivateMsg(msg *packets_gossiper.PrivateMessage)
	GetPrivateMsg() *packets_gossiper.PrivateMessage
}
