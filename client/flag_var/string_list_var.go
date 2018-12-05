package flag_var

import (
	"strings"
)

type StringListVar struct {
	List []string
}

func (slVar *StringListVar) Set(s string) error {
	for _, e := range strings.Split(s, ",") {
		slVar.List = append(slVar.List, e)
	}
	return nil
}

func (slVar *StringListVar) String() string {
	ls := []string{}
	for _, e := range slVar.List {
		ls = append(ls, e)
	}
	return strings.Join(ls, ",")
}
