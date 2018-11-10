package files

import (
	"crypto/sha256"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/timeouts"
	"github.com/gregunz/Peerster/utils"
	"sync"
	"time"
)

const (
	timeoutDuration = 5 * time.Second
)

type downloader struct {
	awaitingMetafiles map[string]*awaitingMetafile
	awaitingChunks    map[string]*awaitingChunk
	mux               sync.Mutex
}

type awaitingChunk struct {
	fileBuilder *fileBuilder
	timeout     *timeouts.Timeout
}

type awaitingMetafile struct {
	filename string
	timeout  *timeouts.Timeout
}

type Downloader interface {
	AddNewFile(filename, hash string)
	AddChunkOrMetafile(hash string, data []byte) []string
	SetTimeout(hash string, callback func())
}

func NewFilesDownloader() *downloader {
	return &downloader{
		awaitingMetafiles: map[string]*awaitingMetafile{},
		awaitingChunks:    map[string]*awaitingChunk{},
	}
}

func (downloader *downloader) AddNewFile(filename, hash string) {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()

	downloader.awaitingMetafiles[hash] = &awaitingMetafile{
		filename: filename,
		timeout:  timeouts.NewTimeout(),
	}
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

func (downloader *downloader) AddChunkOrMetafile(hash string, data []byte) []string {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()

	chunkHash := sha256.Sum256(data)
	if utils.HashToHex(chunkHash[:]) != hash {
		common.HandleAbort("data does not correspond to provided hash", nil)
		return nil
	}

	if awaitingMetafile, ok := downloader.awaitingMetafiles[hash]; ok { // received metafile
		fileBuilder := NewFileBuilder(awaitingMetafile.filename, data)
		awaitingHashes := []string{}

		for _, h := range fileBuilder.hashList {
			chunk := &awaitingChunk{
				fileBuilder: fileBuilder,
				timeout:     timeouts.NewTimeout(),
			}
			downloader.awaitingChunks[h] = chunk
			awaitingHashes = append(awaitingHashes, h)
		}

		delete(downloader.awaitingMetafiles, hash)
		return awaitingHashes

	} else if awaitingChunk, ok := downloader.awaitingChunks[hash]; ok { // received chunk

		awaitingChunk.timeout.Cancel()
		builder := awaitingChunk.fileBuilder
		if builder.AddChunks(data) {
			delete(downloader.awaitingChunks, hash)
		}
		if builder.IsComplete() {
			builder.Build()
		}
	}
	return nil
}
