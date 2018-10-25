package common

import (
	"fmt"
	"log"
)

func HandleError(e error) {
	if e != nil {
		log.Printf("ERROR: %s\n", e)
	}
}
func HandleAbort(msg string, e error) {
	errorString := ""
	if e != nil {
		errorString = fmt.Sprintf(":\n\t->ERROR: %s", e)
	}
	log.Printf("ABORT: %s%s\n", msg, errorString)
}
