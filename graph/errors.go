package graph

import (
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

func checkError(err error) {
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
}
