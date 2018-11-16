package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type _logger struct {
	isSetup bool
	sync.RWMutex
}

var logger = _logger{isSetup: false}

func setup() {

	logger.RLock()
	if logger.isSetup {
		logger.RUnlock()
		return
	}
	logger.RUnlock()

	logger.Lock()
	defer logger.Unlock()

	if logger.isSetup {
		return
	}

	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(os.Stdout)
	logger.isSetup = true
}

func Printlnf(s string, v ...interface{}) {
	setup()
	text := fmt.Sprintf(s, v...)
	log.Printf("%s\n", text)
}
