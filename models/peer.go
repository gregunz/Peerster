package models

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"net"
	"strconv"
	"strings"
)

type Peer struct {
	ip   string
	port string
}

func (p Peer) ToIpPort() string {
	return fmt.Sprintf("%s:%s", p.ip, p.port)
}

func (p Peer) String() string {
	return p.ToIpPort()
}

func StringToPeer(s string) Peer {
	ipPort := strings.Split(s, ":")
	if len(ipPort) != 2 {
		common.HandleError(
			fmt.Errorf("Peer string must be of the form \"ip:port\" instead of %s", s))
	}
	return Peer{
		ip:   ipPort[0],
		port: ipPort[1],
	}
}

func (p Peer) ToUDPAddr() net.UDPAddr {
	port, err := strconv.Atoi(p.port)
	common.HandleError(err)
	return net.UDPAddr{
		IP:   net.ParseIP(p.ip),
		Port: port,
	}
}
