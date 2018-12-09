package gossiper

import (
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/utils"
	"strings"
	"time"
)

func (g *Gossiper) uploadHandler(filename string, fromSharedPath bool) {
	file := g.FilesUploader.IndexFile(filename, fromSharedPath)
	if file != nil {
		txPublish := &packets_gossiper.TxPublish{
			File: packets_gossiper.File{
				Name:         file.Name,
				Size:         int64(file.Size),
				MetafileHash: utils.HexToHash(file.MetaHash),
			},
			HopLimit: hopLimit,
		}
		g.sendPacket(txPublish, g.PeersSet.GetSlice()...)
	}
}

func (g *Gossiper) downloadHandler(requestFile *packets_client.RequestFilePacket) {
	// not checking if can download because it does not enable downloading from different origins yet
	canDownload := g.FilesDownloader.AddNewFile(requestFile.Filename, requestFile.Request)
	if canDownload {
		request := &packets_gossiper.DataRequest{
			Origin:      g.Origin,
			Destination: requestFile.Destination,
			HopLimit:    hopLimit,
			HashValue:   utils.HexToHash(requestFile.Request),
		}
		g.transmit(request, false)
	}
}

func (g *Gossiper) searchHandler(packet *packets_client.SearchFilesPacket) {

	search := g.FilesSearcher.Search(packet.Keywords, packet.Budget)
	g.sendBudgetPacket(&packets_gossiper.SearchRequest{
		Origin:   g.Origin,
		Budget:   search.Budget,
		Keywords: search.Keywords,
	})

	searchTicker := time.NewTicker(doublingBudgetTimeout)
	for range searchTicker.C {

		if search.IsFull() || !search.DoubleBudget() {
			logger.Printlnf("SEARCH with keywords %s STOPPED with budget %d",
				strings.Join(search.Keywords, ","), search.Budget)
			searchTicker.Stop()
		} else {
			search.DoubleBudget()
			g.sendBudgetPacket(&packets_gossiper.SearchRequest{
				Origin:   g.Origin,
				Budget:   search.Budget,
				Keywords: search.Keywords,
			})
		}
	}

}
