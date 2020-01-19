package trips

import (
	"context"
	"sync"
	"sync/atomic"
)

type Trips struct {
	currID uint64
	trips  map[uint64][]byte
	sync.Mutex
}

func NewTrips() *Trips {
	return &Trips{
		currID: 0,
		trips:  make(map[uint64][]byte),
	}
}

func (t *Trips) NewID() uint64 {
	return atomic.AddUint64(&t.currID, 1)
}

// BookRPC books a trip that includes flight and hotel. Represents the start of a Saga
func (t *Trips) BookRPC(ctx context.Context, req *BookReq) (*BookReply, error) {
	// Persist start of saga and start trips
	// proxy.StartSaga(id string, data interface{})
	// Perform business logic
	id := t.NewID()
	t.Lock()
	t.trips[id] = nil
	t.Unlock()
	// Persist end trips
	// proxy.Persist(id string, data interface{})

	// Return success
	return &BookReply{TripID: id}, nil
}

func (t *Trips) CancelRPC(ctx context.Context, req *CancelReq) (*CancelReply, error) {
	// Perform business logic
	t.Lock()
	delete(t.trips, req.GetTripID())
	t.Unlock()
	// Persist comp trips
	// proxy.Persist(id string)
	return &CancelReply{}, nil
}
