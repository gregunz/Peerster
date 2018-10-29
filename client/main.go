package main

import (
	"flag"
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/utils"
	"net"
)

var uiPort uint
var msg string

func init() {
	flag.UintVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.StringVar(&msg, "msg", "", "message to be sent")
}

func main() {
	flag.Parse()

	packet := packets_client.PostMessagePacket{
		Message: msg,
	}

	// port 0 means that os picks on that is available
	_, udpConn := utils.ConnectToIpPort(fmt.Sprintf("localhost:%d", 0))

	udpAddr := utils.IpPortToUDPAddr(fmt.Sprintf("localhost:%d", uiPort))

	sendMessage(udpAddr, udpConn, &packet)
}

func sendMessage(udpAddr *net.UDPAddr, udpConn *net.UDPConn, packet *packets_client.PostMessagePacket) {
	packetBytes, err := protobuf.Encode(packet)
	common.HandleError(err)
	udpConn.WriteToUDP(packetBytes, udpAddr)
}
