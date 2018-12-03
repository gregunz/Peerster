package peers

import (
	"strings"
)

type SetVar struct {
	peers []*Peer
}

func (set *SetVar) Set(s string) error {
	if set.peers == nil {
		set.peers = []*Peer{}
	}

	for _, ipPort := range strings.Split(s, ",") {
		peer := NewPeer(ipPort)
		set.peers = append(set.peers, peer)
	}
	return nil
}

func (set *SetVar) String() string {
	ls := []string{}
	for _, p := range set.peers {
		ls = append(ls, p.Addr.ToIpPort())
	}
	return strings.Join(ls, ",")
}

func (set *SetVar) ToSet() *Set {
	return NewSet(set.peers...)
}
