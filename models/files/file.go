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
	rootPath      = "./"
	sharedPath    = rootPath + "_SharedFiles/"
	downloadsPath = rootPath + "_Downloads/"
	ChunkSize     = 8192
	HashSize      = sha256.Size
)

type fileType struct {
	name     string
	Size     int
	Hashes   []string
	MetaFile []byte
	MetaHash string
}

func nameToPath(name string) string {
	return sharedPath + name
}

func NewFile(name string) *fileType {
	fileBytes, err := ioutil.ReadFile(nameToPath(name))
	if err != nil {
		common.HandleAbort(fmt.Sprintf("could not read %s", name), err)
		return nil
	}

	fileSize := len(fileBytes)
	nChunks := int(math.Ceil(float64(fileSize) / ChunkSize))
	metafile := []byte{}
	hashes := []string{}
	for i := 0; i < nChunks; i++ {
		from := i * ChunkSize
		to := utils.Min((i+1)*ChunkSize, fileSize)
		hash := sha256.Sum256(fileBytes[from:to])
		hashes = append(hashes, utils.HashToHex(hash[:]))
		metafile = append(metafile, hash[:]...)
	}

	metaHash := sha256.Sum256(metafile)

	return &fileType{
		name:     name,
		Size:     fileSize,
		Hashes:   hashes,
		MetaFile: metafile,
		MetaHash: utils.HashToHex(metaHash[:]),
	}
}

func (file *fileType) GetChunkOrMetafile(hash string) ([]byte, error) {
	if hash == file.MetaHash {
		return file.MetaFile, nil
	}
	for i, h := range file.Hashes {
		if h == hash {
			from := i * ChunkSize
			to := utils.Min((i+1)*ChunkSize, file.Size)
			fileBytes, err := ioutil.ReadFile(nameToPath(file.name))
			if err != nil {
				return nil, err
			}
			return fileBytes[from:to], nil
		}
	}
	return nil, fmt.Errorf("no chunks with corresponding hash %s in file %s", hash, file.name)
}
