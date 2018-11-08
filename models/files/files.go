package files

import (
	"crypto/sha256"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/utils"
	"io/ioutil"
	"math"
)

const (
	path      = "./_SharedFiles/"
	chunkSize = 8000
)

type file struct {
	Size     int
	Metafile []byte
	Hash     [32]byte
}

func nameToPath(name string) string {
	return path + name
}

func NewFile(name string) *file {
	fileBytes, err := ioutil.ReadFile(nameToPath(name))
	if err != nil {
		common.HandleAbort(fmt.Sprintf("could not read %s", name), err)
	}

	fileSize := len(fileBytes)
	nChunks := int(math.Ceil(float64(fileSize) / chunkSize))
	metafile := []byte{}
	for i := 0; i < nChunks; i++ {
		from := i * chunkSize
		to := utils.Min((i+1)*chunkSize, fileSize)
		hash := sha256.Sum256(fileBytes[from:to])
		metafile = append(metafile, hash[:]...)
	}

	return &file{
		Size:     fileSize,
		Metafile: metafile,
		Hash:     sha256.Sum256(metafile),
	}
}
