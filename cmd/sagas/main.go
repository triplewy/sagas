package main

import (
	"flag"

	"github.com/triplewy/sagas"
)

var userID string
var roomID string

func init() {
	flag.StringVar(&userID, "user", "user0", "userID used in saga")
	flag.StringVar(&roomID, "room", "room0", "roomID used in saga")
}

func main() {
	flag.Parse()

	config := sagas.DefaultConfig()
	s := sagas.NewCoordinator(config)
	s.StartSaga(userID, roomID)
}
