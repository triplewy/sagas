package hotels

import (
	"testing"
	"time"

	"google.golang.org/grpc/status"

	"github.com/triplewy/sagas/utils"
)

func TestBook(t *testing.T) {
	addr := utils.AvailableAddr()
	server, h := NewServer(addr)
	defer server.GracefulStop()
	client := NewClient(addr)

	t.Run("basic", func(t *testing.T) {
		reply, err := BookRoom(client, "user0", "room0")
		if err != nil {
			t.Fatal(err)
		}
		if reply.GetReservationID() != "1" {
			t.Fatalf("Expected: %v, Got: %v\n", "1", reply.GetReservationID())
		}
	})

	t.Run("blocked network", func(t *testing.T) {
		errCh := make(chan error, 1)

		h.blockNetwork.Store(true)
		go func() {
			time.Sleep(2 * time.Second)
			h.blockNetwork.Store(false)
		}()

		go func() {
			_, err := BookRoom(client, "user1", "room1")
			errCh <- err
			close(errCh)
		}()

		select {
		case err := <-errCh:
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("slow network", func(t *testing.T) {
		h.slowNetwork.Store(true)
		go func() {
			time.Sleep(5 * time.Second)
			h.slowNetwork.Store(false)
		}()

		_, err := BookRoom(client, "user2", "room2")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("already booked", func(t *testing.T) {
		_, err := BookRoom(client, "user3", "room3")
		if err != nil {
			t.Fatal(err)
		}
		_, err = BookRoom(client, "user3", "room3")
		if err == nil {
			t.Fatal("Expected error but got no error")
		}
		st := status.Convert(err)
		if st.Message() != ErrRoomAlreadyBooked.Error() {
			t.Fatalf("Expected error: %v, Got: %v\n", ErrRoomAlreadyBooked, st.Message())
		}
		_, err = BookRoom(client, "user4", "room3")
		if err == nil {
			t.Fatal("Expected error but got no error")
		}
		st = status.Convert(err)
		if st.Message() != ErrRoomUnavailable.Error() {
			t.Fatalf("Expected error: %v, Got: %v\n", ErrRoomUnavailable, st.Message())
		}
	})
}
