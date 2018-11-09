package files

import (
	"crypto/sha256"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/utils"
	"io/ioutil"
)

type fileBuilder struct {
	name         string
	hashList     []string // necessary for ordering of chunks
	hashToChunks map[string][]byte
}

func NewFileBuilder(name string, metafile []byte) *fileBuilder {
	nChunks := len(metafile) / HashSize
	hashToChunks := map[string][]byte{}
	hashList := []string{}
	for i := 0; i < nChunks; i++ {
		from := i * HashSize
		to := (i + 1) * HashSize
		hashString := utils.HashToHex(metafile[from:to])
		hashList = append(hashList, hashString)
		hashToChunks[hashString] = nil
	}
	return &fileBuilder{
		name:         name,
		hashList:     hashList,
		hashToChunks: hashToChunks,
	}
}

func (file *fileBuilder) IsComplete() bool {
	for _, f := range file.hashToChunks {
		if f == nil {
			return false
		}
	}
	return true
}

func (file *fileBuilder) HashOfMissingChunks() [][]byte {
	missingChunks := [][]byte{}
	for h, f := range file.hashToChunks {
		if f == nil {
			missingChunks = append(missingChunks, utils.HexToHash(h))
		}
	}
	return missingChunks
}

func (file *fileBuilder) AddChunks(chunks ...[]byte) {
	for _, chunk := range chunks {
		hash := sha256.Sum256(chunk)
		hashString := utils.HashToHex(hash[:])
		if _, ok := file.hashToChunks[hashString]; ok {
			file.hashToChunks[hashString] = chunk
		}
	}
}

func (file *fileBuilder) Build() bool {
	fileBytes := []byte{}
	for _, hash := range file.hashList {
		chunk := file.hashToChunks[hash]
		if chunk == nil {
			common.HandleAbort("cannot build because some chunks are missing", nil)
			return false
		}
		fileBytes = append(fileBytes, chunk...)
	}
	path := downloadsPath + file.name
	err := ioutil.WriteFile(path, fileBytes, 0644)
	if err != nil {
		common.HandleAbort(fmt.Sprintf("cannot save file in %s", path), err)
		return false
	}
	return true
}