package models

import "fmt"

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

func (msg SimpleMessage) AckPrint(fromClient bool) {
	if fromClient {
		fmt.Printf("CLIENT MESSAGE %s\n", msg.Contents)
	} else {
		fmt.Printf("SIMPLE MESSAGE origin %s from %s contents %s\n",
			msg.OriginalName, msg.RelayPeerAddr, msg.Contents)
	}
}
