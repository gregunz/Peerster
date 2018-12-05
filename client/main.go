package main

import (
	"flag"
	"fmt"
	"github.com/dedis/protobuf"
	"github.com/gregunz/Peerster/client/flag_var"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/utils"
	"net"
)

const (
	defaultUIPort = 8080
	defaultBudget = 2
)

var uiPort uint
var destination string
var filename string
var message string
var request string
var keywords flag_var.StringListVar
var budget uint64

func init() {
	flag.UintVar(&uiPort, "UIPort", defaultUIPort, "port for the UI client")
	flag.StringVar(&destination, "dest", "", "destination for the private message or file request")
	flag.StringVar(&filename, "file", "", "filename to be indexed by the gossiper")
	flag.StringVar(&message, "msg", "", "message to be sent")
	flag.StringVar(&request, "request", "", "request metafile of this hash")
	flag.Var(&keywords, "keywords", "start a file search with those comma-separated keywords")
	flag.Uint64Var(&budget, "budget", defaultBudget, "budget for the file search")
}

func main() {
	flag.Parse()

	packet := inputsToPacket(
		message != "",
		destination != "",
		filename != "",
		request != "",
		len(keywords.List) > 0,
		budget != defaultBudget)

	if packet == nil {
		common.HandleAbort("combination of arguments are not meaningful, use -help flag for more details", nil)
		return
	}
	// port 0 means that os picks on that is available
	_, udpConn := utils.ConnectToIpPort(fmt.Sprintf("localhost:%d", 0))

	udpAddr := utils.IpPortToUDPAddr(fmt.Sprintf("localhost:%d", uiPort))

	sendMessage(udpAddr, udpConn, packet)
}

func inputsToPacket(msg, dest, fn, req, kw, budg bool) packets_client.ClientPacketI {
	var packet packets_client.ClientPacketI = nil

	if msg && !dest && !fn && !req && !kw && !budg { // sending rumor
		packet = &packets_client.PostMessagePacket{
			Message: message,
		}
	} else if msg && dest && !fn && !req && !kw && !budg { // sending private message
		packet = &packets_client.PostMessagePacket{
			Message:     message,
			Destination: destination,
		}
	} else if !msg && fn && req && !kw && !budg { // requesting a file
		packet = &packets_client.RequestFilePacket{
			Destination: destination,
			Filename:    filename,
			Request:     request,
		}
	} else if !msg && fn && !dest && !req && !kw && !budg { // indexing a file
		packet = &packets_client.IndexFilePacket{
			Filename: filename,
		}
	} else if !msg && !fn && !dest && !req && kw { // start a search
		packet = &packets_client.SearchFilesPacket{
			Keywords: keywords.List,
			Budget:   budget,
		}
	}
	return packet
}

func sendMessage(udpAddr *net.UDPAddr, udpConn *net.UDPConn, packet packets_client.ClientPacketI) {
	packetBytes, err := protobuf.Encode(packet.ToClientPacket())
	if err != nil {
		common.HandleError(err)
		return
	}
	_, err2 := udpConn.WriteToUDP(packetBytes, udpAddr)
	if err2 != nil {
		common.HandleAbort("error when sending packet", err2)
	}
}
