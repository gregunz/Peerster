package files

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/utils"
	"sync"
)

type uploader struct {
	filenameToFile map[string]*FileType
	chunksToFile   map[string]*FileType
	FileChan       FileChan
	mux            sync.Mutex
}

type Uploader interface {
	IndexFile(filename string)
	HasChunk(chunkHash []byte) bool
	GetData(chunkHash []byte) []byte
	GetAllFiles() []*FileType
}

func NewFilesUploader(activateChan bool) *uploader {
	return &uploader{
		filenameToFile: map[string]*FileType{},
		chunksToFile:   map[string]*FileType{},
		FileChan:       NewFileChan(activateChan),
	}
}

func (uploader *uploader) IndexFile(filename string) {
	uploader.mux.Lock()
	defer uploader.mux.Unlock()

	file := NewFile(nameToSharedPath(filename))
	if file == nil {
		return
	}
	_, ok := uploader.chunksToFile[file.MetaHash]
	if ok {
		common.HandleAbort("file is already indexed", nil)
		return
	}
	uploader.chunksToFile[file.MetaHash] = file
	uploader.filenameToFile[filename] = file
	uploader.FileChan.Push(file)
	logger.Printlnf("new file indexed with hash %s", file.MetaHash)
	for _, hash := range file.Hashes {
		if _, ok := uploader.chunksToFile[hash]; ok {
			common.HandleError(fmt.Errorf("collision of hashes of some indexed files"))
		}
		uploader.chunksToFile[hash] = file
	}
}

func (uploader *uploader) GetAllFiles() []*FileType {
	uploader.mux.Lock()
	defer uploader.mux.Unlock()

	var files []*FileType
	for _, file := range uploader.filenameToFile {
		files = append(files, file)
	}
	return files
}

func (uploader *uploader) HasChunk(chunkHash []byte) bool {
	uploader.mux.Lock()
	defer uploader.mux.Unlock()

	hashString := utils.HashToHex(chunkHash)
	_, ok := uploader.chunksToFile[hashString]
	return ok
}

func (uploader *uploader) GetData(chunkHash []byte) []byte {
	uploader.mux.Lock()
	defer uploader.mux.Unlock()

	hashString := utils.HashToHex(chunkHash)
	file, ok := uploader.chunksToFile[hashString]
	if !ok {
		common.HandleAbort(fmt.Sprintf("could not find the chunk with hash: %s", hashString), nil)
		return nil
	}
	data, err := file.GetChunkOrMetafile(hashString)
	if err != nil {
		common.HandleError(err)
		return nil
	}
	return data
}
