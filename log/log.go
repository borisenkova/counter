package log

import (
	"fmt"
	"log"
)

const defaultPrefix = 0

type logWriter struct{}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(string(bytes))
}

func InitLogging() {
	log.SetFlags(defaultPrefix)
	log.SetOutput(new(logWriter))
}
