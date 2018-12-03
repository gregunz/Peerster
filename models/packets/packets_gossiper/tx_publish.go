package packets_gossiper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/logger"
)

type TxPublish struct {
	File     File   `json:"file"`
	HopLimit uint32 `json:"hop-limit"`
}

func (packet *TxPublish) AckPrint() {
	logger.Printlnf(packet.String())
}

func (packet *TxPublish) ToGossipPacket() *GossipPacket {
	return &GossipPacket{
		TxPublish: packet,
	}
}

func (packet *TxPublish) String() string {
	return fmt.Sprintf("TX PUBLISH file <%s> with hop-limit %d", packet.File.String(), packet.HopLimit)
}

func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
	err := binary.Write(h, binary.LittleEndian, uint32(len(t.File.Name)))
	if err != nil {
		common.HandleAbort("unexpected error when computing hash of tx-publish", err)
		return
	}
	h.Write([]byte(t.File.Name))
	h.Write(t.File.MetafileHash)
	copy(out[:], h.Sum(nil))
	return
}
