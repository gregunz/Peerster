package packets_gossiper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/utils"
)

type File struct {
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	MetafileHash []byte `json:"metafile-hash"`
}

func (file *File) String() string {
	return fmt.Sprintf("FILE named %s size %d with metafile hash %s",
		file.Name, file.Size, utils.HashToHex(file.MetafileHash))
}

func (file *File) Hash() (out [32]byte) {
	h := sha256.New()
	err := binary.Write(h, binary.LittleEndian, uint32(len(file.Name)))
	if err != nil {
		common.HandleAbort("unexpected error when computing hash of tx-publish", err)
		return
	}
	h.Write([]byte(file.Name))
	h.Write(file.MetafileHash)
	copy(out[:], h.Sum(nil))
	return
}
