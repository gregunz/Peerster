package common

import "fmt"

func HandleError(e error) {
	if e != nil {
		fmt.Printf("ERROR: %s\n", e)
	}
}
func HandleAbort(msg string, e error) {
	errorString := ""
	if e != nil {
		errorString = fmt.Sprintf(":\n\t->ERROR: %s", e)
	}
	fmt.Printf("ABORT: %s%s\n", msg, errorString)
}
