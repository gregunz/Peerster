package models

import (
	"github.com/gregunz/Peerster/utils"
)

type Peer struct {
	Addr Address
}

func NewPeer(s string) Peer {
	addr := utils.IpPortToUDPAddr(s)
	return Peer{
		Addr: Address{UDPAddr: addr},
	}
}
