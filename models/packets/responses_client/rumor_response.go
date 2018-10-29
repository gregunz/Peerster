package responses_client

import "github.com/gregunz/Peerster/models/packets/packets_gossiper"

type RumorResponse struct {
	Message *packets_gossiper.RumorMessage `json:"message"`
}
