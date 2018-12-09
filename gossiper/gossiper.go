package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gregunz/Peerster/blockchain"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/conv"
	"github.com/gregunz/Peerster/models/files"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"github.com/gregunz/Peerster/models/routing"
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

	Origin          string
	GossipAddr      *peers.Address
	ClientAddr      *peers.Address
	GUIPort         uint
	FromClientChan  chan *packets_client.ClientPacket
	FromGossipChan  chan *GossipChannelElement
	PeersSet        *peers.Set
	VectorClock     *vector_clock.VectorClock
	RoutingTable    *routing.Table
	Conversations   *conv.Conversations
	FilesUploader   *files.Uploader
	FilesDownloader *files.Downloader
	FilesSearcher   *files.Searcher
	BlockChainFile  *blockchain.BCF
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

	peersSet.SetNewNodeChan(guiEnabled)
	routingTable := routing.NewTable(name, guiEnabled)
	vectorClock := vector_clock.NewVectorClock(guiEnabled)
	conversations := conv.NewConversations(guiEnabled)
	uploader := files.NewFilesUploader(guiEnabled)
	downloader := files.NewFilesDownloader(guiEnabled)
	searcher := files.NewSearcher(guiEnabled)

	return &Gossiper{
		mode:           mode,
		debug:          debug,
		clientConn:     clientConn,
		gossiperConn:   peerConn,
		rTimerDuration: time.Duration(rTimerDuration) * time.Second,

		Origin:          name,
		GossipAddr:      address,
		ClientAddr:      clientAddr,
		GUIPort:         guiPort,
		FromClientChan:  make(chan *packets_client.ClientPacket), // bufferedChanSize),
		FromGossipChan:  make(chan *GossipChannelElement),        // bufferedChanSize),
		PeersSet:        peersSet,
		VectorClock:     vectorClock,
		RoutingTable:    routingTable,
		Conversations:   conversations,
		FilesUploader:   uploader,
		FilesDownloader: downloader,
		FilesSearcher:   searcher,
		BlockChainFile:  blockchain.NewBCF(),
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

	group.Add(1)
	go g.blockchainRoutine(group)
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

func (g *Gossiper) blockchainRoutine(group *sync.WaitGroup) {
	defer group.Done()

	// mining routine
	group.Add(1)
	go g.BlockChainFile.MiningRoutine(group)

	// handling new mined blocks
	for {
		newBlock := g.BlockChainFile.MineChan.Get()
		if newBlock.IsAfterGenesis() {
			time.Sleep(5 * time.Second)
		}
		g.sendPacket(newBlock.ToBlockPublish(blockHopLimit), g.PeersSet.GetSlice()...)
		//time.Sleep(50 * time.Millisecond)
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

func (g *Gossiper) sendBudgetPacket(packet packets_gossiper.BudgetPacket, exceptFromPeer ...*peers.Peer) {
	logger.Printlnf("SEARCH with budget=%d", packet.GetBudget())
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
			} else if packet.IsTxPublish() {
				g.handleTxPublish(packet.TxPublish, fromPeer)
			} else if packet.IsBlockPublish() {
				g.handleBlockPublish(packet.BlockPublish, fromPeer)
			}
		}()
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
