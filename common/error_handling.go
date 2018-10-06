package common

import "fmt"

func HandleError(e error) {
	if e != nil {
		fmt.Printf("ERROR: %s\n", e)
	}
}
