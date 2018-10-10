package packets

import "fmt"

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

func (ps PeerStatus) String() string {
	return fmt.Sprintf("peer %s nextID %d", ps.Identifier, ps.NextID)
}
