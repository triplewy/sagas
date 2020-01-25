package sagas

import (
	"log"
	"os"
)

// Error is high-risk logger
var Error *log.Logger

// Initialize the loggers
func init() {
	Error = log.New(os.Stderr, "ERROR: ", log.Ltime|log.Lshortfile)
}
