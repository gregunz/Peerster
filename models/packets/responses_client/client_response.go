package responses_client

import (
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"github.com/microcosm-cc/bluemonday"
)

type ClientResponse struct {
	GetId   *GetIdResponse                   `json:"get-id"`
	Peer    *PeerResponse                    `json:"peer"`
	Rumor   *packets_gossiper.RumorMessage   `json:"rumor"`
	Private *packets_gossiper.PrivateMessage `json:"private"`
}

func NewGetIdResponse(id string, policy *bluemonday.Policy) *ClientResponse {
	return &ClientResponse{GetId: &GetIdResponse{Id: policy.Sanitize(id)}}
}

func NewPeerResponse(peer *peers.Peer, policy *bluemonday.Policy) *ClientResponse {
	return &ClientResponse{Peer: &PeerResponse{Address: policy.Sanitize(peer.Addr.ToIpPort())}}
}

func NewRumorResponse(msg *packets_gossiper.RumorMessage, policy *bluemonday.Policy) *ClientResponse {
	return &ClientResponse{Rumor: &packets_gossiper.RumorMessage{
		Origin: policy.Sanitize(msg.Origin),
		ID:     msg.ID,
		Text:   policy.Sanitize(msg.Text),
	}}
}

func NewPrivateResponse(msg *packets_gossiper.PrivateMessage, policy *bluemonday.Policy) *ClientResponse {
	return &ClientResponse{Private: &packets_gossiper.PrivateMessage{
		Origin:      policy.Sanitize(msg.Origin),
		Id:          msg.Id,
		Text:        policy.Sanitize(msg.Text),
		Destination: policy.Sanitize(msg.Destination),
		HopLimit:    msg.HopLimit,
	}}
}
