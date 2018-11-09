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
	rootPath      = "../"
	sharedPath    = rootPath + "_SharedFiles/"
	downloadsPath = rootPath + "_Downloads/"
	ChunkSize     = 8192
	HashSize      = sha256.Size
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
	nChunks := int(math.Ceil(float64(fileSize) / ChunkSize))
	var metafile []byte
	for i := 0; i < nChunks; i++ {
		from := i * ChunkSize
		to := utils.Min((i+1)*ChunkSize, fileSize)
		hash := sha256.Sum256(fileBytes[from:to])
		metafile = append(metafile, hash[:]...)
	}

	return &fileType{
		name:     name,
		Size:     fileSize,
		Metafile: metafile,
	}
}

func (file *fileType) Hash() [HashSize]byte {
	return sha256.Sum256(file.Metafile)
}

func (file *fileType) NumChunks() int {
	return len(file.Metafile) / HashSize
}

func (file *fileType) GetAllHashes() map[string]*fileType {
	hashes := map[string]*fileType{}
	nChunks := file.NumChunks()
	for i := 0; i < nChunks; i++ {
		fromMeta := i * HashSize
		toMeta := (i + 1) * HashSize
		hashes[utils.HashToHex(file.Metafile[fromMeta:toMeta])] = file
	}
	return hashes
}

func (file *fileType) GetChunk(hash string) ([]byte, error) {
	nChunks := file.NumChunks()
	for i := 0; i < nChunks; i++ {
		fromMeta := i * HashSize
		toMeta := (i + 1) * HashSize
		if utils.HashToHex(file.Metafile[fromMeta:toMeta]) == hash {
			fromFile := i * ChunkSize
			toFile := utils.Min((i+1)*ChunkSize, file.Size)
			fileBytes, err := ioutil.ReadFile(nameToPath(file.name))
			if err != nil {
				return nil, err
			}
			return fileBytes[fromFile:toFile], nil
		}
	}
	return nil, fmt.Errorf("no chunks with corresponding hash: %s", string(hash[:]))
}
