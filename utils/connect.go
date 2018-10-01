package utils

import (
	"github.com/gregunz/Peerster/common"
	"net"
)

func IpPortToUDPAddr(ipPort string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp4", ipPort)
	common.HandleError(err)
	return addr
}

func ConnectToUDPAddr(addr *net.UDPAddr) *net.UDPConn {
	conn, err := net.ListenUDP("udp4", addr)
	common.HandleError(err)
	return conn
}

func ConnectToIpPort(ipPort string) (*net.UDPAddr, *net.UDPConn) {
	addr := IpPortToUDPAddr(ipPort)
	conn := ConnectToUDPAddr(addr)
	return addr, conn
}
