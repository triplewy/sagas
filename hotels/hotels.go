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
	ErrRoomAlreadyBooked   = errors.New("Room has already been booked by user")
	ErrRoomUnavailable     = errors.New("Room is unavailable to book")
	ErrReservationNotFound = errors.New("ReservationID does not exist")
	ErrIncorrectType       = errors.New("Incorrect type for map value")
	ErrNetworkBlocked      = errors.New("network is blocked")
	ErrInvalidMapType      = errors.New("invalid value type in map")
)

type Status int

const (
	Booked Status = iota + 1
	Canceled
)

type Reservation struct {
	ID     string
	UserID string
	RoomID string
	Status Status
}

type Hotels struct {
	currID       *atomic.Uint64
	Rooms        cmap.ConcurrentMap
	Reservations cmap.ConcurrentMap
	requests     cmap.ConcurrentMap

	BlockNetwork *atomic.Bool
	SlowNetwork  *atomic.Bool
}

func NewHotels() *Hotels {
	return &Hotels{
		currID:       atomic.NewUint64(0),
		Rooms:        cmap.New(),
		Reservations: cmap.New(),
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
	h.Reservations.Set(reservationID, Reservation{
		ID:     reservationID,
		UserID: req.GetUserID(),
		RoomID: req.GetRoomID(),
		Status: Booked,
	})
	return &BookReply{ReservationID: reservationID}, nil
}

func (h *Hotels) CancelRPC(ctx context.Context, req *CancelReq) (*CancelReply, error) {
	value, ok := h.Reservations.Get(req.GetReservationID())
	if !ok {
		return nil, ErrReservationNotFound
	}
	reservation, ok := value.(Reservation)
	if !ok {
		return nil, ErrIncorrectType
	}
	h.Rooms.Remove(reservation.RoomID)
	newReservation := Reservation{
		ID:     reservation.ID,
		UserID: reservation.UserID,
		RoomID: reservation.RoomID,
		Status: Canceled,
	}
	h.Reservations.Set(req.GetReservationID(), newReservation)
	return &CancelReply{}, nil
}
