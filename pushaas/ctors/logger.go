package ctors

import (
	"log"
	"os"
)

func NewLogger() *log.Logger {
	logger := log.New(os.Stdout, "pushaas", 0)
	return logger
}
