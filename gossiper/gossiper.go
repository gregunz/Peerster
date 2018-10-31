package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/conv"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"github.com/gregunz/Peerster/models/routing"
	"github.com/gregunz/Peerster/models/updates"
	"github.com/gregunz/Peerster/models/vector_clock"
	"github.com/gregunz/Peerster/utils"
	"log"
	"net"
	"sync"
	"time"
)

const (
	timeoutDuration     = 1 * time.Second
	antiEntropyDuration = 1 * time.Second
	udpPacketSize       = 4096
	hopLimit            = 10
)

type Gossiper struct {
	mode           *Mode
	clientConn     *net.UDPConn
	gossiperConn   *net.UDPConn
	rTimerDuration time.Duration

	Origin         string
	GossipAddr     *peers.Address
	ClientAddr     *peers.Address
	GUIPort        uint
	FromClientChan chan *packets_client.PostMessagePacket
	FromGossipChan chan *GossipChannelElement
	NodeChan       peers.NodeChan
	RumorChan      vector_clock.RumorChan
	OriginChan     routing.OriginChan
	PrivateMsgChan conv.PrivateMsgChan
	PeersSet       *peers.Set
	VectorClock    *vector_clock.VectorClock
	RoutingTable   *routing.Table
	Conversations  *conv.Conversations

	mux sync.Mutex
}

func NewGossiper(simple bool, address *peers.Address, name string, uiPort uint, guiPort uint, peersSet *peers.Set,
	rTimerDuration uint) *Gossiper {

	log.Printf("Gossiper created: named <%s> listening peers on ip:port <%s> "+
		"and listening local clients on port <%d> with peers <%s>\n",
		name, address.ToIpPort(), uiPort, peersSet.ToString("> <"))

	mode := NewDefaultMode()
	if simple {
		mode = NewSimpleMode()
	}

	clientAddr := peers.NewAddress(fmt.Sprintf("localhost:%d", uiPort))
	_, peerConn := utils.ConnectToIpPort(address.ToIpPort())
	_, clientConn := utils.ConnectToIpPort(clientAddr.ToIpPort())

	updatesChannels := updates.NewChannels()
	routingTable := routing.NewTable(name, updatesChannels)
	vectorClock := vector_clock.NewVectorClock(updatesChannels)
	conversations := conv.NewConversations(updatesChannels)
	peersSet.SetNodeChan(updatesChannels)

	return &Gossiper{
		mode:           mode,
		ClientAddr:     clientAddr,
		clientConn:     clientConn,
		gossiperConn:   peerConn,
		rTimerDuration: time.Duration(rTimerDuration) * time.Second,

		Origin:         name,
		GossipAddr:     address,
		GUIPort:        guiPort,
		FromClientChan: make(chan *packets_client.PostMessagePacket),
		FromGossipChan: make(chan *GossipChannelElement),
		NodeChan:       updatesChannels,
		RumorChan:      updatesChannels,
		PrivateMsgChan: updatesChannels,
		OriginChan:     updatesChannels,
		PeersSet:       peersSet,
		VectorClock:    vectorClock,
		RoutingTable:   routingTable,
		Conversations:  conversations,
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
	g.listen(g.clientConn, group, func(buffer []byte, _ string) {
		var packet packets_client.PostMessagePacket
		if err := protobuf.Decode(buffer, &packet); err != nil {
			common.HandleAbort("could not decode client packet", err)
			return
		}
		if packet.Message != "" {
			g.FromClientChan <- &packet
		}
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
		g.FromGossipChan <- &GossipChannelElement{
			Packet: &packet,
			From:   g.getOrAddPeer(fromIpPort),
		}
	})
}

func (g *Gossiper) listen(conn *net.UDPConn, group *sync.WaitGroup, callback func([]byte, string)) {
	defer conn.Close()
	defer group.Done()
	wholeBuffer := make([]byte, udpPacketSize)
	for {
		n, udpAddr, err := conn.ReadFromUDP(wholeBuffer)
		common.HandleError(err)
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
			go g.sendPacket(g.VectorClock.ToStatusPacket().ToGossipPacket(), g.PeersSet.GetRandom())
		}
	}
}

func (g *Gossiper) broadcastRoutePacket() {
	routePacket := g.VectorClock.GetOrCreateHandler(g.Origin).CreateAndSaveNextMessage("")
	go g.sendPacket(routePacket.ToGossipPacket(), g.PeersSet.GetSlice()...)
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
			packet.AckPrint()
			if g.mode.isSimple() {
				msg := &packets_gossiper.SimpleMessage{
					Contents:      packet.Message,
					RelayPeerAddr: g.GossipAddr.ToIpPort(),
					OriginalName:  g.Origin,
				}
				go g.sendPacket(msg.ToGossipPacket(), g.PeersSet.GetSlice()...)
			} else if packet.Destination == "" {
				meHandler := g.VectorClock.GetOrCreateHandler(g.Origin)
				rumorMessage := meHandler.CreateAndSaveNextMessage(packet.Message)

				if randomPeer := g.PeersSet.GetRandom(); randomPeer != nil {
					go g.sendPacket(rumorMessage.ToGossipPacket(), randomPeer)
				}
			} else {
				meHandler := g.Conversations.GetOrCreateHandler(g.Origin)
				msg := meHandler.CreateAndSaveNextMessage(packet.Message, packet.Destination, hopLimit)
				toPeer := g.RoutingTable.GetOrCreateHandler(msg.Destination).GetPeer()
				if msg.Destination != g.Origin && toPeer != nil {
					go g.sendPacket(msg.ToGossipPacket(), toPeer)
				}
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

	go g.sendPacket(msgToSend.ToGossipPacket(), toPeers...)
}

func (g *Gossiper) handleRumor(msg *packets_gossiper.RumorMessage, fromPeer *peers.Peer) {

	// saving message
	g.VectorClock.GetOrCreateHandler(msg.Origin).Save(msg)
	g.RoutingTable.GetOrCreateHandler(msg.Origin).AckRumor(msg, fromPeer)

	msgToSend := &packets_gossiper.RumorMessage{
		ID:     msg.ID,
		Text:   msg.Text,
		Origin: msg.Origin,
	}

	// sendPacket to a random peer TODO: ASK IF WE MUST EXCLUDE fromPeer
	if randomPeer := g.PeersSet.GetRandom(fromPeer); randomPeer != nil {
		go g.sendPacket(msgToSend.ToGossipPacket(), randomPeer)
	}

	// send back status packet to sender (= ack of the rumor)
	go g.sendPacket(g.VectorClock.ToStatusPacket().ToGossipPacket(), fromPeer)
}

func (g *Gossiper) handleStatus(packet *packets_gossiper.StatusPacket, fromPeer *peers.Peer) {
	rumorMsg, remoteHasMsg := g.VectorClock.Compare(packet.ToMap())

	if rumorMsg != nil { // has a msg to send
		go g.sendPacket(rumorMsg.ToGossipPacket(), fromPeer) // send the new message
	}
	if remoteHasMsg { // remote has new message
		go g.sendPacket(g.VectorClock.ToStatusPacket().ToGossipPacket(), fromPeer) // send status to remote
	}
	if rumorMsg == nil && !remoteHasMsg { // is up to date
		fromPeer.Timeout.Trigger()
		fmt.Printf("IN SYNC WITH %s\n", fromPeer.Addr.ToIpPort())
	} else {
		fromPeer.Timeout.Cancel()
	}
}

func (g *Gossiper) handlePrivate(msg *packets_gossiper.PrivateMessage) {
	if msg.Destination != g.Origin {
		msgToSend := msg.Hopped()
		toPeer := g.RoutingTable.GetOrCreateHandler(msg.Destination).GetPeer()
		if msgToSend.HopLimit > 0 && toPeer != nil {
			go g.sendPacket(msgToSend.ToGossipPacket(), toPeer)
		}
	} else {
		g.Conversations.GetOrCreateHandler(msg.Origin).Save(msg)
	}
}

func (g *Gossiper) handleGossip(group *sync.WaitGroup) {
	defer group.Done()
	for {
		elem := <-g.FromGossipChan
		packet, fromPeer := elem.Packet, elem.From

		go func() {
			packet.AckPrint(fromPeer, g.Origin)
			g.PeersSet.AckPrint()

			if packet.IsSimple() {
				g.handleSimple(packet.Simple, fromPeer)
			}
			if packet.IsRumor() {
				g.handleRumor(packet.Rumor, fromPeer)
			}
			if packet.IsStatus() {
				g.handleStatus(packet.Status, fromPeer)
			}
			if packet.IsPrivate() {
				g.handlePrivate(packet.Private)
			}
		}()
	}
}

func (g *Gossiper) sendPacket(packet *packets_gossiper.GossipPacket, to ...*peers.Peer) {

	if err := packet.Check(); err != nil {
		common.HandleAbort(fmt.Sprintf("cannot sendPacket incorrect packet"), err)
		return
	}
	if len(to) == 0 {
		common.HandleAbort(fmt.Sprintf("cannot sendPacket to zero peers"), nil)
		return
	}
	packetBytes, err := protobuf.Encode(packet)
	if err != nil {
		common.HandleAbort(fmt.Sprintf("error during packet encoding"), err)
		return
	}

	for _, p := range to {
		go func(p *peers.Peer) {
			if p != nil && !p.Addr.Equals(g.GossipAddr) {
				g.handleSendPacket(packet, p)
				g.gossiperConn.WriteToUDP(packetBytes, p.Addr.UDP())
			} else {
				//common.HandleAbort(fmt.Sprintf("trying to send to peer <%s>", p), nil)
			}
		}(p)
	}
}

func (g *Gossiper) handleSendPacket(packet *packets_gossiper.GossipPacket, toPeer *peers.Peer) {

	if packet.IsRumor() {
		packet.Rumor.SendPrintMongering(toPeer)

		// start timeout to this peer if none is already active
		toPeer.Timeout.SetIfNotActive(timeoutDuration, func() {
			if flipped := utils.FlipCoin(); flipped {
				if randomPeer := g.PeersSet.GetRandom(toPeer); randomPeer != nil {
					go g.sendPacket(packet, randomPeer)
					packet.Rumor.SendPrintFlipped(randomPeer)
				}
			}
		})
	}

}
