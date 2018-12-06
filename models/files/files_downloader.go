package files

import (
	"crypto/sha256"
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/timeouts"
	"github.com/gregunz/Peerster/utils"
	"os"
	"sync"
	"time"
)

const (
	timeoutDuration = 5 * time.Second
)

type Downloader struct {
	awaitingMetafiles             map[string]*awaitingMetafile
	downloadedMetafilesToFilename map[string]string
	awaitingChunks                map[string]*awaitingChunk
	currentDownloads              map[string]*fileBuilder
	FileChan                      StoredFileChan
	mux                           sync.RWMutex
}

type awaitingMetafile struct {
	filename string
	timeouts map[string]*timeouts.Timeout
}

func (awaitingMetafile *awaitingMetafile) CancelTimeouts() {
	for _, timeout := range awaitingMetafile.timeouts {
		timeout.Cancel()
	}
}

type awaitingChunk struct {
	fileBuilder *fileBuilder
	index       int
	timeouts    map[string]*timeouts.Timeout
}

func (awaitingChunk *awaitingChunk) CancelTimeouts() {
	for _, timeout := range awaitingChunk.timeouts {
		timeout.Cancel()
	}
}

type AddChunkOrMetaFileOutput struct {
	AwaitingMetafile []string
	FileName         string
	ChunkIndex       int
	FileIsBuilt      bool
}

func NewFilesDownloader(activateChan bool) *Downloader {
	return &Downloader{
		awaitingMetafiles:             map[string]*awaitingMetafile{},
		downloadedMetafilesToFilename: map[string]string{},
		awaitingChunks:                map[string]*awaitingChunk{},
		currentDownloads:              map[string]*fileBuilder{},
		FileChan:                      NewFileChan(activateChan),
	}
}

func (downloader *Downloader) AddNewFile(filename, metafileHash string) bool {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()

	if _, ok := downloader.downloadedMetafilesToFilename[metafileHash]; ok {
		common.HandleAbort("already downloaded (or currently downloading) this file", nil)
		return false
	}
	if _, err := os.Stat(downloadsPath + filename); !os.IsNotExist(err) {
		common.HandleAbort(fmt.Sprintf("already a file named %s in %s", filename, downloadsPath), nil)
		return false
	}
	downloader.awaitingMetafiles[metafileHash] = &awaitingMetafile{
		filename: filename,
		timeouts: map[string]*timeouts.Timeout{},
	}
	return true
}

func (downloader *Downloader) getTimeout(hash, destination string) *timeouts.Timeout {
	var timeout *timeouts.Timeout
	if awaitingMetafile, ok := downloader.awaitingMetafiles[hash]; ok {
		timeout, ok = awaitingMetafile.timeouts[destination]
		if !ok {
			timeout = timeouts.NewTimeout()
			awaitingMetafile.timeouts[destination] = timeout
		}
	} else if awaitingChunk, ok := downloader.awaitingChunks[hash]; ok {
		timeout, ok = awaitingChunk.timeouts[destination]
		if !ok {
			timeout = timeouts.NewTimeout()
			awaitingChunk.timeouts[destination] = timeout
		}
	}
	return timeout
}

func (downloader *Downloader) SetTimeout(hash, destination string, callback func()) {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()

	timeout := downloader.getTimeout(hash, destination)
	if timeout != nil {
		timeout.SetIfNotActive(timeoutDuration, callback)
	}
}

func (downloader *Downloader) AddChunkOrMetafile(hash string, data []byte) *AddChunkOrMetaFileOutput {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()

	chunkHash := sha256.Sum256(data)
	dataHash := utils.HashToHex(chunkHash[:])
	if dataHash != hash {
		common.HandleAbort(fmt.Sprintf("data does not correspond to provided hash (%s != %s)", hash, dataHash), nil)
		return &AddChunkOrMetaFileOutput{ChunkIndex: -1}
	}

	if awaitingMetafile, ok := downloader.awaitingMetafiles[hash]; ok { // received metafile

		awaitingMetafile.CancelTimeouts()
		fileBuilder := NewFileBuilder(awaitingMetafile.filename, hash, data)
		downloader.currentDownloads[hash] = fileBuilder

		awaitingHashes := []string{}
		for idx, h := range fileBuilder.hashList {
			chunk := &awaitingChunk{
				fileBuilder: fileBuilder,
				index:       idx + 1, // starting at 1 (zero is reserved for metafile)
				timeouts:    map[string]*timeouts.Timeout{},
			}
			downloader.awaitingChunks[h] = chunk
			awaitingHashes = append(awaitingHashes, h)
		}

		//delete(downloader.awaitingMetafiles, hash)
		downloader.downloadedMetafilesToFilename[hash] = fileBuilder.name
		return &AddChunkOrMetaFileOutput{
			AwaitingMetafile: awaitingHashes,
			FileName:         fileBuilder.name,
			ChunkIndex:       0,
			FileIsBuilt:      false,
		}

	} else if awaitingChunk, ok := downloader.awaitingChunks[hash]; ok { // received chunk

		awaitingChunk.CancelTimeouts()
		builder := awaitingChunk.fileBuilder
		if builder.AddChunks(data) {
			//delete(downloader.awaitingChunks, hash)
		}
		fileIsBuilt := false
		if builder.IsComplete() {
			file := builder.Build()
			if file != nil {
				logger.Printlnf("RECONSTRUCTED file %s", file.Name)
				delete(downloader.currentDownloads, builder.metahash)
				fileIsBuilt = true
				downloader.FileChan.Push(file)
			} else {
				common.HandleError(fmt.Errorf("build of file failed"))
			}
		}
		return &AddChunkOrMetaFileOutput{
			AwaitingMetafile: nil,
			FileName:         builder.name,
			ChunkIndex:       awaitingChunk.index,
			FileIsBuilt:      fileIsBuilt,
		}
	}

	//TODO: handle no match error (no consequences for now)
	return &AddChunkOrMetaFileOutput{ChunkIndex: -1}
}

func (downloader *Downloader) GetAllSearchResults(keywords []string) []*packets_gossiper.SearchResult {
	downloader.mux.RLock()
	defer downloader.mux.RUnlock()

	results := []*packets_gossiper.SearchResult{}
	for _, builder := range downloader.currentDownloads {
		if utils.Match(builder.name, keywords) {
			results = append(results, builder.ToSearchResult())
		}
	}
	return results
}
