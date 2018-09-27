package models

import (
	"fmt"
	"strings"
)

type Peers []Address

func (p *Peers) String() string {
	return fmt.Sprint(*p)
}

func (p *Peers) Set(value string) error {
	for _, ipPort := range strings.Split(value, ",") {
		*p = append(*p, StringToAddress(ipPort))
	}
	return nil
}

func (p Peers) ToStrings() []string {
	ls := make([]string, 0)
	for _, a := range p {
		ls = append(ls, a.String())
	}
	return ls
}

func (p Peers) ToString(sep string) string {
	return strings.Join(p.ToStrings(), sep)
}
