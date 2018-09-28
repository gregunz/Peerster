package gossiper

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models"
	"net"
)

type Gossiper struct {
	address *net.UDPAddr
	conn    *net.UDPConn
	Name    string
	Peers   *models.Peers
}

func NewGossiper(address, name string) *Gossiper {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	common.HandleError(err)
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	common.HandleError(err)

	peers := models.EmptyPeers()
	return &Gossiper{
		address: udpAddr,
		conn:    udpConn,
		Name:    name,
		Peers:   peers,
	}
}

func (g *Gossiper) AddPeer(peer string) {
	g.Peers.AddPeer(peer)
}

func (g *Gossiper) BroadcastFromClient(message *models.SimpleMessage, senderName string, senderAddr models.Peer) {
	ackClientMessage(message)
	message.OriginalName = senderName
	message.RelayPeerAddr = senderAddr.String()
	g.broadcast(message, g.Peers)
}

func (g *Gossiper) BroadcastFromPeer(message *models.SimpleMessage, senderAddr models.Peer) {
	ackPeerMessage(message)
	relayPeerAddr := message.RelayPeerAddr
	message.RelayPeerAddr = senderAddr.String()
	g.AddPeer(relayPeerAddr)
	g.broadcast(message, g.Peers)
}

func (g *Gossiper) broadcast(message *models.SimpleMessage, to *models.Peers) {
	//packetBytes, err := protobuf.Encode(message)
	//common.HandleError(err)
	/*for _, p := range g.Peers.GetPeersList() {

	}
	net.UDPConn.WriteToUDP(packetBytes, upd_addr)
	*/
}

func ackClientMessage(message *models.SimpleMessage) {
	fmt.Printf("CLIENT MESSAGE %s\n", message.Contents)
}

func ackPeerMessage(message *models.SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
		message.OriginalName,
		message.RelayPeerAddr,
		message.Contents)
}
