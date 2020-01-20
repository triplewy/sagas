package hotels

import (
	context "context"
	"errors"
	"strconv"

	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/atomic"
)

// Errors involving hotel business logic
var (
	ErrRoomAlreadyBooked = errors.New("Room has already been booked by user")
	ErrRoomUnavailable   = errors.New("Room is unavailable to book")
	ErrNetworkBlocked    = errors.New("network is blocked")
	ErrInvalidMapType    = errors.New("invalid value type in map")
)

type Status int

const (
	Booked Status = iota + 1
	Canceled
)

type Hotels struct {
	currID       *atomic.Uint64
	Rooms        cmap.ConcurrentMap
	reservations cmap.ConcurrentMap
	requests     cmap.ConcurrentMap

	BlockNetwork *atomic.Bool
	SlowNetwork  *atomic.Bool
}

func NewHotels() *Hotels {
	return &Hotels{
		currID:       atomic.NewUint64(0),
		Rooms:        cmap.New(),
		reservations: cmap.New(),
		requests:     cmap.New(),

		BlockNetwork: atomic.NewBool(false),
		SlowNetwork:  atomic.NewBool(false),
	}
}

func (h *Hotels) newID() string {
	return strconv.FormatUint(h.currID.Inc(), 10)
}

func (h *Hotels) BookRPC(ctx context.Context, req *BookReq) (*BookReply, error) {
	// Check if room has already been booked
	if value, ok := h.Rooms.Get(req.RoomID); ok {
		if bookedUserID, ok := value.(string); ok {
			// Switch error based on if user has booked room themselves
			if req.UserID == bookedUserID {
				return nil, ErrRoomAlreadyBooked
			}
			return nil, ErrRoomUnavailable
		}
		return nil, ErrInvalidMapType
	}
	h.Rooms.Set(req.RoomID, req.UserID)

	// Set reservation to Booked
	reservationID := h.newID()
	h.reservations.Set(reservationID, Booked)
	return &BookReply{ReservationID: reservationID}, nil
}

func (h *Hotels) CancelRPC(ctx context.Context, req *CancelReq) (*CancelReply, error) {
	h.reservations.Set(req.GetReservationID(), Canceled)
	return &CancelReply{}, nil
}
