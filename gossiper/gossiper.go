package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models"
	"github.com/gregunz/Peerster/utils"
	"net"
	"sync"
	"time"
)

var timeout_duration = 1 * time.Second
var anti_entropy_duration = 1 * time.Second

type Gossiper struct {
	simple        bool
	address       *models.Address
	peerConn      *net.UDPConn
	clientConn    *net.UDPConn
	name          string
	peersSet      *models.PeersSet
	rumorsHandler *models.RumorHandlers
	mux           sync.Mutex
}

func NewGossiper(simple bool, address *models.Address, name string, uiPort uint, peers *models.PeersSet) *Gossiper {

	fmt.Printf("Creating Peerster named <%s> listening peers on ip:port <%s> " +
		"and listening local clients on port <%d> with peers <%s>\n",
		name, address.ToIpPort(), uiPort, peers.ToString("> <"))

	_, peerConn := utils.ConnectToIpPort(address.ToIpPort())
	_, clientConn := utils.ConnectToIpPort(fmt.Sprintf("localhost:%d", uiPort))

	return &Gossiper{
		simple:        simple,
		address:       address,
		peerConn:      peerConn,
		clientConn:    clientConn,
		name:          name,
		peersSet:      peers,
		rumorsHandler: models.NewRumorsHandler(),
	}
}

func (g *Gossiper) getOrAddPeer(ipPort string) *models.Peer {
	peer, _ := g.peersSet.GetAndError(ipPort)
	if peer != nil {
		return peer
	} else {
		peer = models.NewPeer(ipPort)
		g.peersSet.AddPeer(peer)
		return peer
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
		var packet models.ClientPacket
		protobuf.Decode(buffer, &packet)
		go func() {
			g.handleClient(&packet)
		}()
	})
}

func (g *Gossiper) listenPeers(group *sync.WaitGroup) {
	group.Add(1)
	go g.listen(g.peerConn, group, func(buffer []byte, fromIpPort string) {
		var packet models.GossipPacket
		protobuf.Decode(buffer, &packet)
		common.HandleError(packet.Check())
		go func() {
			g.handlePeers(&packet, g.getOrAddPeer(fromIpPort))
		}()
	})
}

func (g *Gossiper) antiEntropy(group *sync.WaitGroup) {
	group.Add(1)
	go func (){
		defer group.Done()
		ticker := time.NewTicker(anti_entropy_duration)
		for range ticker.C {
			go g.broadcast(g.rumorsHandler.ToStatusPacket().ToGossipPacket(), g.peersSet.GetRandom())
		}
	}()
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

func (g *Gossiper) handleSimple(msg *models.SimpleMessage, fromPeer *models.Peer) {
	msgToSend := &models.SimpleMessage{
		Contents: msg.Contents,
		RelayPeerAddr: g.address.ToIpPort(),
		OriginalName: msg.OriginalName,
	}
	toPeers := g.peersSet.Filter(fromPeer).GetSlice() // not resending to sender

	go g.broadcast(msgToSend.ToGossipPacket(), toPeers...)
}



func (g *Gossiper) handleRumor(msg *models.RumorMessage, fromPeer *models.Peer) {

	// saving message
	g.rumorsHandler.Save(msg)

	msgToSend := &models.RumorMessage{
		ID: msg.ID,
		Text: msg.Text,
		Origin: msg.Origin,
	}

	// broadcast to a random peer TODO: ASK IF WE MUST EXCLUDE fromPeer
	if randomPeer:= g.peersSet.GetRandom(fromPeer); randomPeer != nil {
		go g.broadcast(msgToSend.ToGossipPacket(), randomPeer)
		msgToSend.SendPrint(randomPeer, false)
	}

	// send back status packet to sender
	go g.broadcast(g.rumorsHandler.ToStatusPacket().ToGossipPacket(), fromPeer)
}


func (g *Gossiper) handleStatus(packet *models.StatusPacket, fromPeer *models.Peer) {
	fromPeer.StopTimeout()
	rumorMsg, remoteHasMsg := g.rumorsHandler.Compare(packet.Want)

	if rumorMsg != nil { // has a msg to send
		go g.broadcast(rumorMsg.ToGossipPacket(), fromPeer) // send the new message
		rumorMsg.SendPrint(fromPeer, false)
	} else if remoteHasMsg { // remote has new message //TODO: check if both cannot happen (else if)
		go g.broadcast(g.rumorsHandler.ToStatusPacket().ToGossipPacket(), fromPeer) // send status to remote
	} else  { // is up to date
		fromPeer.TriggerTimeout()
		fmt.Printf("IN SYNC WITH %s\n", fromPeer.Addr.ToIpPort())
	}
	// prints
	packet.AckPrint(fromPeer)
}

func (g *Gossiper) handleClient(packet *models.ClientPacket) {
	packet.AckPrint()
	if g.simple {
		msg := &models.SimpleMessage{
			Contents: packet.Message,
			RelayPeerAddr: g.address.ToIpPort(),
			OriginalName: g.name,
		}
		g.broadcast(msg.ToGossipPacket(), g.peersSet.GetSlice()...)
	} else {
		g.mux.Lock()
		defer g.mux.Unlock()

		meHandler := g.rumorsHandler.GetOrCreateHandler(g.name)
		rumor := meHandler.NextMessage(packet.Message)

		randomPeer := g.peersSet.GetRandom()
		g.broadcast(rumor.ToGossipPacket(), randomPeer)
	}
}

func (g *Gossiper) handlePeers(packet *models.GossipPacket, fromPeer *models.Peer) {

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

func (g *Gossiper) broadcast(packet *models.GossipPacket, to ...*models.Peer) {
	common.HandleError(packet.Check())
	if len(to) == 0 {
		common.HandleError(fmt.Errorf("cannot broadcast to zero peers"))
	}
	packetBytes, err := protobuf.Encode(packet)
	common.HandleError(err)

	for _, p := range to {
		go func(p *models.Peer) {
			if p != nil && !p.Addr.Equals(g.address) {
				if packet.IsRumor() {
					p.SetTimeout(timeout_duration, func() {
						if flipped := utils.FlipCoin(); flipped {
							randomPeer := g.peersSet.GetRandom()
							g.broadcast(packet, randomPeer)
							packet.Rumor.SendPrint(randomPeer, flipped)
						}
					})
				}
				g.peerConn.WriteToUDP(packetBytes, p.Addr.UDP())
			} else {
				if p == nil {
					common.HandleError(fmt.Errorf("tried to broadcast to peer '%s'", p))
				}
			}
		}(p)
	}
}

/*
func (g *Gossiper) SaveRumor(msg *models.RumorMessage) {
	g.rumorsHandler.Save(msg)
}
*/