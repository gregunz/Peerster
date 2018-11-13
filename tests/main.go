package main

import (
	"fmt"
	"github.com/gregunz/Peerster/utils"
)

func main() {
	port := "localhost:0"
	addr, conn := utils.ConnectToIpPort(port)
	fmt.Println(addr.String())
	fmt.Println(conn.LocalAddr())
	conn.Close()
}
