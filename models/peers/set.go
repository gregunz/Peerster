package peers

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"math/rand"
	"strings"
	"sync"
)

type Set struct {
	peersMap  map[string]*Peer
	PeersChan chan *Peer
	mux       sync.Mutex
}

func NewPeersSet(peers ...*Peer) *Set {
	newPeersSet := &Set{}
	newPeersSet.init()
	for _, p := range peers {
		newPeersSet.add(p)
	}
	return newPeersSet
}

func (set *Set) init() {
	alreadyInit := true
	if set.peersMap == nil {
		set.peersMap = make(map[string]*Peer)
		alreadyInit = false
	}
	if set.PeersChan == nil {
		set.PeersChan = make(chan *Peer, 1)
		alreadyInit = false
	}
	if alreadyInit {
		common.HandleError(fmt.Errorf("peers Set already initialized"))
	}

}

func (set *Set) string() string {
	return fmt.Sprintf("PEERS %s", set.toString(","))
}

func (set Set) String() string {
	set.mux.Lock()
	defer set.mux.Unlock()

	return set.string()
}

func (set *Set) Set(s string) error {
	set.mux.Lock()
	defer set.mux.Unlock()

	if set.peersMap == nil {
		set.init()
	}

	for _, ipPort := range strings.Split(s, ",") {
		set.addIpPort(ipPort)
	}
	return nil
}

func (set *Set) toStrings() []string {
	ls := []string{}
	for _, p := range set.peersMap {
		ls = append(ls, p.Addr.ToIpPort())
	}
	return ls
}

func (set *Set) ToStrings() []string {
	set.mux.Lock()
	defer set.mux.Unlock()

	return set.toStrings()
}

func (set Set) toString(sep string) string {
	return strings.Join(set.toStrings(), sep)
}

func (set Set) ToString(sep string) string {
	set.mux.Lock()
	defer set.mux.Unlock()

	return set.toString(sep)
}

func (set *Set) addIpPort(ipPort string) *Peer {
	peer := NewPeer(ipPort)
	set.add(peer)
	return peer
}

func (set *Set) AddIpPort(ipPort string) *Peer {
	set.mux.Lock()
	defer set.mux.Unlock()

	return set.addIpPort(ipPort)
}

func (set *Set) add(peer *Peer) {
	if set.peersMap == nil {
		set.init()
	}
	if peer == nil {
		common.HandleAbort("not adding nil to PeerSet", nil)
		return
	}
	_, ok := set.peersMap[peer.ID()]
	if ok {
		// not overwriting if peer already present
		common.HandleError(fmt.Errorf("adding a Peer that is already in PeerSet"))
	} else {
		go func() { set.PeersChan <- peer }()
		set.peersMap[peer.ID()] = peer
	}
}

func (set *Set) Add(peer *Peer) {
	set.mux.Lock()
	defer set.mux.Unlock()
	set.add(peer)
}

func (set *Set) getSlice() []*Peer {
	peersList := []*Peer{}
	for _, p := range set.peersMap {
		peersList = append(peersList, p)
	}
	return peersList
}

func (set *Set) GetSlice() []*Peer {
	set.mux.Lock()
	defer set.mux.Unlock()

	return set.getSlice()
}

func (set *Set) filter(peer ...*Peer) *Set {
	newPeersSet := NewPeersSet()
	for _, p := range set.peersMap {
		isNotFiltered := true
		for _, filteredPeer := range peer {
			if p.ID() == filteredPeer.ID() {
				isNotFiltered = false
				break
			}
		}
		if isNotFiltered {
			newPeersSet.add(p)
		}
	}
	return newPeersSet
}

func (set *Set) Filter(peer ...*Peer) *Set {
	set.mux.Lock()
	defer set.mux.Unlock()

	return set.filter(peer...)
}

func (set *Set) GetRandom(except ...*Peer) *Peer {
	set.mux.Lock()
	defer set.mux.Unlock()

	peersSetCopy := set.filter(except...)
	if len(peersSetCopy.peersMap) > 0 {
		idx := rand.Int() % len(peersSetCopy.peersMap)
		return peersSetCopy.getSlice()[idx]
	}
	return nil
}

func (set *Set) AckPrint() {
	set.mux.Lock()
	defer set.mux.Unlock()

	if set.nonEmpty() {
		fmt.Println(set.string())
	}
}

func (set *Set) isEmpty() bool {
	return len(set.peersMap) == 0
}

func (set *Set) IsEmpty() bool {
	set.mux.Lock()
	defer set.mux.Unlock()
	return set.isEmpty()
}

func (set *Set) nonEmpty() bool {
	return !set.isEmpty()
}

func (set *Set) NonEmpty() bool {
	set.mux.Lock()
	defer set.mux.Unlock()

	return set.nonEmpty()
}

func (set *Set) Extend(other *Set) {
	set.mux.Lock()
	other.mux.Lock()
	defer set.mux.Unlock()
	defer other.mux.Unlock()

	for k, v := range other.peersMap {
		set.peersMap[k] = v
	}
}

func (set *Set) Union(other *Set) *Set {
	newPeersSet := NewPeersSet()
	newPeersSet.Extend(set)
	newPeersSet.Extend(other)
	return newPeersSet
}

func (set *Set) get(ipPort string) (*Peer, error) {
	p, ok := set.peersMap[ipPort]
	if !ok {
		return nil, fmt.Errorf("trying to Get a Peer that is not in PeerSet")
	}
	return p, nil
}

func (set *Set) Get(ipPort string) *Peer {
	set.mux.Lock()
	defer set.mux.Unlock()

	peer, err := set.get(ipPort)
	common.HandleError(err)
	return peer
}

func (set *Set) GetAndError(ipPort string) (*Peer, error) {
	set.mux.Lock()
	defer set.mux.Unlock()

	return set.get(ipPort)
}

func (set *Set) Has(ipPort string) bool {
	set.mux.Lock()
	defer set.mux.Unlock()

	_, ok := set.peersMap[ipPort]
	return ok
}
