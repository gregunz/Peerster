package models

import (
	"fmt"
	"strings"
	"sync"
)

type Peers struct {
	peers map[string]Peer
	mux   sync.Mutex
}

func (peers *Peers) String() string {
	return fmt.Sprint(*peers)
}

func (peers *Peers) Set(value string) error {
	if peers.peers == nil {
		peers.peers = make(map[string]Peer)
	}
	for _, ipPort := range strings.Split(value, ",") {
		peers.AddPeer(ipPort)
	}
	return nil
}

func (peers Peers) ToStrings() []string {
	ls := []string{}
	for _, a := range peers.peers {
		ls = append(ls, a.String())
	}
	return ls
}

func (peers Peers) ToString(sep string) string {
	return strings.Join(peers.ToStrings(), sep)
}

func (peers *Peers) AddPeer(peer string) {
	peers.mux.Lock()
	peers.peers[peer] = StringToPeer(peer)
	peers.mux.Unlock()
}

func EmptyPeers() *Peers {
	return &Peers{
		peers: make(map[string]Peer),
	}
}

/*
func (peers *Peers) GetPeersList() []Peer {
	peersList := []Peer{}
	for _, p := range peers.peers {
		peersList = append(peersList, p)
	}
	return peersList
}
*/

func (peers *Peers) Filter(peer Peer) *Peers {
	peersMap := make(map[string]Peer)
	for ipPort, p := range peers.peers {
		if ipPort != peer.String() {
			peersMap[ipPort] = p
		}
	}
	return &Peers{
		peers: peersMap,
	}
}
