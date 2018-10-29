package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/origin_handlers"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"github.com/gregunz/Peerster/utils"
	"log"
	"net"
	"sync"
	"time"
)

var timeout_duration = 1 * time.Second
var anti_entropy_duration = 1 * time.Second

type Gossiper struct {
	mode          *GossiperMode
	clientAddress *peers.Address
	clientConn    *net.UDPConn
	gossiperConn  *net.UDPConn
	Name          string
	Addr          *peers.Address
	GUIPort       uint
	ClientChan    chan *packets_client.PostMessagePacket
	GossipChan    chan *GossipChannelElement
	PeersSet      *peers.PeersSet
	VectorClock   origin_handlers.VectorClock
	RoutingTable  origin_handlers.RoutingTable

	mux sync.Mutex
}

func NewGossiper(simple bool, address *peers.Address, name string, uiPort uint, guiPort uint, peersSet *peers.PeersSet) *Gossiper {

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

	originToHandlers := origin_handlers.NewOriginToHandlers(name)

	return &Gossiper{
		mode:          mode,
		Name:          name,
		GUIPort:       guiPort,
		ClientChan:    make(chan *packets_client.PostMessagePacket, 1),
		GossipChan:    make(chan *GossipChannelElement, 1),
		Addr:          address,
		clientAddress: clientAddr,
		clientConn:    clientConn,
		gossiperConn:  peerConn,
		PeersSet:      peersSet,
		VectorClock:   originToHandlers.ToVectorClock(),
		RoutingTable:  originToHandlers.ToRoutingTable(),
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
}

func (g *Gossiper) listenClient(group *sync.WaitGroup) {
	g.listen(g.clientConn, group, func(buffer []byte, _ string) {
		var packet packets_client.PostMessagePacket
		protobuf.Decode(buffer, &packet)
		g.ClientChan <- &packet
	})
}

func (g *Gossiper) listenGossip(group *sync.WaitGroup) {
	g.listen(g.gossiperConn, group, func(buffer []byte, fromIpPort string) {
		var packet packets_gossiper.GossipPacket
		if err := protobuf.Decode(buffer, &packet); err != nil {
			common.HandleAbort(fmt.Sprintf("received incorrect packet from <%s>", fromIpPort), err)
			return
		}
		if err := packet.Check(); err != nil {
			common.HandleAbort(fmt.Sprintf("received incorrect packet from <%s>", fromIpPort), err)
			return
		}
		g.GossipChan <- &GossipChannelElement{
			Packet: &packet,
			From:   g.getOrAddPeer(fromIpPort),
		}
	})
}

func (g *Gossiper) listen(conn *net.UDPConn, group *sync.WaitGroup, callback func([]byte, string)) {
	defer conn.Close()
	defer group.Done()
	buffer := make([]byte, 4096)
	for {
		n, udpAddr, err := conn.ReadFromUDP(buffer)
		common.HandleError(err)
		callback(buffer[:n], udpAddr.String())
	}
}

func (g *Gossiper) antiEntropy(group *sync.WaitGroup) {
	defer group.Done()
	ticker := time.NewTicker(anti_entropy_duration)
	for range ticker.C {
		if randomPeer := g.PeersSet.GetRandom(); randomPeer != nil {
			go g.sendPacket(g.VectorClock.ToStatusPacket().ToGossipPacket(), g.PeersSet.GetRandom())
		}
	}
}

func (g *Gossiper) handleSimple(msg *packets_gossiper.SimpleMessage, fromPeer *peers.Peer) {
	msgToSend := &packets_gossiper.SimpleMessage{
		Contents:      msg.Contents,
		RelayPeerAddr: g.Addr.ToIpPort(),
		OriginalName:  msg.OriginalName,
	}
	toPeers := g.PeersSet.Filter(fromPeer).GetSlice() // not resending to sender

	go g.sendPacket(msgToSend.ToGossipPacket(), toPeers...)
}

func (g *Gossiper) handleRumor(msg *packets_gossiper.RumorMessage, fromPeer *peers.Peer) {

	// saving message
	g.VectorClock.Save(msg)
	g.RoutingTable.AckRumor(msg, fromPeer)

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

func (g *Gossiper) handleClient(group *sync.WaitGroup) {
	defer group.Done()
	for {
		packet := <-g.ClientChan

		go func() {
			packet.AckPrint()
			if g.mode.isSimple() {
				msg := &packets_gossiper.SimpleMessage{
					Contents:      packet.Message,
					RelayPeerAddr: g.Addr.ToIpPort(),
					OriginalName:  g.Name,
				}
				g.sendPacket(msg.ToGossipPacket(), g.PeersSet.GetSlice()...)
			} else {
				meHandler := g.VectorClock.GetOrCreateHandler(g.Name)
				rumorMessage := meHandler.CreateNextMessage(packet.Message)
				g.VectorClock.Save(rumorMessage)

				if randomPeer := g.PeersSet.GetRandom(); randomPeer != nil {
					g.sendPacket(rumorMessage.ToGossipPacket(), randomPeer)
				}

			}
		}()
	}
}

func (g *Gossiper) handleGossip(group *sync.WaitGroup) {
	defer group.Done()
	for {
		elem := <-g.GossipChan
		packet, fromPeer := elem.Packet, elem.From

		go func() {
			packet.AckPrint(fromPeer)
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
			if p != nil && !p.Addr.Equals(g.Addr) {
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

		// start timeout to this peer
		toPeer.Timeout.SetIfNotActive(timeout_duration, func() {
			if flipped := utils.FlipCoin(); flipped {
				if randomPeer := g.PeersSet.GetRandom(toPeer); randomPeer != nil {
					g.sendPacket(packet, randomPeer)
					packet.Rumor.SendPrintFlipped(randomPeer)
				}
			}
		})

	}

}
