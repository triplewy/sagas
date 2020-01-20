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
		value, ok := h.reservations.Get(id)
		if !ok {
			t.Fatal("Expected reservations to have key reservationID but does not exist")
		}
		status, ok := value.(Status)
		if !ok {
			t.Fatal(ErrInvalidMapType)
		}
		if status != Canceled {
			t.Fatalf("Expected room status: %v, Got: %v\n", Canceled, status)
		}
	})
}
