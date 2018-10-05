package models

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"math/rand"
	"strings"
	"sync"
)

type PeersSet struct {
	peersMap map[string]*Peer
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

func (peers *PeersSet) ToStrings() []string {
	peers.mux.Lock()
	defer peers.mux.Unlock()

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

func (peers *PeersSet) AddPeer(peer *Peer) {
	if peers.peersMap == nil {
		peers.init()
	}

	peers.mux.Lock()
	defer peers.mux.Unlock()
	peers.peersMap[peer.Addr.ToIpPort()] = peer
}

func NewEmptyPeersSet() *PeersSet {
	return &PeersSet{
		peersMap: make(map[string]*Peer),
	}
}

func (peers *PeersSet) GetSlice() []*Peer {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	peersList := []*Peer{}
	for _, p := range peers.peersMap {
		peersList = append(peersList, p)
	}
	return peersList
}

func (peers *PeersSet) Filter(peer *Peer) *PeersSet {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	peersMap := make(map[string]*Peer)
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
	peers.mux.Lock()
	defer peers.mux.Unlock()

	if peers.peersMap == nil {
		peers.peersMap = make(map[string]*Peer)
	} else {
		common.HandleError(fmt.Errorf("PeersSet already initialized"))
	}
}

func (peers *PeersSet) GetRandom() *Peer {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	idx := rand.Int() % len(peers.peersMap)
	return peers.GetSlice()[idx]

}

func (peers *PeersSet) ToStatusPacket() *StatusPacket {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	want := []PeerStatus{}
	for _, p := range peers.peersMap {
		want = append(want, *p.ToPeerStatus())
	}
	return &StatusPacket{
		Want: want,
	}
}

func (peers *PeersSet) SaveRumor(msg *RumorMessage, fromPeer *Peer) {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	peer := peers.peersMap[fromPeer.ID()]
	peer.SaveRumor(*msg)
}

func (peers *PeersSet) AckPrint() {
	if len(peers.peersMap) > 0 {
		fmt.Printf("PEERS %s", peers.ToString(","))
	}
}
