package packets_client

import "fmt"

type IndexFilePacket struct {
	Filename string `json:"filename"`
}

func (packet IndexFilePacket) String() string {
	return fmt.Sprintf("INDEX FILE %s", packet.Filename)
}

func (packet *IndexFilePacket) AckPrint() {
	fmt.Println(packet.String())
}

func (packet *IndexFilePacket) ToClientPacket() *ClientPacket {
	return &ClientPacket{
		IndexFile: packet,
	}
}
