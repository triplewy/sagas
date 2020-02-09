package sagas

import (
	"errors"
)

// Errors encountered by client API
var (
	ErrSagaAborted    = errors.New("saga aborted during execution")
	ErrSagaUnfinished = errors.New("saga did not finish during execution. Should never happen")
)

// BookRoom starts a new saga that makes a single transaction to book a hotel room
func BookRoom(c *Coordinator, userID, roomID string) error {
	dag := map[VertexID]map[VertexID]SagaEdge{
		1: {},
	}

	requestID1 := c.NewRequestID()
	requestID2 := c.NewRequestID()

	vertices := map[VertexID]SagaVertex{
		1: SagaVertex{
			VertexID: 1,
			TFunc: SagaFunc{
				URL:       "hotels/book",
				Method:    "POST",
				RequestID: requestID1,
				Body: map[string]string{
					"userID": userID,
					"roomID": roomID,
				},
				Resp: make(map[string]string),
			},
			CFunc: SagaFunc{
				URL:       "hotel/cancel",
				Method:    "POST",
				RequestID: requestID2,
				Body: map[string]string{
					"userID": userID,
				},
				Resp: make(map[string]string),
			},
			TransferFields: []string{"reservationID"},
			Status:         NotReached,
		},
	}
	saga := Saga{DAG: dag, Vertices: vertices}
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

// BookMultipleRooms starts a new saga that makes a multiple transactions to book hotel rooms
func BookMultipleRooms(c *Coordinator, userID string, roomIDs []string) error {
	dag := make(map[VertexID]map[VertexID]SagaEdge)
	vertices := map[VertexID]SagaVertex{}

	for i, roomID := range roomIDs {
		vID := VertexID(i)
		requestID1 := c.NewRequestID()
		requestID2 := c.NewRequestID()

		vertices[vID] = SagaVertex{
			VertexID: vID,
			TFunc: SagaFunc{
				URL:       "hotels/book",
				Method:    "POST",
				RequestID: requestID1,
				Body: map[string]string{
					"userID": userID,
					"roomID": roomID,
				},
				Resp: make(map[string]string),
			},
			CFunc: SagaFunc{
				URL:       "hotel/cancel",
				Method:    "POST",
				RequestID: requestID2,
				Body: map[string]string{
					"userID": userID,
				},
				Resp: make(map[string]string),
			},
			Status: NotReached,
		}
		dag[vID] = map[VertexID]SagaEdge{}
	}

	saga := Saga{DAG: dag, Vertices: vertices}
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
