package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/clock"
	"github.com/gregunz/Peerster/models/packets"
	"github.com/gregunz/Peerster/models/peers"
	"github.com/gregunz/Peerster/utils"
	"net"
	"sync"
	"time"
)

var timeout_duration = 1 * time.Second
var anti_entropy_duration = 1 * time.Second

type Gossiper struct {
	simple      bool
	address     *peers.Address
	peerConn    *net.UDPConn
	clientConn  *net.UDPConn
	name        string
	peersSet    *peers.PeersSet
	vectorClock *clock.VectorClock
	mux         sync.Mutex
}

func NewGossiper(simple bool, address *peers.Address, name string, uiPort uint, peers *peers.PeersSet) *Gossiper {

	fmt.Printf("Creating Peerster named <%s> listening peers on ip:port <%s> "+
		"and listening local clients on port <%d> with peers <%s>\n",
		name, address.ToIpPort(), uiPort, peers.ToString("> <"))

	_, peerConn := utils.ConnectToIpPort(address.ToIpPort())
	_, clientConn := utils.ConnectToIpPort(fmt.Sprintf("localhost:%d", uiPort))

	//if peerConn == nil || clientConn == nil {
	//	common.HandleAbort("could not connect to ")
	//}

	return &Gossiper{
		simple:      simple,
		address:     address,
		peerConn:    peerConn,
		clientConn:  clientConn,
		name:        name,
		peersSet:    peers,
		vectorClock: clock.NewVectorClock(name),
	}
}

func (g *Gossiper) getOrAddPeer(ipPort string) *peers.Peer {
	peer, _ := g.peersSet.GetAndError(ipPort) // not nil is like handling error
	if peer != nil {
		return peer
	} else {
		return g.peersSet.AddIpPort(ipPort)
	}
}

func (g *Gossiper) Start() {
	var group sync.WaitGroup

	g.listenClient(&group)
	g.listenPeers(&group)
	g.antiEntropy(&group)

	fmt.Println("Ready!")
	group.Wait()
}

func (g *Gossiper) listenClient(group *sync.WaitGroup) {
	group.Add(1)
	go g.listen(g.clientConn, group, func(buffer []byte, _ string) {
		var packet packets.ClientPacket
		protobuf.Decode(buffer, &packet)
		go func(packet *packets.ClientPacket) {
			g.handleClient(packet)
		}(&packet)
	})
}

func (g *Gossiper) listenPeers(group *sync.WaitGroup) {
	group.Add(1)
	go g.listen(g.peerConn, group, func(buffer []byte, fromIpPort string) {
		var packet packets.GossipPacket
		protobuf.Decode(buffer, &packet)
		if err := packet.Check(); err != nil {
			common.HandleAbort(fmt.Sprintf("received incorrect packet from <%s>", fromIpPort), err)
			return
		}
		go func(packet *packets.GossipPacket, peer *peers.Peer) {
			g.handlePeers(packet, peer)
		}(&packet, g.getOrAddPeer(fromIpPort))
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
	group.Add(1)
	go func() {
		defer group.Done()
		ticker := time.NewTicker(anti_entropy_duration)
		for range ticker.C {
			go g.sendPacket(g.vectorClock.ToStatusPacket().ToGossipPacket(), g.peersSet.GetRandom())
		}
	}()
}

func (g *Gossiper) handleSimple(msg *packets.SimpleMessage, fromPeer *peers.Peer) {
	msgToSend := &packets.SimpleMessage{
		Contents:      msg.Contents,
		RelayPeerAddr: g.address.ToIpPort(),
		OriginalName:  msg.OriginalName,
	}
	toPeers := g.peersSet.Filter(fromPeer).GetSlice() // not resending to sender

	go g.sendPacket(msgToSend.ToGossipPacket(), toPeers...)
}

func (g *Gossiper) handleRumor(msg *packets.RumorMessage, fromPeer *peers.Peer) {

	// saving message
	g.vectorClock.Save(msg)

	msgToSend := &packets.RumorMessage{
		ID:     msg.ID,
		Text:   msg.Text,
		Origin: msg.Origin,
	}

	// sendPacket to a random peer TODO: ASK IF WE MUST EXCLUDE fromPeer
	if randomPeer := g.peersSet.GetRandom(fromPeer); randomPeer != nil {
		go g.sendPacket(msgToSend.ToGossipPacket(), randomPeer)
	}

	// send back status packet to sender (= ack of the rumor)
	go g.sendPacket(g.vectorClock.ToStatusPacket().ToGossipPacket(), fromPeer)
}

func (g *Gossiper) handleStatus(packet *packets.StatusPacket, fromPeer *peers.Peer) {
	rumorMsg, remoteHasMsg := g.vectorClock.Compare(packet.ToMap())

	if rumorMsg != nil { // has a msg to send
		go g.sendPacket(rumorMsg.ToGossipPacket(), fromPeer) // send the new message
	}
	if remoteHasMsg { // remote has new message //TODO: check if both cannot happen (else if)
		go g.sendPacket(g.vectorClock.ToStatusPacket().ToGossipPacket(), fromPeer) // send status to remote
	}
	if rumorMsg == nil && !remoteHasMsg { // is up to date
		fromPeer.Timeout.Trigger()
		fmt.Printf("IN SYNC WITH %s\n", fromPeer.Addr.ToIpPort())
	} else {
		fromPeer.Timeout.Cancel()
	}
}

func (g *Gossiper) handleClient(packet *packets.ClientPacket) {
	packet.AckPrint()
	if g.simple {
		msg := &packets.SimpleMessage{
			Contents:      packet.Message,
			RelayPeerAddr: g.address.ToIpPort(),
			OriginalName:  g.name,
		}
		g.sendPacket(msg.ToGossipPacket(), g.peersSet.GetSlice()...)
	} else {
		meHandler := g.vectorClock.GetOrCreateHandler(g.name)
		rumorMessage := meHandler.NextMessage(packet.Message)

		if randomPeer := g.peersSet.GetRandom(); randomPeer != nil {
			g.sendPacket(rumorMessage.ToGossipPacket(), randomPeer)
		}

	}
}

func (g *Gossiper) handlePeers(packet *packets.GossipPacket, fromPeer *peers.Peer) {

	packet.AckPrint(fromPeer)
	g.peersSet.AckPrint()

	if packet.Simple != nil {
		g.handleSimple(packet.Simple, fromPeer)
	}
	if packet.Rumor != nil {
		g.handleRumor(packet.Rumor, fromPeer)
	}
	if packet.Status != nil {
		g.handleStatus(packet.Status, fromPeer)
	}

}

func (g *Gossiper) sendPacket(packet *packets.GossipPacket, to ...*peers.Peer) {
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
			if p != nil && !p.Addr.Equals(g.address) {
				g.handleSendPacket(packet, p)
				g.peerConn.WriteToUDP(packetBytes, p.Addr.UDP())
			} else {
				common.HandleAbort(fmt.Sprintf("trying to send to peer <%s>", p), nil)
			}
		}(p)
	}
}

func (g *Gossiper) handleSendPacket(packet *packets.GossipPacket, toPeer *peers.Peer) {

	if packet.IsRumor() {
		packet.Rumor.SendPrintMongering(toPeer)

		// start timeout to this peer
		toPeer.Timeout.Set(timeout_duration, func() {
			if flipped := utils.FlipCoin(); flipped {
				if randomPeer := g.peersSet.GetRandom(toPeer); randomPeer != nil {
					g.sendPacket(packet, randomPeer)
					packet.Rumor.SendPrintFlipped(randomPeer)
				}
			}
		})

	}

}
