package main

import (
	"flag"
	"fmt"
)

var uiPort string
var msg string

func init() {
	flag.StringVar(&uiPort, "UIPort", "8080", "port for the UI client")
	flag.StringVar(&msg, "name", "", "message to be sent")
}

func main() {
	flag.Parse()

	fmt.Println(uiPort)
	fmt.Println(msg)
}
