package hotels

import (
	"testing"

	"github.com/triplewy/sagas/utils"
)

func TestCancel(t *testing.T) {
	addr := utils.AvailableAddr()
	server, h := NewServer(addr)
	defer server.GracefulStop()
	client := NewClient(addr)

	t.Run("basic", func(t *testing.T) {
		id, err := BookRoom(client, "user0", "room0")
		if err != nil {
			t.Fatal(err)
		}
		err = CancelRoom(client, "user0", id)
		if err != nil {
			t.Fatal(err)
		}
		value, ok := h.Reservations.Get(id)
		if !ok {
			t.Fatal("Expected reservations to have key reservationID but does not exist")
		}
		reservation, ok := value.(Reservation)
		if !ok {
			t.Fatal(ErrInvalidMapType)
		}
		if reservation.Status != Canceled {
			t.Fatalf("Expected room status: %v, Got: %v\n", Canceled, reservation.Status)
		}
		_, ok = h.Rooms.Get(reservation.RoomID)
		if ok {
			t.Fatal("Expected reservation RoomID to not exist in Rooms")
		}
	})
}
