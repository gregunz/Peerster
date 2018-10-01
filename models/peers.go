package models

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"strings"
	"sync"
)

type PeersSet struct {
	peersMap map[string]Peer
	mux      sync.Mutex
}

func (peers *PeersSet) String() string {
	return fmt.Sprint(*peers)
}

func (peers *PeersSet) Set(value string) error {
	if peers.peersMap == nil {
		peers.init()
	}
	for _, ipPort := range strings.Split(value, ",") {
		peers.AddIpPort(ipPort)
	}
	return nil
}

func (peers PeersSet) ToStrings() []string {
	ls := []string{}
	for _, p := range peers.peersMap {
		ls = append(ls, p.Addr.ToIpPort())
	}
	return ls
}

func (peers PeersSet) ToString(sep string) string {
	return strings.Join(peers.ToStrings(), sep)
}

func (peers *PeersSet) AddIpPort(ipPort string) {
	peers.AddPeer(NewPeer(ipPort))
}

func (peers *PeersSet) AddPeer(peer Peer) {
	if peers.peersMap == nil {
		peers.init()
	}
	peers.mux.Lock()
	peers.peersMap[peer.Addr.ToIpPort()] = peer
	peers.mux.Unlock()
}

func NewPeersSet() *PeersSet {
	return &PeersSet{
		peersMap: make(map[string]Peer),
	}
}

func (peers *PeersSet) GetSlice() []Peer {
	peersList := []Peer{}
	for _, p := range peers.peersMap {
		peersList = append(peersList, p)
	}
	return peersList
}

func (peers *PeersSet) Filter(peer Peer) *PeersSet {
	peersMap := make(map[string]Peer)
	for ipPort, p := range peers.peersMap {
		if ipPort != peer.Addr.ToIpPort() {
			peersMap[ipPort] = p
		}
	}
	return &PeersSet{
		peersMap: peersMap,
	}
}

func (peers *PeersSet) init() {
	if peers.peersMap == nil {
		peers.peersMap = make(map[string]Peer)
	} else {
		common.HandleError(fmt.Errorf("PeersSet already initialized"))
	}
}
