package utils

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"net"
)

func IpPortToUDPAddr(ipPort string) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp4", ipPort)
	if err != nil {
		common.HandleError(err)
		return nil
	}
	if addr.IP == nil {
		common.HandleError(fmt.Errorf("cannot resolve <ip:port> = <%s>", ipPort))
		return nil
	}
	return addr
}

func ConnectToUDPAddr(addr *net.UDPAddr) *net.UDPConn {
	if addr == nil {
		common.HandleError(fmt.Errorf("cannot connect to nil udp address"))
		return nil
	}
	conn, err := net.ListenUDP("udp4", addr)
	common.HandleError(err)
	return conn
}

func ConnectToIpPort(ipPort string) (*net.UDPAddr, *net.UDPConn) {
	addr := IpPortToUDPAddr(ipPort)
	conn := ConnectToUDPAddr(addr)
	return addr, conn
}
