package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models"
	"github.com/gregunz/Peerster/utils"
	"net"
	"sync"
)

type Gossiper struct {
	address    *models.Address
	peerConn   *net.UDPConn
	clientConn *net.UDPConn
	name       string
	peers      *models.PeersSet

}

func NewGossiper(address *models.Address, name string, uiPort uint, peers *models.PeersSet) *Gossiper {

	fmt.Printf("Creating Peerster named <%s> listening peers on ip:port <%s> " +
		"and listening local clients on port <%d> with peers <%s>\n",
		name, address.ToIpPort(), uiPort, peers.ToString("> <"))

	_, peerConn := utils.ConnectToIpPort(address.ToIpPort())
	_, clientConn := utils.ConnectToIpPort(fmt.Sprintf("localhost:%d", uiPort))

	return &Gossiper{
		address:    address,
		peerConn:   peerConn,
		clientConn: clientConn,
		name:       name,
		peers:      peers,
	}
}

func (g *Gossiper) AddPeer(peer *models.Peer) {
	//TODO: Handle concurrency here
	if !peer.Addr.Equals(g.address) {
		g.peers.AddPeer(peer)
	}
}

func (g *Gossiper) Start() {
	var group sync.WaitGroup

	g.listenClient(&group)
	g.listenPeers(&group)

	fmt.Println("Ready!")
	group.Wait()
}

func (g *Gossiper) listenClient(group *sync.WaitGroup) {
	group.Add(1)
	go g.listen(g.clientConn, group, func(packet *models.GossipPacket, p *models.Peer) {
		g.handle(packet, p, true)
	})
}

func (g *Gossiper) listenPeers(group *sync.WaitGroup) {
	group.Add(1)
	go g.listen(g.peerConn, group, func(packet *models.GossipPacket, p *models.Peer) {
		g.handle(packet, p, false)
	})
}

func (g *Gossiper) listen(conn *net.UDPConn, group *sync.WaitGroup, callback func(*models.GossipPacket, *models.Peer)) {
	defer conn.Close()
	defer group.Done()
	buffer := make([]byte, 4096)
	for {
		n, udpAddr, err := conn.ReadFromUDP(buffer)
		common.HandleError(err)
		var packet models.GossipPacket
		protobuf.Decode(buffer[:n], &packet)
		common.HandleError(packet.Check())
		callback(&packet, models.NewPeer(udpAddr.String()))
		g.peers.AckPrint()
	}
}

func (g *Gossiper) handleSimple(msg *models.SimpleMessage, fromPeer *models.Peer, fromClient bool) {
	var msgToSend models.SimpleMessage
	var toPeers []*models.Peer

	msgToSend.Contents = msg.Contents
	msgToSend.RelayPeerAddr = g.address.ToIpPort()

	if fromClient {
		msgToSend.OriginalName = g.name
		toPeers = g.peers.GetSlice()
	} else {
		msgToSend.OriginalName = msg.OriginalName
		toPeers = g.peers.Filter(fromPeer).GetSlice() // not resending to sender
	}
	// prints
	msg.AckPrint(fromClient)
	// broadcast to all peers except sender if from client
	g.broadcast(msgToSend.ToGossipPacket(), toPeers...)
}

func (g *Gossiper) handleRumor(msg *models.RumorMessage, fromPeer *models.Peer, fromClient bool) {

	g.peers.SaveRumor(msg, fromPeer)

	var msgToSend models.RumorMessage
	msgToSend.ID = msg.ID
	msgToSend.Text = msg.Text
	if fromClient {
		msgToSend.Origin = g.name
	} else {
		msgToSend.Origin = msg.Origin
	}
	// prints
	msg.AckPrint(fromPeer)
	// broadcast to a random peer
	// TODO: SET TIMER
	// timer := time.NewTicker(1 * time.Second)​
	randomPeer := g.peers.GetRandom()
	g.broadcast(msgToSend.ToGossipPacket(), randomPeer)
	msg.SendPrint(randomPeer, false)
	// send back status packet to sender
	g.broadcast(g.peers.ToStatusPacket().ToGossipPacket(), fromPeer)
}

func (g *Gossiper) handleStatus(packet *models.StatusPacket, fromPeer *models.Peer, fromClient bool) {

	// prints
	packet.AckPrint(fromPeer)
}

func (g *Gossiper) handle(packet *models.GossipPacket, fromPeer *models.Peer, fromClient bool) {

	if !fromClient {
		g.AddPeer(fromPeer)
	}

	if packet.Simple != nil {
		g.handleSimple(packet.Simple, fromPeer, fromClient)
	}
	if packet.Rumor != nil {
		g.handleRumor(packet.Rumor, fromPeer, fromClient)
	}
	if packet.Status != nil {
		g.handleStatus(packet.Status, fromPeer, fromClient)
	}

}

func (g *Gossiper) broadcast(packet *models.GossipPacket, to ...*models.Peer) {
	common.HandleError(packet.Check())
	packetBytes, err := protobuf.Encode(packet)
	common.HandleError(err)

	for _, p := range to {
		// TODO: check if go routine is necessary here
		go g.peerConn.WriteToUDP(packetBytes, p.Addr.UDP())
	}
}
