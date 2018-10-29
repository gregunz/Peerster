package main

import (
	"flag"
	"fmt"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/models/peers"
	"github.com/gregunz/Peerster/www"
	"sync"
)

var uiPort uint
var guiEnabled bool
var guiPort uint
var gossipAddr peers.Address
var name string
var peersSet peers.SetVar
var simple bool
var rTimerSeconds uint

var DefaultIpPort = "127.0.0.1:5000"

func init() {
	flag.UintVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.BoolVar(&guiEnabled, "GUI", false, "whether GUI is enabled (set to true if GUIPort != 0)")
	flag.UintVar(&guiPort, "GUIPort", 0, "port for the GUI client (if 0, a port is randomly assigned)")
	flag.Var(&gossipAddr, "gossipAddr", fmt.Sprintf("ip:port for the gossiper (default \"%s\")", DefaultIpPort))
	flag.StringVar(&name, "name", "", "name of the gossiper")
	flag.Var(&peersSet, "peers", "comma-separated list of peers of the form ip:port")
	flag.UintVar(&rTimerSeconds, "rtimer", 0, "route rumors sending period in seconds, 0 to disable sending of route rumors")
	flag.BoolVar(&simple, "simple", false, "run gossiper in simple broadcast mode")
}

func main() {
	parse()
	var group sync.WaitGroup

	g := gossiper.NewGossiper(simple, &gossipAddr, name, uiPort, guiPort, peersSet.ToSet(), rTimerSeconds)
	g.Start(&group)

	if guiEnabled {
		server := www.NewWebServer(g)
		server.Start()
	}

	group.Wait()
}

func parse() {
	flag.Parse()
	if gossipAddr.IsEmpty() {
		gossipAddr = *peers.NewAddress(DefaultIpPort)
	}
	if guiPort > 0 {
		guiEnabled = true
	}
}
