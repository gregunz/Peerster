package responses_client

import "github.com/gregunz/Peerster/models/packets/packets_gossiper"

type ClientResponse struct {
	GetId *GetIdResponse `json:"get-id"`
	Peer  *PeerResponse  `json:"peer"`
	Rumor *RumorResponse `json:"rumor"`
}

func NewGetIdResponse(id string) *ClientResponse {
	return &ClientResponse{GetId: &GetIdResponse{Id: id}}
}

func NewPeerResponse(address string) *ClientResponse {
	return &ClientResponse{Peer: &PeerResponse{Address: address}}
}

func NewRumorResponse(rumor *packets_gossiper.RumorMessage) *ClientResponse {
	return &ClientResponse{Rumor: &RumorResponse{Message: rumor}}
}
