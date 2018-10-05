package models

import (
	"fmt"
	"net"
)

type Address struct {
	UDPAddr *net.UDPAddr
}

func (addr *Address) ToIpPort() string {
	if addr.UDPAddr != nil {
		return fmt.Sprintf("%s:%d", addr.UDPAddr.IP, addr.UDPAddr.Port)
	}
	return ""
}

func (addr Address) String() string {
	return addr.ToIpPort()
}

func (addr *Address) Set(value string) error {
	udpAddr, err := net.ResolveUDPAddr("udp4", value)
	addr.UDPAddr = udpAddr
	if err != nil {
		return err
	}
	return nil
}

func (addr *Address) Equals(other *Address) bool {
	return addr.ToIpPort() == other.ToIpPort()
}
