package packets_gossiper

import (
	"fmt"
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
