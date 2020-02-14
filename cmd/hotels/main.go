package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/triplewy/sagas/hotels"
)

var addr string

func init() {
	flag.StringVar(&addr, "addr", ":50051", "address for hotels server")
}

func main() {
	flag.Parse()

	hotels.NewServer(addr)
	log.Printf("Server listening on %v\n", addr)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
}
