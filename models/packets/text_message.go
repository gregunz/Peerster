package packets

import "fmt"

type TextMessage struct {
	Message string `json:"message"`
}

func (packet *TextMessage) AckPrint() {
	fmt.Printf("CLIENT MESSAGE %s\n", packet.Message)
}

func (packet TextMessage) String() string {
	return fmt.Sprintf("TEXT MESSAGE %s\n", packet.Message)
}
