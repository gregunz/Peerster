package models

import (
	"fmt"
	"strings"
)

type Address struct {
	ip   string
	port string
}

func (a Address) String() string {
	return fmt.Sprintf("%s:%s", a.ip, a.port)
}

func StringToAddress(s string) Address {
	ipPort := strings.Split(s, ":")
	return Address{
		ip:   ipPort[0],
		port: ipPort[1],
	}
}
