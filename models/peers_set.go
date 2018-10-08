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

func NewPeersSet(peers ...*Peer) *PeersSet {
	newPeersSet := &PeersSet{}
	newPeersSet.init()
	for _, p := range peers {
		newPeersSet.addPeer(p)
	}
	return newPeersSet
}

func (peersSet *PeersSet) init() {
	if peersSet.peersMap == nil {
		peersSet.peersMap = make(map[string]*Peer)
	} else {
		common.HandleError(fmt.Errorf("PeersSet already initialized"))
	}
}

func (peersSet *PeersSet) string() string {
	return fmt.Sprintf("PEERS %s", peersSet.toString(","))
}

func (peersSet PeersSet) String() string {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	return peersSet.string()
}

func (peersSet *PeersSet) Set(s string) error {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	if peersSet.peersMap == nil {
		peersSet.init()
	}

	for _, ipPort := range strings.Split(s, ",") {
		peersSet.addIpPort(ipPort)
	}
	return nil
}

func (peersSet *PeersSet) toStrings() []string {
	ls := []string{}
	for _, p := range peersSet.peersMap {
		ls = append(ls, p.Addr.ToIpPort())
	}
	return ls
}

func (peersSet *PeersSet) ToStrings() []string {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	return peersSet.toStrings()
}

func (peersSet PeersSet) toString(sep string) string {
	return strings.Join(peersSet.toStrings(), sep)
}


func (peersSet PeersSet) ToString(sep string) string {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	return peersSet.toString(sep)
}

func (peersSet *PeersSet) addIpPort(ipPort string) {
	peersSet.addPeer(NewPeer(ipPort))
}

func (peersSet *PeersSet) AddIpPort(ipPort string) {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	peersSet.addIpPort(ipPort)
}

func (peersSet *PeersSet) addPeer(peer *Peer) {
	if peersSet.peersMap == nil {
		peersSet.init()
	}
	_, ok := peersSet.peersMap[peer.ID()]
	if ok {
		// not overwriting if peer already present
		common.HandleError(fmt.Errorf("adding a Peer that is already in PeerSet"))
	} else {
		peersSet.peersMap[peer.ID()] = peer
	}
}

func (peersSet *PeersSet) AddPeer(peer *Peer) {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()
	peersSet.addPeer(peer)
}

func (peersSet *PeersSet) getSlice() []*Peer {
	peersList := []*Peer{}
	for _, p := range peersSet.peersMap {
		peersList = append(peersList, p)
	}
	return peersList
}

func (peersSet *PeersSet) GetSlice() []*Peer {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	return peersSet.getSlice()
}


func (peersSet *PeersSet) filter(peer ...*Peer) *PeersSet {
	newPeersSet := NewPeersSet()
	for _, p := range peersSet.peersMap {
		isNotFiltered := true
		for _, filteredPeer := range peer {
			if p.ID() == filteredPeer.ID() {
				isNotFiltered = false
				break
			}
		}
		if isNotFiltered {
			newPeersSet.addPeer(p)
		}
	}
	return newPeersSet
}

func (peersSet *PeersSet) Filter(peer ...*Peer) *PeersSet {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	return peersSet.filter(peer...)
}

func (peersSet *PeersSet) GetRandom(except ...*Peer) *Peer {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	peersSetCopy := peersSet.filter(except...)
	idx := rand.Int() % len(peersSetCopy.peersMap)
	return peersSet.getSlice()[idx]

}

func (peersSet *PeersSet) ToStatusPacket() *StatusPacket {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	want := []PeerStatus{}
	for _, p := range peersSet.peersMap {
		want = append(want, *p.ToPeerStatus())
	}
	return &StatusPacket{
		Want: want,
	}
}

/*
func (peers *PeersSet) SaveRumor(msg *RumorMessage, fromPeer *Peer) {
	peers.mux.Lock()
	defer peers.mux.Unlock()

	peer := peers.peersMap[fromPeer.ID()]
	peer.SaveRumor(msg)
}
*/

func (peersSet *PeersSet) AckPrint(filterMe *Peer) {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	filteredPeerSet := peersSet.filter(filterMe)
	if filteredPeerSet.nonEmpty() {
		fmt.Println(filteredPeerSet.string())
	}
}

func (peersSet *PeersSet) isEmpty() bool {
	return len(peersSet.peersMap) == 0
}

func (peersSet *PeersSet) IsEmpty() bool {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()
	return peersSet.isEmpty()
}

func (peersSet *PeersSet) nonEmpty() bool {
	return !peersSet.isEmpty()
}

func (peersSet *PeersSet) NonEmpty() bool {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	return peersSet.nonEmpty()
}

func (peersSet *PeersSet) Extend(other *PeersSet) {
	peersSet.mux.Lock()
	other.mux.Lock()
	defer peersSet.mux.Unlock()
	defer other.mux.Unlock()

	for k, v := range other.peersMap {
		peersSet.peersMap[k] = v
	}
}

func (peersSet *PeersSet) Union(other *PeersSet) *PeersSet {
	newPeersSet := NewPeersSet()
	newPeersSet.Extend(peersSet)
	newPeersSet.Extend(other)
	return newPeersSet
}

func (peersSet *PeersSet) Compare(want []PeerStatus) (*RumorMessage, bool) {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	msgToSend := []*RumorMessage{}
	remoteHasMsg := false

	for _, ps := range want {
		peer, ok := peersSet.peersMap[ps.Identifier]
		if ok && peer.LatestID >= ps.NextID {
			msgToSend = append(msgToSend, peer.Rumors[ps.NextID])
		} else if ok && peer.LatestID < ps.NextID - 1 {
			remoteHasMsg = true
		}
	}

	var randomMsg *RumorMessage = nil
	if len(msgToSend) > 0 {
		randomMsg = msgToSend[rand.Int() % len(msgToSend)]
	}
	return randomMsg, remoteHasMsg
}

func (peersSet *PeersSet) get(ipPort string) (*Peer, error) {
	p, ok := peersSet.peersMap[ipPort]
	if !ok {
		return nil, fmt.Errorf("trying to Get a Peer that is not in PeerSet")
	}
	return p, nil
}

func (peersSet *PeersSet) Get(ipPort string) *Peer {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	peer, err := peersSet.get(ipPort)
	common.HandleError(err)
	return peer
}

func (peersSet *PeersSet) GetAndError(ipPort string) (*Peer, error) {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	return peersSet.get(ipPort)
}

func (peersSet *PeersSet) Has(ipPort string) bool {
	peersSet.mux.Lock()
	defer peersSet.mux.Unlock()

	_, ok := peersSet.peersMap[ipPort]
	return ok
}
