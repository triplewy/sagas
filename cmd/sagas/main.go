package main

import (
	"flag"
	"log"

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

	c := sagas.NewCoordinator(config)
	defer c.Cleanup()

	err := sagas.NewHotelSaga(c, userID, roomID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("saga successful!")
}
