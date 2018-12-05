package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/conv"
	"github.com/gregunz/Peerster/models/files"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"github.com/gregunz/Peerster/models/routing"
	"github.com/gregunz/Peerster/models/updates"
	"github.com/gregunz/Peerster/models/vector_clock"
	"github.com/gregunz/Peerster/utils"
	"net"
	"sync"
	"time"
)

const (
	//bufferedChanSize    = 1000 * 1000 // not sure if useful
	udpPacketMaxSize      = 65536
	hopLimit              = 10
	blockHopLimit         = 20
	timeoutDuration       = 1 * time.Second
	antiEntropyDuration   = 1 * time.Second
	doublingBudgetTimeout = 1 * time.Second
	debug                 = false
)

type Gossiper struct {
	mode           *Mode
	debug          bool
	clientConn     *net.UDPConn
	gossiperConn   *net.UDPConn
	rTimerDuration time.Duration

	Origin              string
	GossipAddr          *peers.Address
	ClientAddr          *peers.Address
	GUIPort             uint
	FromClientChan      chan *packets_client.ClientPacket
	FromGossipChan      chan *GossipChannelElement
	NodeChan            peers.NodeChan
	RumorChan           vector_clock.RumorChan
	OriginChan          routing.OriginChan
	PrivateMsgChan      conv.PrivateMsgChan
	PeersSet            *peers.Set
	VectorClock         vector_clock.VectorClock
	RoutingTable        routing.Table
	Conversations       conv.Conversation
	FilesUploader       files.Uploader
	IndexedFilesChan    files.FileChan
	FilesDownloader     files.Downloader
	DownloadedFilesChan files.FileChan
	FilesSearcher       files.Searcher
}

func NewGossiper(simple bool, address *peers.Address, name string, uiPort uint, guiPort uint, peersSet *peers.Set,
	rTimerDuration uint, guiEnabled bool) *Gossiper {

	logger.Printlnf("Gossiper named <%s> listening peers on ip:port <%s> "+
		"and listening local clients on port <%d> with peers <(%s)>",
		name, address.ToIpPort(), uiPort, peersSet.ToString("), ("))

	mode := NewDefaultMode()
	if simple {
		mode = NewSimpleMode()
	}

	clientAddr := peers.NewAddress(fmt.Sprintf("localhost:%d", uiPort))
	_, peerConn := utils.ConnectToIpPort(address.ToIpPort())
	_, clientConn := utils.ConnectToIpPort(clientAddr.ToIpPort())

	if name == "" || peerConn == nil || clientConn == nil {
		logger.Printlnf("could not create gossiper with those arguments")
		return nil
	}

	updatesChannels := updates.NewChannels(guiEnabled)
	peersSet.SetNodeChan(updatesChannels)
	routingTable := routing.NewTable(name, updatesChannels)
	vectorClock := vector_clock.NewVectorClock(updatesChannels)
	conversations := conv.NewConversations(updatesChannels)
	uploader := files.NewFilesUploader(guiEnabled)
	downloader := files.NewFilesDownloader(guiEnabled)
	searcher := files.NewSearcher()

	return &Gossiper{
		mode:           mode,
		debug:          debug,
		clientConn:     clientConn,
		gossiperConn:   peerConn,
		rTimerDuration: time.Duration(rTimerDuration) * time.Second,

		Origin:              name,
		GossipAddr:          address,
		ClientAddr:          clientAddr,
		GUIPort:             guiPort,
		FromClientChan:      make(chan *packets_client.ClientPacket), // bufferedChanSize),
		FromGossipChan:      make(chan *GossipChannelElement),        // bufferedChanSize),
		NodeChan:            updatesChannels,
		RumorChan:           updatesChannels,
		PrivateMsgChan:      updatesChannels,
		OriginChan:          updatesChannels,
		PeersSet:            peersSet,
		VectorClock:         vectorClock,
		RoutingTable:        routingTable,
		Conversations:       conversations,
		FilesUploader:       uploader,
		IndexedFilesChan:    uploader.FileChan,
		FilesDownloader:     downloader,
		DownloadedFilesChan: downloader.FileChan,
		FilesSearcher:       searcher,
	}
}

func (g *Gossiper) getOrAddPeer(ipPort string) *peers.Peer {
	peer, _ := g.PeersSet.GetAndError(ipPort) // not nil is like handling error
	if peer != nil {
		return peer
	} else {
		return g.PeersSet.AddIpPort(ipPort)
	}
}

func (g *Gossiper) Start(group *sync.WaitGroup) {
	group.Add(1)
	go g.listenClient(group)

	group.Add(1)
	go g.handleClient(group)

	group.Add(1)
	go g.listenGossip(group)

	group.Add(1)
	go g.handleGossip(group)

	group.Add(1)
	go g.antiEntropy(group)

	group.Add(1)
	go g.routeRumorTicker(group)
}

func (g *Gossiper) listenClient(group *sync.WaitGroup) {
	g.listen(g.clientConn, group, func(buffer []byte, fromIpPort string) {
		var packet packets_client.ClientPacket
		if err := protobuf.Decode(buffer, &packet); err != nil {
			common.HandleAbort("could not decode client packet", err)
			return
		}
		if g.debug {
			logger.Printlnf("<< receiving on client <%s> from <%s>", packet.String(), fromIpPort)
		}
		g.FromClientChan <- &packet
	})
}

func (g *Gossiper) listenGossip(group *sync.WaitGroup) {
	g.listen(g.gossiperConn, group, func(buffer []byte, fromIpPort string) {
		var packet packets_gossiper.GossipPacket
		if err := protobuf.Decode(buffer, &packet); err != nil {
			common.HandleAbort(fmt.Sprintf("could not decode packet of <%s>", fromIpPort), err)
			return
		}
		if err := packet.Check(); err != nil {
			common.HandleAbort(fmt.Sprintf("received incorrect packet from <%s>", fromIpPort), err)
			return
		}
		if g.debug {
			logger.Printlnf("<< receiving on gossiper <%s> from <%s>", packet.String(), fromIpPort)
		}
		g.FromGossipChan <- &GossipChannelElement{
			Packet: &packet,
			From:   g.getOrAddPeer(fromIpPort),
		}
	})
}

func (g *Gossiper) listen(conn *net.UDPConn, group *sync.WaitGroup, callback func([]byte, string)) {
	defer conn.Close()
	defer group.Done()
	wholeBuffer := make([]byte, udpPacketMaxSize)
	for {
		n, udpAddr, err := conn.ReadFromUDP(wholeBuffer)
		if err != nil {
			common.HandleAbort("could not read UDP packet", err)
			continue
		}
		buffer := make([]byte, n)
		copy(buffer, wholeBuffer[:n])
		go callback(buffer, udpAddr.String())
	}
}

func (g *Gossiper) antiEntropy(group *sync.WaitGroup) {
	defer group.Done()
	ticker := time.NewTicker(antiEntropyDuration)
	for range ticker.C {
		if randomPeer := g.PeersSet.GetRandom(); randomPeer != nil {
			g.sendPacket(g.VectorClock.ToStatusPacket(), g.PeersSet.GetRandom())
		}
	}
}

func (g *Gossiper) broadcastRoutePacket() {
	routePacket := g.VectorClock.GetOrCreateHandler(g.Origin).CreateAndSaveNextMessage("")
	g.sendPacket(routePacket, g.PeersSet.GetSlice()...)
}

func (g *Gossiper) routeRumorTicker(group *sync.WaitGroup) {
	defer group.Done()
	if g.rTimerDuration.Seconds() > 0 {
		g.broadcastRoutePacket()
		ticker := time.NewTicker(g.rTimerDuration)
		for range ticker.C {
			g.broadcastRoutePacket()
		}
	}
}

func (g *Gossiper) handleClient(group *sync.WaitGroup) {
	defer group.Done()
	for {
		packet := <-g.FromClientChan

		go func() {
			if !g.debug {
				packet.AckPrint()
			}
			if g.mode.isSimple() && packet.IsPostMessage() { // SIMPLE MODE
				g.handleClientSimpleMode(packet)
			} else { // NORMAL MODE
				g.handleClientNormalMode(packet)
			}
		}()
	}
}

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
			for _, search := range g.FilesSearcher.GetFullSearches() {
				if search.IsFull() {
					// let's download the file now from all origins involved
					for _, requestFile := range search.ToRequestFiles(packet.RequestFile.Filename, packet.RequestFile.Request) {
						g.downloadHandler(requestFile)
					}
				}
			}
		}
	} else if packet.IsIndexFile() {
		g.FilesUploader.IndexFile(packet.IndexFile.Filename, true)
	} else if packet.IsSearchFiles() {
		g.searchHandler(packet.SearchFiles)
	}
}

func (g *Gossiper) downloadHandler(requestFile *packets_client.RequestFilePacket) {
	// not checking if can download because it does not enable downloading from different origins yet
	//canDownload :=
	g.FilesDownloader.AddNewFile(requestFile.Filename, requestFile.Request)
	//if canDownload {
	request := &packets_gossiper.DataRequest{
		Origin:      g.Origin,
		Destination: requestFile.Destination,
		HopLimit:    hopLimit,
		HashValue:   utils.HexToHash(requestFile.Request),
	}
	g.transmit(request, false)
	//}
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
			logger.Printlnf("STOPPING SEARCH")
			searchTicker.Stop()
		} else {

			logger.Printlnf("SEARCHING AGAIN with budget=%d, %s", search.Budget, *search)
			search.DoubleBudget()
			logger.Printlnf("SEARCHING AGAIN with budget=%d", search.Budget)
			g.sendBudgetPacket(&packets_gossiper.SearchRequest{
				Origin:   g.Origin,
				Budget:   search.Budget,
				Keywords: search.Keywords,
			})
		}
	}

}

func (g *Gossiper) sendBudgetPacket(packet packets_gossiper.BudgetPacket, exceptFromPeer ...*peers.Peer) {
	if packet.GetBudget() > 0 {
		toPeers := g.PeersSet.Filter(exceptFromPeer...).GetSlice()
		if len(toPeers) > 0 {
			packets := packet.DividePacket(len(toPeers))
			for i, peer := range toPeers {
				packet := packets[i]
				if packet.GetBudget() > 0 {
					g.sendPacket(packet, peer)
				}
			}
		}
	}
}

func (g *Gossiper) handleGossip(group *sync.WaitGroup) {
	defer group.Done()
	for {
		elem := <-g.FromGossipChan
		packet, fromPeer := elem.Packet, elem.From

		go func() {
			if !g.debug {
				packet.AckPrint(fromPeer, g.Origin)
				g.PeersSet.AckPrint()
			}

			if packet.IsSimple() {
				g.handleSimple(packet.Simple, fromPeer)
			} else if packet.IsRumor() {
				g.handleRumor(packet.Rumor, fromPeer)
			} else if packet.IsStatus() {
				g.handleStatus(packet.Status, fromPeer)
			} else if packet.IsPrivate() {
				g.handlePrivate(packet.Private)
			} else if packet.IsDataRequest() {
				g.handleDataRequest(packet.DataRequest)
			} else if packet.IsDataReply() {
				g.handleDataReply(packet.DataReply)
			} else if packet.IsSearchRequest() {
				g.handleSearchRequest(packet.SearchRequest, fromPeer)
			} else if packet.IsSearchReply() {
				g.handleSearchReply(packet.SearchReply)
			}
		}()
	}
}

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

func (g *Gossiper) transmit(packetToTransmit packets_gossiper.Transmittable, decreaseHop bool) {
	if decreaseHop {
		packetToTransmit = packetToTransmit.Hopped()
	}
	if packetToTransmit.IsTransmittable() && packetToTransmit.Dest() != g.Origin {
		toPeer := g.RoutingTable.GetOrCreateHandler(packetToTransmit.Dest()).GetPeer()
		if toPeer != nil {
			g.sendPacket(packetToTransmit, toPeer)
		}
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
			g.FilesUploader.IndexFile(filename, false)
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
		for _, search := range g.FilesSearcher.GetFullSearches() {
			if search.IsFull() {
				// hw03 print
				logger.Printlnf("SEARCH FINISHED")
			}
		}
	} else {
		g.transmit(packet, true)
	}
}

func (g *Gossiper) sendPacket(packet packets_gossiper.GossipPacketI, to ...*peers.Peer) {

	packetToSend := packet.ToGossipPacket()
	if err := packetToSend.Check(); err != nil {
		common.HandleAbort(fmt.Sprintf("cannot sendPacket incorrect packet"), err)
		return
	}
	if len(to) == 0 {
		common.HandleAbort(fmt.Sprintf("cannot sendPacket to zero peers"), nil)
		return
	}
	packetBytes, err := protobuf.Encode(packet.ToGossipPacket())
	if err != nil {
		common.HandleAbort(fmt.Sprintf("error during packet encoding"), err)
		return
	}

	for _, p := range to {
		go func(p *peers.Peer) {
			if p != nil && !p.Addr.Equals(g.GossipAddr) {
				g.handleSendPacket(packet, p)
				if g.debug {
					logger.Printlnf(">> sending <%s> to <%s>", packet.ToGossipPacket().String(), p.Addr.ToIpPort())
				}
				_, err := g.gossiperConn.WriteToUDP(packetBytes, p.Addr.UDP())
				if err != nil {
					common.HandleAbort("error when sending packet", err)
				}
			} else {
				//common.HandleAbort(fmt.Sprintf("trying to send to peer <%s>", p), nil)
			}
		}(p)
	}
}

func (g *Gossiper) handleSendPacket(packet packets_gossiper.GossipPacketI, toPeer *peers.Peer) {
	packetToSend := packet.ToGossipPacket()
	if packetToSend.IsRumor() {
		packetToSend.Rumor.SendPrintMongering(toPeer)

		// start timeout to this peer if none is already active
		toPeer.FlipTimeout.SetIfNotActive(timeoutDuration, func() {
			if flipped := utils.FlipCoin(); flipped {
				if randomPeer := g.PeersSet.GetRandom(toPeer); randomPeer != nil {
					g.sendPacket(packet, randomPeer)
					packetToSend.Rumor.SendPrintFlipped(randomPeer)
				}
			}
		})
	} else if packetToSend.IsDataRequest() {
		hashString := utils.HashToHex(packetToSend.DataRequest.HashValue)
		g.FilesDownloader.SetTimeout(hashString, packetToSend.DataRequest.Destination, func() {
			g.transmit(packetToSend.DataRequest, false)
		})
	}

}
