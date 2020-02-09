package main

import (
	"flag"
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

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
}
