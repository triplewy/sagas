package hotels

import (
	context "context"
	"errors"
	"sync"

	"go.uber.org/atomic"
)

var (
	ErrReservationDoesNotExist = errors.New("Reservation does not exist")
	ErrNetworkBlocked          = errors.New("network is blocked")
)

type Status int

const (
	Booked Status = iota + 1
	Canceled
)

type Hotels struct {
	currID       *atomic.Uint64
	blockNetwork *atomic.Bool
	reservations map[uint64]Status
	sync.Mutex
}

func NewHotels() *Hotels {
	return &Hotels{
		currID:       atomic.NewUint64(0),
		blockNetwork: atomic.NewBool(false),
		reservations: make(map[uint64]Status),
	}
}

func (h *Hotels) newID() uint64 {
	return h.currID.Inc()
}

func (h *Hotels) BookRPC(ctx context.Context, req *BookReq) (*BookReply, error) {
	h.Lock()
	defer h.Unlock()

	id := h.newID()
	if _, ok := h.reservations[id]; !ok {
		h.reservations[id] = Booked
	}

	return &BookReply{Id: id}, nil
}

func (h *Hotels) CancelRPC(ctx context.Context, req *CancelReq) (*CancelReply, error) {
	h.Lock()
	defer h.Unlock()

	h.reservations[req.GetId()] = Canceled

	return &CancelReply{}, nil
}
