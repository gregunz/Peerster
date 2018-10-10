package peers

import (
	"fmt"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/utils"
	"net"
)

type Address struct {
	udpAddr *net.UDPAddr
}

func (addr *Address) ToIpPort() string {
	if addr.IsEmpty() {
		common.HandleError(fmt.Errorf(
			"cannot return <ip:port> of a nil address, returning empty string"))
		return "<nil>"
	}
	return fmt.Sprintf("%s:%d", addr.udpAddr.IP, addr.udpAddr.Port)
}

func (addr Address) String() string {
	if addr.IsEmpty() {
		return ""
	}
	return addr.ToIpPort()
}

func (addr *Address) Set(s string) error {
	*addr = *NewAddress(s)
	return nil
}

func (addr *Address) Equals(other *Address) bool {
	return addr.ToIpPort() == other.ToIpPort()
}

func (addr *Address) UDP() *net.UDPAddr {
	return addr.udpAddr
}

func NewAddress(ipPort string) *Address {
	return &Address{
		udpAddr: utils.IpPortToUDPAddr(ipPort),
	}
}

func (addr *Address) IsEmpty() bool {
	return addr.udpAddr == nil
}

func (addr *Address) NonEmpty() bool {
	return !addr.IsEmpty()
}
