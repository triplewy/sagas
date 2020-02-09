package main

import (
	"errors"
	"flag"

	"github.com/abiosoft/ishell"

	"github.com/triplewy/sagas"
)

var (
	ErrInvalidNumArgs = errors.New("Number of args provided is not valid")
)
var addr string

func init() {
	flag.StringVar(&addr, "addr", ":50050", "server address")
}

func main() {
	flag.Parse()

	client := sagas.NewClient(addr)

	shell := ishell.New()

	shell.Println("Sagas Interactive Shell")

	shell.AddCmd(&ishell.Cmd{
		Name: "book",
		Help: "book <userID> <roomID>",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 2 {
				c.Err(ErrInvalidNumArgs)
			}
			err := sagas.BookRoom(client, c.Args[0], c.Args[1])
			if err != nil {
				c.Println(err)
			} else {
				c.Println("Saga successfull!")
			}
		},
	})

	shell.Run()
}
