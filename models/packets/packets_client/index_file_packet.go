package packets_client

import "fmt"

type IndexFilePacket struct {
	File string `json:"file"`
}

func (packet IndexFilePacket) String() string {
	return fmt.Sprintf("INDEX FILE %s", packet.File)
}

func (packet *IndexFilePacket) AckPrint() {
	fmt.Println(packet.String())
}

func (packet *IndexFilePacket) ToClientPacket() *ClientPacket {
	return &ClientPacket{
		IndexFile: packet,
	}
}
