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
	sharedPath    = "./_SharedFiles/"
	downloadsPath = "./_Downloads/"
	chunkSize     = 8000
	hashSize      = sha256.Size
)

type fileType struct {
	name     string
	Size     int
	Metafile []byte
}

func nameToPath(name string) string {
	return sharedPath + name
}

func NewFile(name string) *fileType {
	fileBytes, err := ioutil.ReadFile(nameToPath(name))
	if err != nil {
		common.HandleAbort(fmt.Sprintf("could not read %s", name), err)
	}

	fileSize := len(fileBytes)
	nChunks := int(math.Ceil(float64(fileSize) / chunkSize))
	var metafile []byte
	for i := 0; i < nChunks; i++ {
		from := i * chunkSize
		to := utils.Min((i+1)*chunkSize, fileSize)
		hash := sha256.Sum256(fileBytes[from:to])
		metafile = append(metafile, hash[:]...)
	}

	return &fileType{
		name:     name,
		Size:     fileSize,
		Metafile: metafile,
	}
}

func (file *fileType) Hash() [hashSize]byte {
	return sha256.Sum256(file.Metafile)
}

func (file *fileType) NumChunks() int {
	return len(file.Metafile) / hashSize
}

func (file *fileType) GetAllHashes() map[string]*fileType {
	hashes := map[string]*fileType{}
	nChunks := file.NumChunks()
	for i := 0; i < nChunks; i++ {
		fromMeta := i * hashSize
		toMeta := (i + 1) * hashSize
		hashes[utils.HashToHex(file.Metafile[fromMeta:toMeta])] = file
	}
	return hashes
}

func (file *fileType) GetChunk(hash string) ([]byte, error) {
	nChunks := file.NumChunks()
	for i := 0; i < nChunks; i++ {
		fromMeta := i * hashSize
		toMeta := (i + 1) * hashSize
		if utils.HashToHex(file.Metafile[fromMeta:toMeta]) == hash {
			fromFile := i * chunkSize
			toFile := utils.Min((i+1)*chunkSize, file.Size)
			fileBytes, err := ioutil.ReadFile(nameToPath(file.name))
			if err != nil {
				return nil, err
			}
			return fileBytes[fromFile:toFile], nil
		}
	}
	return nil, fmt.Errorf("no chunks with corresponding hash: %s", string(hash[:]))
}
