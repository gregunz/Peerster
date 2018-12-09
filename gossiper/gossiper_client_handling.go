package gossiper

import (
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
)

func (g *Gossiper) handleClientSimpleMode(packet *packets_client.ClientPacket) {
	msg := &packets_gossiper.SimpleMessage{
		Contents:      packet.PostMessage.Message,
		RelayPeerAddr: g.GossipAddr.ToIpPort(),
		OriginalName:  g.Origin,
	}
	g.sendPacket(msg, g.PeersSet.GetSlice()...)
}

func (g *Gossiper) handleClientNormalMode(packet *packets_client.ClientPacket) {
	if packet.IsPostMessage() && packet.PostMessage.Destination == "" {
		meHandler := g.VectorClock.GetOrCreateHandler(g.Origin)
		rumorMessage := meHandler.CreateAndSaveNextMessage(packet.PostMessage.Message)
		if randomPeer := g.PeersSet.GetRandom(); randomPeer != nil {
			g.sendPacket(rumorMessage, randomPeer)
		}
	} else if packet.IsPostMessage() && packet.PostMessage.Destination != "" {
		meHandler := g.Conversations.GetOrCreateHandler(g.Origin)
		msg := meHandler.CreateAndSaveNextMessage(packet.PostMessage.Message, packet.PostMessage.Destination, hopLimit)
		g.transmit(msg, false)
	} else if packet.IsRequestFile() {
		if packet.RequestFile.Destination != "" { // download file from node (hw02)
			g.downloadHandler(packet.RequestFile)
		} else { // download a searched file (hw03)
			for _, search := range g.FilesSearcher.GetAllSearches() {
				// let's download the file now from all origins involved
				for _, requestFile := range search.ToRequestFiles(packet.RequestFile.Filename, packet.RequestFile.Request) {
					g.downloadHandler(requestFile)
				}

			}
		}
	} else if packet.IsIndexFile() {
		g.uploadHandler(packet.IndexFile.Filename, true)
	} else if packet.IsSearchFiles() {
		g.searchHandler(packet.SearchFiles)
	}
}
