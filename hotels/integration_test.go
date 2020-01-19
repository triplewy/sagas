package hotels

import (
	"testing"
	"time"

	"github.com/triplewy/sagas/utils"
)

func TestBook(t *testing.T) {
	addr := utils.AvailableAddr()
	server, h := NewServer(addr)
	defer server.GracefulStop()
	client := NewClient(addr)

	reply := BookHotel(client, 0)
	if reply.GetId() != 1 {
		t.Fatalf("Expected: %v, Got: %v\n", 1, reply.GetId())
	}

	replyCh := make(chan *BookReply, 1)
	h.blockNetwork.Store(true)
	go func() {
		time.Sleep(2 * time.Second)
		h.blockNetwork.Store(false)
	}()

	go func() {
		replyCh <- BookHotel(client, 0)
	}()

	reply = <-replyCh
	if reply.GetId() != 1 {
		t.Fatalf("Expected: %v, Got: %v\n", 1, reply.GetId())
	}
}
