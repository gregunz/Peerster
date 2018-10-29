package packets_gossiper

import "fmt"

type PeerStatus struct {
	Identifier string `json:"identifier"`
	NextID     uint32 `json:"next-id"`
}

func (ps PeerStatus) String() string {
	return fmt.Sprintf("peer %s nextID %d", ps.Identifier, ps.NextID)
}
