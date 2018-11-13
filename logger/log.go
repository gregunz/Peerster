package logger

import (
	"log"
	"os"
)

var isSetup = false

func setup() {
	if !isSetup {
		log.SetOutput(os.Stdout)
		isSetup = true
	}
}

func Printlnf(s string, v ...interface{}) {
	setup()
	log.Printf(s+"\n", v)
}
