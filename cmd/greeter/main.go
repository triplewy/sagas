package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/triplewy/sagas/greeter"
)

func main() {
	greeter.NewServer(":50051")
	log.Printf("greeter-server listening on %v\n", "localhost:50051")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
}
