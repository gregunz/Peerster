package files

import (
	"crypto/sha256"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/utils"
	"io/ioutil"
	"math"
	"path/filepath"
)

const (
	rootPath      = "./"
	sharedPath    = rootPath + "_SharedFiles/"
	downloadsPath = rootPath + "_Downloads/"
	ChunkSize     = 8192
	HashSize      = sha256.Size
)

type FileType struct {
	Name     string
	Size     int
	Hashes   []string
	MetaFile []byte
	MetaHash string
}

func nameToSharedPath(name string) string {
	return sharedPath + name
}

func nameToDownloadsPath(name string) string {
	return downloadsPath + name
}

func NewFile(path string) *FileType {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		common.HandleAbort(fmt.Sprintf("could not read %s", path), err)
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

	return &FileType{
		Name:     filepath.Base(path),
		Size:     fileSize,
		Hashes:   hashes,
		MetaFile: metafile,
		MetaHash: utils.HashToHex(metaHash[:]),
	}
}

func (file *FileType) GetChunkOrMetafile(hash string) ([]byte, error) {
	if hash == file.MetaHash {
		return file.MetaFile, nil
	}
	for i, h := range file.Hashes {
		if h == hash {
			from := i * ChunkSize
			to := utils.Min((i+1)*ChunkSize, file.Size)
			fileBytes, err := ioutil.ReadFile(nameToSharedPath(file.Name))
			if err != nil {
				return nil, err
			}
			return fileBytes[from:to], nil
		}
	}
	return nil, fmt.Errorf("no chunks with corresponding hash %s in file %s", hash, file.Name)
}
