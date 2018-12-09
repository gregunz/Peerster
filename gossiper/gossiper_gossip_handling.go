package gossiper

import (
	"github.com/gregunz/Peerster/blockchain"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"github.com/gregunz/Peerster/utils"
	"strings"
)

func (g *Gossiper) handleSimple(msg *packets_gossiper.SimpleMessage, fromPeer *peers.Peer) {
	msgToSend := &packets_gossiper.SimpleMessage{
		Contents:      msg.Contents,
		RelayPeerAddr: g.GossipAddr.ToIpPort(),
		OriginalName:  msg.OriginalName,
	}
	toPeers := g.PeersSet.Filter(fromPeer).GetSlice() // not resending to sender

	g.sendPacket(msgToSend, toPeers...)
}

func (g *Gossiper) handleRumor(msg *packets_gossiper.RumorMessage, fromPeer *peers.Peer) {

	// saving message
	g.VectorClock.GetOrCreateHandler(msg.Origin).Save(msg)
	if msg.Origin != g.Origin {
		g.RoutingTable.GetOrCreateHandler(msg.Origin).AckRumor(msg, fromPeer)
	}

	msgToSend := &packets_gossiper.RumorMessage{
		ID:     msg.ID,
		Text:   msg.Text,
		Origin: msg.Origin,
	}

	// sendPacket to a random peer
	if randomPeer := g.PeersSet.GetRandom(fromPeer); randomPeer != nil {
		g.sendPacket(msgToSend, randomPeer)
	}

	// send back status packet to sender (= ack of the rumor)
	g.sendPacket(g.VectorClock.ToStatusPacket(), fromPeer)
}

func (g *Gossiper) handleStatus(packet *packets_gossiper.StatusPacket, fromPeer *peers.Peer) {
	rumorMsg, remoteHasMsg := g.VectorClock.Compare(packet.ToMap())

	if rumorMsg != nil { // has a msg to send
		g.sendPacket(rumorMsg, fromPeer) // send the new message
	}
	if remoteHasMsg { // remote has new message
		g.sendPacket(g.VectorClock.ToStatusPacket(), fromPeer) // send status to remote
	}
	if rumorMsg == nil && !remoteHasMsg { // is up to date
		fromPeer.FlipTimeout.Trigger()
		if !g.debug {
			logger.Printlnf("IN SYNC WITH %s", fromPeer.Addr.ToIpPort())
		}
	} else {
		fromPeer.FlipTimeout.Cancel()
	}
}

func (g *Gossiper) transmit(packetToTransmit packets_gossiper.Transmittable, decreaseHop bool, toPeers ...*peers.Peer) {
	if decreaseHop {
		packetToTransmit = packetToTransmit.Hopped()
	}
	if packetToTransmit.IsTransmittable() && packetToTransmit.Dest() != g.Origin {
		toPeer := g.RoutingTable.GetOrCreateHandler(packetToTransmit.Dest()).GetPeer()
		if toPeer != nil {
			toPeers = append(toPeers, toPeer)
		}
		g.sendPacket(packetToTransmit, toPeer)
	}
}

func (g *Gossiper) handlePrivate(msg *packets_gossiper.PrivateMessage) {
	if msg.Destination != g.Origin {
		g.transmit(msg, true)
	} else { // message is for us
		g.Conversations.GetOrCreateHandler(msg.Origin).Save(msg)
	}
}

func (g *Gossiper) handleDataRequest(packet *packets_gossiper.DataRequest) {
	if packet.Destination != g.Origin {
		g.transmit(packet, true)
	} else { // packet is for us
		if g.FilesUploader.HasChunk(packet.HashValue) {
			data := g.FilesUploader.GetData(packet.HashValue)
			reply := &packets_gossiper.DataReply{
				Origin:      g.Origin,
				Destination: packet.Origin,
				HopLimit:    hopLimit,
				HashValue:   packet.HashValue,
				Data:        data,
			}
			g.transmit(reply, false)
		}
	}
}

func (g *Gossiper) handleDataReply(packet *packets_gossiper.DataReply) {
	if packet.Destination != g.Origin {
		g.transmit(packet, true)
	} else { // packet is for us
		dataHash := utils.HashToHex(packet.HashValue)
		output := g.FilesDownloader.AddChunkOrMetafile(dataHash, packet.Data)
		awaitingHashes, index, filename, fileIsBuilt := output.AwaitingMetafile, output.ChunkIndex, output.FileName, output.FileIsBuilt
		if index == 0 { // metafile
			logger.Printlnf("DOWNLOADING metafile of %s from %s", filename, packet.Origin)
		} else if index > 0 { // chunk
			logger.Printlnf("DOWNLOADING %s chunk %d from %s", filename, index, packet.Origin)
		}
		if len(awaitingHashes) > 0 {
			for _, hashString := range awaitingHashes {
				packetToSend := &packets_gossiper.DataRequest{
					Origin:      g.Origin,
					Destination: packet.Origin,
					HopLimit:    hopLimit,
					HashValue:   utils.HexToHash(hashString),
				}
				g.transmit(packetToSend, false)
			}
		} else if fileIsBuilt { // download complete
			g.uploadHandler(filename, false)
		} else {
			logger.Printlnf("already received this data (hash=%s)", dataHash)
		}
	}

}

func (g *Gossiper) handleSearchRequest(packet *packets_gossiper.SearchRequest, fromPeer *peers.Peer) {

	// forward request with budget - 1
	newPacket := &packets_gossiper.SearchRequest{
		Origin:   packet.Origin,
		Budget:   packet.Budget - 1,
		Keywords: packet.Keywords,
	}
	g.sendBudgetPacket(newPacket, fromPeer)

	// reply to the request given the matching results

	matchingResults := g.FilesDownloader.GetAllSearchResults(packet.Keywords)
	for _, indexedFile := range g.FilesUploader.GetAllFiles() {
		if utils.Match(indexedFile.Name, packet.Keywords) {
			matchingResults = append(matchingResults, indexedFile.ToSearchResult())
		}
	}

	if len(matchingResults) > 0 {
		reply := &packets_gossiper.SearchReply{
			Origin:      g.Origin,
			Destination: packet.Origin,
			HopLimit:    hopLimit,
			Results:     matchingResults,
		}
		g.transmit(reply, false)
	}
}

func (g *Gossiper) handleSearchReply(packet *packets_gossiper.SearchReply) {
	if packet.Destination == g.Origin {
		g.FilesSearcher.Ack(packet)
		for _, search := range g.FilesSearcher.GetLatestFullSearches() {
			// hw03 print
			logger.Printlnf("SEARCH FINISHED")
			// print with more details
			logger.Printlnf("SEARCH with keywords %s FINISHED with budget %d",
				strings.Join(search.Keywords, ","), search.Budget)
		}
	} else {
		g.transmit(packet, true)
	}
}

func (g *Gossiper) handleTxPublish(packet *packets_gossiper.TxPublish, fromPeer *peers.Peer) {
	g.transmit(packet, true, g.PeersSet.Filter(fromPeer).GetSlice()...)
	g.BlockChainFile.AddTx(blockchain.NewTx(packet))
}

func (g *Gossiper) handleBlockPublish(packet *packets_gossiper.BlockPublish, fromPeer *peers.Peer) {
	g.transmit(packet, true, g.PeersSet.Filter(fromPeer).GetSlice()...)
	g.BlockChainFile.AddBlock(&packet.Block)
}
