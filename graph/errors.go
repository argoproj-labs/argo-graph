package graph

import log "github.com/sirupsen/logrus"

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
