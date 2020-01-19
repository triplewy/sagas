package sagas

import (
	"sync/atomic"
)

type Sagas struct {
	currID uint64
	tail   uint64
	store  *Store
}

func NewSagas() *Sagas {
	return &Sagas{
		currID: 0,
		tail:   0,
		store:  NewStore(),
	}
}

func (s *Sagas) NewID() uint64 {
	return atomic.AddUint64(&s.currID, 1)
}

func (s *Sagas) StartSaga() {
	// Get new saga ID
	id := s.NewID()
	// Persist start saga to log
	s.store.Append(Log{
		SagaID: id,
		Name:   "saga",
		Status: Status_Start,
		Data:   []byte{},
	})
	// Concurrently call bookHotel and bookFlight
	go func() {
		// Persist start of transaction to log
		s.store.Append(Log{
			SagaID: id,
			Name:   "hotel",
			Status: Status_Start,
			Data:   []byte{},
		})
		// Call remote rpc to book hotel
		// If err != nil, retry
		// Persist transaction end or abort to log depending on reply
	}()

	go func() {
		s.store.Append(Log{
			SagaID: id,
			Name:   "flight",
			Status: Status_Start,
			Data:   []byte{},
		})
		// Call remote rpc to book hotel
		// If err != nil, retry
		// Persist transaction end or abort to log depending on reply
	}()
}
