package main

import (
	"flag"
	"fmt"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/models"
)

var uiPort uint
var gossipAddr models.Address
var name string
var peers models.PeersSet
var simple bool

var DefaultIpPort = "127.0.0.1:5000"

func init() {
	flag.UintVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.Var(&gossipAddr, "gossipAddr", fmt.Sprintf("ip:port for the gossiper (default \"%s\")", DefaultIpPort))
	flag.StringVar(&name, "name", "", "name of the gossiper")
	flag.Var(&peers, "peers", "comma-separated list of peers of the form ip:port")
	flag.BoolVar(&simple, "simple", false, "run gossiper in simple broadcast mode")
}

func main() {
	parse()
	g := gossiper.NewGossiper(simple, &gossipAddr, name, uiPort, &peers)
	g.Start()
}

func parse() {
	flag.Parse()
	if gossipAddr.IsEmpty() {
		gossipAddr = *models.NewAddress(DefaultIpPort)
	}
}