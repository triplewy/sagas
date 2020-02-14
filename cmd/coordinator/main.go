package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/triplewy/sagas"
)

var addr string

func init() {
	flag.StringVar(&addr, "addr", ":50050", "server address")
}

func main() {
	flag.Parse()
	config := sagas.DefaultConfig()

	c := sagas.NewCoordinator(config, sagas.NewBadgerDB(config.Path, config.InMemory))
	defer c.Cleanup()

	s := sagas.NewServer(addr, c)
	defer s.GracefulStop()

	log.Printf("Server listening on %v\n", addr)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
}
