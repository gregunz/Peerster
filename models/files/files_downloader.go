package files

import (
	"crypto/sha256"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/timeouts"
	"github.com/gregunz/Peerster/utils"
	"os"
	"sync"
	"time"
)

const (
	timeoutDuration = 5 * time.Second
)

type downloader struct {
	awaitingMetafiles             map[string]*awaitingMetafile
	downloadedMetafilesToFilename map[string]string
	awaitingChunks                map[string]*awaitingChunk
	FileChan                      FileChan
	mux                           sync.Mutex
}

type awaitingChunk struct {
	fileBuilder *fileBuilder
	index       int
	timeout     *timeouts.Timeout
}

type awaitingMetafile struct {
	filename string
	timeout  *timeouts.Timeout
}

type Downloader interface {
	AddNewFile(filename, hash string) bool
	AddChunkOrMetafile(hash string, data []byte) ([]string, string, int)
	SetTimeout(hash string, callback func())
}

func NewFilesDownloader(activateChan bool) *downloader {
	return &downloader{
		awaitingMetafiles:             map[string]*awaitingMetafile{},
		downloadedMetafilesToFilename: map[string]string{},
		awaitingChunks:                map[string]*awaitingChunk{},
		FileChan:                      NewFileChan(activateChan),
	}
}

func (downloader *downloader) AddNewFile(filename, metafileHash string) bool {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()

	if _, ok := downloader.downloadedMetafilesToFilename[metafileHash]; ok {
		common.HandleAbort("already downloaded (or currently downloading) this file", nil)
		return false
	}
	if _, err := os.Stat(nameToDownloadsPath(filename)); !os.IsNotExist(err) {
		common.HandleAbort(fmt.Sprintf("already a file named %s in %s", filename, downloadsPath), nil)
		return false
	}
	downloader.awaitingMetafiles[metafileHash] = &awaitingMetafile{
		filename: filename,
		timeout:  timeouts.NewTimeout(),
	}
	return true
}

func (downloader *downloader) SetTimeout(hash string, callback func()) {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()
	if awaitingMetafile, ok := downloader.awaitingMetafiles[hash]; ok {
		awaitingMetafile.timeout.SetIfNotActive(timeoutDuration, callback)
	} else if awaitingChunk, ok := downloader.awaitingChunks[hash]; ok {
		awaitingChunk.timeout.SetIfNotActive(timeoutDuration, callback)
	}
}

func (downloader *downloader) AddChunkOrMetafile(hash string, data []byte) ([]string, string, int) {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()

	chunkHash := sha256.Sum256(data)
	dataHash := utils.HashToHex(chunkHash[:])
	if dataHash != hash {
		common.HandleAbort(fmt.Sprintf("data does not correspond to provided hash (%s != %s)", hash, dataHash), nil)
		return nil, "", -1
	}

	if awaitingMetafile, ok := downloader.awaitingMetafiles[hash]; ok { // received metafile

		awaitingMetafile.timeout.Cancel()
		fileBuilder := NewFileBuilder(awaitingMetafile.filename, data)
		awaitingHashes := []string{}

		for idx, h := range fileBuilder.hashList {
			chunk := &awaitingChunk{
				fileBuilder: fileBuilder,
				index:       idx + 1, // starting at 1 (zero is reserved for metafile)
				timeout:     timeouts.NewTimeout(),
			}
			downloader.awaitingChunks[h] = chunk
			awaitingHashes = append(awaitingHashes, h)
		}

		delete(downloader.awaitingMetafiles, hash)
		downloader.downloadedMetafilesToFilename[hash] = fileBuilder.name
		return awaitingHashes, fileBuilder.name, 0

	} else if awaitingChunk, ok := downloader.awaitingChunks[hash]; ok { // received chunk

		awaitingChunk.timeout.Cancel()
		builder := awaitingChunk.fileBuilder
		if builder.AddChunks(data) {
			delete(downloader.awaitingChunks, hash)
		}
		if builder.IsComplete() {
			file := builder.Build()
			if file != nil {
				logger.Printlnf("RECONSTRUCTED file %s", file.Name)
				downloader.FileChan.Push(file)
			} else {
				common.HandleError(fmt.Errorf("build of file failed"))
			}
		}
		return nil, builder.name, awaitingChunk.index
	}

	//TODO: handle no match error (no consequences for now)
	return nil, "", -1
}
