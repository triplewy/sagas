package sagas

import "errors"

// Errors encountered by client API
var (
	ErrSagaAborted    = errors.New("saga aborted during execution")
	ErrSagaUnfinished = errors.New("saga did not finish during execution. Should never happen")
)

// NewHotelSaga starts a new saga that makes a single transaction to book a hotel room
func NewHotelSaga(c *Coordinator, userID, roomID string) error {
	dag := map[VertexID]map[VertexID]struct{}{
		1: {},
	}
	vertices := map[VertexID]SagaVertex{
		1: SagaVertex{
			VertexID: 1,
			TFunc: SagaFunc{
				FuncID:    "hotel_book",
				RequestID: "__required___",
				Input: map[string]interface{}{
					"userID": userID,
					"roomID": roomID,
				},
				Output: map[string]interface{}{},
			},
			CFunc: SagaFunc{
				FuncID:    "hotel_cancel",
				RequestID: "___required___",
				Input:     map[string]interface{}{},
				Output:    map[string]interface{}{},
			},
			Status: NotReached,
		},
	}
	saga := NewSaga(dag, vertices)
	sagaID := c.NewSagaID()
	// replyCh has buffer of 1 so that coordinator is not blocking if reply chan is not being read
	replyCh := make(chan Saga, 1)
	c.createCh <- createMsg{
		sagaID:  sagaID,
		saga:    saga,
		replyCh: replyCh,
	}
	replySaga := <-replyCh
	finished, aborted := CheckFinishedOrAbort(replySaga.Vertices)
	if aborted {
		return ErrSagaAborted
	}
	if !finished {
		return ErrSagaUnfinished
	}
	return nil
}
