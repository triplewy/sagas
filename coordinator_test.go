package sagas

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/triplewy/sagas/hotels"
	"gotest.tools/assert"
)

// Need envoy to be running for this test to work
func TestCoordinator(t *testing.T) {
	config := DefaultConfig()

	hServer, h := hotels.NewServer(config.HotelsAddr)
	defer hServer.GracefulStop()

	c := NewCoordinator(config, NewBadgerDB(config.Path, config.InMemory))
	defer c.Cleanup()

	cServer := NewServer(config.CoordinatorAddr, c)
	defer cServer.GracefulStop()

	client := NewClient(config.CoordinatorAddr)

	t.Run("1 transaction", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			err := BookRoom(client, "user0", "room0")
			assert.NilError(t, err)
			value, ok := h.Rooms.Get("room0")
			assert.Assert(t, ok)
			assert.Equal(t, "user0", value.(string))
		})

		t.Run("abort", func(t *testing.T) {
			err := BookRoom(client, "user0", "room0")
			assert.Equal(t, err, ErrSagaAborted)
		})

		t.Run("entity temporary fail", func(t *testing.T) {
			h.BlockNetwork.Store(true)
			go func() {
				time.Sleep(3 * time.Second)
				h.BlockNetwork.Store(false)
			}()
			err := BookRoom(client, "user1", "room1")
			assert.NilError(t, err)
			value, ok := h.Rooms.Get("room1")
			assert.Assert(t, ok)
			assert.Equal(t, "user1", value.(string))
		})
	})

	t.Run("2 transactions", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			err := BookMultipleRooms(client, "user2", []string{"room20", "room21"})
			assert.NilError(t, err)
			value, ok := h.Rooms.Get("room20")
			assert.Assert(t, ok)
			assert.Equal(t, "user2", value.(string))
			value, ok = h.Rooms.Get("room21")
			assert.Assert(t, ok)
			assert.Equal(t, "user2", value.(string))
		})

		t.Run("1 abort", func(t *testing.T) {
			err := BookRoom(client, "user3", "room30")
			assert.NilError(t, err)
			value, ok := h.Rooms.Get("room30")
			assert.Assert(t, ok)
			assert.Equal(t, "user3", value.(string))
			err = BookMultipleRooms(client, "user3", []string{"room30", "room31"})
			assert.Equal(t, err, ErrSagaAborted)
			_, ok = h.Rooms.Get("room31")
			assert.Assert(t, !ok)
		})

		t.Run("2 abort", func(t *testing.T) {
			err := BookRoom(client, "user4", "room40")
			assert.NilError(t, err)
			err = BookRoom(client, "user4", "room41")
			assert.NilError(t, err)
			value, ok := h.Rooms.Get("room40")
			assert.Assert(t, ok)
			assert.Equal(t, "user4", value.(string))
			value, ok = h.Rooms.Get("room41")
			assert.Assert(t, ok)
			assert.Equal(t, "user4", value.(string))
			err = BookMultipleRooms(client, "user4", []string{"room40", "room41"})
			assert.Equal(t, err, ErrSagaAborted)
		})
	})
}

func TestCoordinatorFailure(t *testing.T) {
	config := DefaultConfig()
	server, h := hotels.NewServer(config.HotelsAddr)
	defer server.GracefulStop()

	t.Run("1 transaction", func(t *testing.T) {
		t.Run("before rpc", func(t *testing.T) {
			// Block network so sagas cannot communicate with entity service
			h.BlockNetwork.Store(true)

			// Start first coordinator
			cmd := exec.Command("bin/sagas", "-user", "user5", "-room", "room5")
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

			// Check if reservation failed
			_, ok := h.Rooms.Get("room5")
			assert.Assert(t, !ok)

			c := NewCoordinator(config, NewBadgerDB(config.Path, config.InMemory))
			defer c.Cleanup()

			h.BlockNetwork.Store(false)

			time.Sleep(2 * time.Second)

			lastLog, err := c.logs.GetLog(c.logs.LastIndex())
			assert.NilError(t, err)

			fmt.Printf("%#v\n", lastLog)
		})

		t.Run("after rpc", func(t *testing.T) {

		})
	})
}
