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
var dest string
var filename string
var msg string
var request string

func init() {
	flag.UintVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.StringVar(&dest, "dest", "", "destination for the private message or file request")
	flag.StringVar(&filename, "file", "", "filename to be indexed by the gossiper")
	flag.StringVar(&msg, "msg", "", "message to be sent")
	flag.StringVar(&request, "request", "", "request metafile of this hash")
}

func main() {
	flag.Parse()

	packet := inputsToPacket(msg, dest, filename, request)
	if packet == nil {
		common.HandleAbort("combination of arguments are not meaningful, use -help flag for more details", nil)
		return
	}
	// port 0 means that os picks on that is available
	_, udpConn := utils.ConnectToIpPort(fmt.Sprintf("localhost:%d", 0))

	udpAddr := utils.IpPortToUDPAddr(fmt.Sprintf("localhost:%d", uiPort))

	sendMessage(udpAddr, udpConn, packet)
}

func inputsToPacket(msg, dest, filename, request string) packets_client.ClientPacketI {
	var packet packets_client.ClientPacketI = nil

	if msg != "" && dest == "" && filename == "" && request == "" { // sending rumor
		packet = &packets_client.PostMessagePacket{
			Message: msg,
		}
	} else if msg != "" && dest != "" && filename == "" && request == "" { // sending private message
		packet = &packets_client.PostMessagePacket{
			Message:     msg,
			Destination: dest,
		}
	} else if msg == "" && dest != "" && filename != "" && request != "" { // requesting a file
		packet = &packets_client.RequestFilePacket{
			Destination: dest,
			File:        filename,
			Request:     request,
		}
	} else if msg == "" && filename != "" && dest == "" && request == "" {
		packet = &packets_client.IndexFilePacket{
			File: filename,
		}
	}

	return packet
}

func sendMessage(udpAddr *net.UDPAddr, udpConn *net.UDPConn, packet packets_client.ClientPacketI) {
	packetBytes, err := protobuf.Encode(packet.ToClientPacket())
	common.HandleError(err)
	udpConn.WriteToUDP(packetBytes, udpAddr)
}
