package files

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/utils"
	"sync"
)

type Uploader struct {
	filenameToFile map[string]*FileType
	chunksToFile   map[string]*FileType
	FileChan       StoredFileChan
	mux            sync.Mutex
}

func NewFilesUploader(activateChan bool) *Uploader {
	return &Uploader{
		filenameToFile: map[string]*FileType{},
		chunksToFile:   map[string]*FileType{},
		FileChan:       NewFileChan(activateChan),
	}
}

func (uploader *Uploader) IndexFile(filename string, isSharedPath bool) *FileType {
	uploader.mux.Lock()
	defer uploader.mux.Unlock()

	var path string
	if isSharedPath {
		path = sharedPath
	} else {
		path = downloadsPath
	}
	file := NewFile(filename, path)
	if file == nil {
		return nil
	}
	_, ok := uploader.chunksToFile[file.MetaHash]
	if ok {
		common.HandleAbort("file is already indexed", nil)
		return nil
	}
	uploader.chunksToFile[file.MetaHash] = file
	uploader.filenameToFile[filename] = file
	uploader.FileChan.Push(file)
	logger.Printlnf("new file named %s indexed with hash %s", file.Name, file.MetaHash)
	for _, hash := range file.Hashes {
		if _, ok := uploader.chunksToFile[hash]; ok {
			common.HandleError(fmt.Errorf("collision of hashes of some indexed files"))
		}
		uploader.chunksToFile[hash] = file
	}
	return file
}

func (uploader *Uploader) GetAllFiles() []*FileType {
	uploader.mux.Lock()
	defer uploader.mux.Unlock()

	var files []*FileType
	for _, file := range uploader.filenameToFile {
		files = append(files, file)
	}
	return files
}

func (uploader *Uploader) HasChunk(chunkHash []byte) bool {
	uploader.mux.Lock()
	defer uploader.mux.Unlock()

	hashString := utils.HashToHex(chunkHash)
	_, ok := uploader.chunksToFile[hashString]
	return ok
}

func (uploader *Uploader) GetData(chunkHash []byte) []byte {
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
