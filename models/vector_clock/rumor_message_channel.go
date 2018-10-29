package vector_clock

import "github.com/gregunz/Peerster/models/packets/packets_gossiper"

type RumorChan interface {
	AddRumor(*packets_gossiper.RumorMessage)
	GetRumor() *packets_gossiper.RumorMessage
}
