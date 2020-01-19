package trips

import (
	"context"
	"log"
	"testing"

	"github.com/triplewy/sagas"
)

func TestBook(t *testing.T) {
	addr := sagas.AvailableAddr()
	server := NewServer(addr)
	defer server.GracefulStop()
	client := NewClient(addr)

	reply, err := client.BookRPC(context.Background(), &BookReq{
		UserID:   0,
		HotelID:  0,
		FlightID: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	log.Println(reply)
}
