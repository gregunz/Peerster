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
	Name       string
	Peers      *models.PeersSet
}

func NewGossiper(address *models.Address, name string, uiPort uint, peers *models.PeersSet) *Gossiper {

	_, peerConn := utils.ConnectToIpPort(address.ToIpPort())
	_, clientConn := utils.ConnectToIpPort(fmt.Sprintf("localhost:%d", uiPort))

	return &Gossiper{
		address:    address,
		peerConn:   peerConn,
		clientConn: clientConn,
		Name:       name,
		Peers:      peers,
	}
}

func (g *Gossiper) AddPeer(peer models.Peer) {
	g.Peers.AddPeer(peer)
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
	go g.listen(g.clientConn, group, func(packetBytes []byte, addr *net.UDPAddr) {
		var packet models.GossipPacket
		protobuf.Decode(packetBytes, &packet)
		g.broadcast(&packet, true)
	})
}

func (g *Gossiper) listenPeers(group *sync.WaitGroup) {
	group.Add(1)
	go g.listen(g.peerConn, group, func(packetBytes []byte, addr *net.UDPAddr) {
		var packet models.GossipPacket
		protobuf.Decode(packetBytes, &packet)
		g.broadcast(&packet, false)
	})
}

func (g *Gossiper) listen(conn *net.UDPConn, group *sync.WaitGroup, callback func([]byte, *net.UDPAddr)) {
	defer conn.Close()
	defer group.Done()
	buffer := make([]byte, 65535)
	for {
		n, udpAddr, err := conn.ReadFromUDP(buffer)
		common.HandleError(err)
		callback(buffer[:n], udpAddr)
	}
}

func (g Gossiper) broadcast(packet *models.GossipPacket, fromClient bool) {
	common.HandleError(packet.Check())

	toPeers := g.Peers

	if packet.Simple != nil {
		msg := packet.Simple
		msg.AckPrint(fromClient)

		if fromClient {
			msg.OriginalName = g.Name
			msg.RelayPeerAddr = g.address.ToIpPort()
		} else {
			relayPeerAddr := msg.RelayPeerAddr
			newPeer := models.NewPeer(relayPeerAddr)
			g.AddPeer(newPeer)
			toPeers = g.Peers.Filter(newPeer) // not resending to sender
		}
		packet.Simple = msg
	}

	if packet.Rumor != nil {
		// handle rumor message
	}

	if packet.Status != nil {
		// handle status packet

	}

	common.HandleError(packet.Check())
	packetBytes, err := protobuf.Encode(packet)
	common.HandleError(err)

	for _, p := range toPeers.GetSlice() {
		// TODO: check if go routine is necessary here
		go g.peerConn.WriteToUDP(packetBytes, p.Addr.UDPAddr)
	}

}
