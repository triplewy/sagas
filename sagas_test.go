package sagas

import (
	"log"
	"os/exec"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/triplewy/sagas/hotels"
)

func TestSagasBasic(t *testing.T) {
	config := DefaultConfig()

	server, _ := hotels.NewServer(config.HotelsAddr)
	hClient := hotels.NewClient(config.HotelsAddr)
	defer server.GracefulStop()

	s := NewSagas(config)
	defer s.Cleanup()

	t.Run("1 transaction", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			s.StartSaga("user0", "room0")
			index := s.LastIndex()
			if index != 4 {
				t.Fatalf("Expected: %d, Got: %d\n", 4, index)
			}
			log, err := s.GetLog(index)
			if err != nil {
				t.Fatal(err)
			}
			expectedLog := Log{
				SagaID: 1,
				Name:   "saga",
				Status: End,
				Data:   []byte{},
			}
			if !expectedLog.Equal(log) {
				t.Fatalf("Expected: %v, Got: %v\n", expectedLog, log)
			}
			_, err = hotels.BookRoom(hClient, "user0", "room0")
			st := status.Convert(err)
			if st.Message() != hotels.ErrRoomAlreadyBooked.Error() {
				t.Fatalf("Expected: %v, Got: %v\n", hotels.ErrRoomAlreadyBooked, st.Message())
			}
		})

		t.Run("fail", func(t *testing.T) {
			_, err := hotels.BookRoom(hClient, "user1", "room1")
			if err != nil {
				t.Fatal(err)
			}
			s.StartSaga("user1", "room1")
			index := s.LastIndex()
			log, err := s.GetLog(index - 1)
			if err != nil {
				t.Fatal(err)
			}
			st := status.New(codes.Internal, hotels.ErrRoomAlreadyBooked.Error())
			if log.Name != "hotel" || string(log.Data) != st.Err().Error() {
				t.Fatalf("Expected: %v, Got: %v\n", st.Err().Error(), string(log.Data))
			}
		})
	})
}

func TestSagasCoordinatorFailure(t *testing.T) {
	config := DefaultConfig()
	server, h := hotels.NewServer(config.HotelsAddr)
	defer server.GracefulStop()

	t.Run("1 transaction", func(t *testing.T) {
		t.Run("before rpc", func(t *testing.T) {
			// Block network so sagas cannot communicate with entity service
			h.BlockNetwork.Store(true)

			// Start first coordinator
			cmd := exec.Command("bin/sagas", "-user", "user2", "-room", "room2")
			err := cmd.Start()
			if err != nil {
				t.Fatal(err)
			}
			// Wait for coordinator to make some requests
			time.Sleep(2 * time.Second)

			// Kill coordinator
			err = cmd.Process.Kill()
			if err != nil {
				t.Fatal(err)
			}

			s := NewSagas(config)
			defer s.Cleanup()

			log.Println(s.LastIndex())
		})

		t.Run("after rpc", func(t *testing.T) {

		})
	})
}
