package files

import (
	"crypto/sha256"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
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

type FileType struct {
	Name     string
	Path     string
	Size     uint64
	Hashes   []string // list of chunk hashes
	MetaFile []byte
	MetaHash string // hash of metafile
}

func NewFile(filename string, path string) *FileType {
	filepath := path + filename
	fileBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		common.HandleAbort(fmt.Sprintf("could not read %s", filepath), err)
		return nil
	}

	fileSize := uint64(len(fileBytes))
	nChunks := int(math.Ceil(float64(fileSize) / ChunkSize))
	metafile := []byte{}
	hashes := []string{}
	for i := 0; i < nChunks; i++ {
		from := i * ChunkSize
		to := utils.Min(uint64(i+1)*ChunkSize, fileSize)
		hash := sha256.Sum256(fileBytes[from:to])
		hashes = append(hashes, utils.HashToHex(hash[:]))
		metafile = append(metafile, hash[:]...)
	}

	metaHash := sha256.Sum256(metafile)

	return &FileType{
		Name:     filename,
		Path:     path,
		Size:     fileSize,
		Hashes:   hashes,
		MetaFile: metafile,
		MetaHash: utils.HashToHex(metaHash[:]),
	}
}

func (file *FileType) FilePath() string {
	return file.Path + file.Name
}

func (file *FileType) GetChunkOrMetafile(hash string) ([]byte, error) {
	if hash == file.MetaHash {
		return file.MetaFile, nil
	}
	for i, h := range file.Hashes {
		if h == hash {
			from := i * ChunkSize
			to := utils.Min(uint64(i+1)*ChunkSize, file.Size)
			fileBytes, err := ioutil.ReadFile(file.FilePath())
			if err != nil {
				return nil, err
			}
			return fileBytes[from:to], nil
		}
	}
	return nil, fmt.Errorf("no chunks with corresponding hash %s in file %s", hash, file.FilePath())
}

func (file *FileType) ToSearchResult() *packets_gossiper.SearchResult {
	chunkMap := []uint64{}
	for i, _ := range file.Hashes {
		chunkMap = append(chunkMap, uint64(i+1)) // + 1 because zero is reserved for metafile
	}
	return &packets_gossiper.SearchResult{
		FileName:     file.Name,
		MetafileHash: utils.HexToHash(file.MetaHash),
		ChunkMap:     chunkMap,
		ChunkCount:   uint64(len(file.Hashes)),
	}
}
