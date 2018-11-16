package main

import (
	"flag"
	"fmt"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"github.com/gregunz/Peerster/www"
	"sync"
)

const (
	defaultIpPort = "127.0.0.1:5000"
)

var uiPort uint
var guiPort uint
var gossipAddr peers.Address
var name string
var peersSet peers.SetVar
var simple bool
var rTimerSeconds uint

func init() {
	flag.UintVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.UintVar(&guiPort, "GUIPort", 0, "port for the GUI client (if 0, gui is disabled)")
	flag.Var(&gossipAddr, "gossipAddr", fmt.Sprintf("ip:port for the gossiper (default \"%s\")", defaultIpPort))
	flag.StringVar(&name, "name", "", "name of the gossiper")
	flag.Var(&peersSet, "peers", "comma-separated list of peers of the form ip:port")
	flag.UintVar(&rTimerSeconds, "rtimer", 0, "route rumors sending period in seconds, 0 to disable sending of route rumors")
	flag.BoolVar(&simple, "simple", false, "run gossiper in simple broadcast mode")
}

func main() {
	parse()
	guiEnabled := guiPort > 0

	var group sync.WaitGroup
	peerster := gossiper.NewGossiper(simple, &gossipAddr, name, uiPort, guiPort, peersSet.ToSet(), rTimerSeconds, guiEnabled)

	if peerster != nil {

		peerster.Start(&group)

		if guiEnabled {
			server := www.NewWebServer(peerster)
			server.Start(group)
		}
	}
	group.Wait()
}

func parse() {
	flag.Parse()
	if gossipAddr.IsEmpty() {
		gossipAddr = *peers.NewAddress(defaultIpPort)
	}
}
