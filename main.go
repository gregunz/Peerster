package main

import (
	"flag"
	"fmt"
	"github.com/gregunz/Peerster/models"
)

var uiPort string
var gossipAddr string
var name string
var peers models.Peers
var simple bool

func init() {
	flag.StringVar(&uiPort, "UIPort", "8080", "port for the UI client")
	flag.StringVar(&gossipAddr, "gossipAddr", "127.0.0.1:5000", "ip:port for the gossiper")
	flag.StringVar(&name, "name", "", "name of the gossiper")
	flag.Var(&peers, "peers", "comma-separated list of peers of the form ip:port")
	flag.BoolVar(&simple, "simple", false, "run gossiper in simple broadcast mode")
}

func main() {
	flag.Parse()

	fmt.Println(uiPort)
	fmt.Println(gossipAddr)
	fmt.Println(name)
	fmt.Println(peers.ToString(","))
	fmt.Println(simple)
}
