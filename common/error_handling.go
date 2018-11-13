package common

import (
	"fmt"
	"github.com/gregunz/Peerster/logger"
)

func HandleError(e error) {
	if e != nil {
		logger.Printlnf("ERROR: %s", e)
	}
}
func HandleAbort(msg string, e error) {
	errorString := ""
	if e != nil {
		errorString = fmt.Sprintf(":\n\t->ERROR: %s", e)
	}
	logger.Printlnf("ABORT: %s%s\n", msg, errorString)
}
