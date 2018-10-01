package main

import (
	"flag"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/models"
)

var uiPort uint
var gossipAddr models.Address
var name string
var peers models.PeersSet
var simple bool

func init() {
	flag.UintVar(&uiPort, "UIPort", 8080, "port for the UI client")
	flag.Var(&gossipAddr, "gossipAddr", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
	flag.StringVar(&name, "name", "", "name of the gossiper")
	flag.Var(&peers, "peers", "comma-separated list of peers of the form ip:port")
	flag.BoolVar(&simple, "simple", false, "run gossiper in simple broadcast mode")
}

func main() {
	flag.Parse()

	g := gossiper.NewGossiper(&gossipAddr, name, uiPort, &peers)
	g.Start()
}
