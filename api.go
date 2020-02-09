package sagas

import (
	"context"
	"errors"
)

// Errors encountered by client API
var (
	ErrSagaAborted    = errors.New("saga aborted during execution")
	ErrSagaUnfinished = errors.New("saga did not finish during execution. Should never happen")
)

// StartSagaRPC starts a saga
func (c *Coordinator) StartSagaRPC(ctx context.Context, req *SagaReq) (*SagaResp, error) {
	vertices := make(map[VertexID]SagaVertex, len(req.GetVertices()))
	dag := make(map[VertexID]map[VertexID]SagaEdge, len(req.GetVertices()))

	// Populate vertices
	for id, v := range req.GetVertices() {
		tf := v.GetT()
		cf := v.GetC()
		vertices[VertexID(id)] = SagaVertex{
			VertexID: VertexID(id),
			TFunc: SagaFunc{
				URL:       tf.GetUrl(),
				Method:    tf.GetUrl(),
				RequestID: c.logs.NewRequestID(),
				Body:      tf.GetBody(),
				Resp:      make(map[string]string, 0),
			},
			CFunc: SagaFunc{
				URL:       cf.GetUrl(),
				Method:    cf.GetUrl(),
				RequestID: c.logs.NewRequestID(),
				Body:      cf.GetBody(),
				Resp:      make(map[string]string, 0),
			},
			TransferFields: v.GetTransferFields(),
			Status:         NotReached,
		}
	}

	// Populate dag
	for _, edge := range req.GetEdges() {
		dag[VertexID(edge.GetStartId())][VertexID(edge.GetEndId())] = SagaEdge{Fields: edge.GetTransferFields()}
	}

	saga := Saga{DAG: dag, Vertices: vertices}
	sagaID := c.logs.NewSagaID()

	replyCh := make(chan Saga, 1)
	c.createCh <- createMsg{
		sagaID:  sagaID,
		saga:    saga,
		replyCh: replyCh,
	}
	replySaga := <-replyCh

	finished, aborted := CheckFinishedOrAbort(replySaga.Vertices)

	if aborted {
		return nil, ErrSagaAborted
	}
	if !finished {
		return nil, ErrSagaUnfinished
	}
	return &SagaResp{}, nil
}

// BookRoom starts a new saga that makes a single transaction to book a hotel room
func BookRoom(c *Coordinator, userID, roomID string) error {
	dag := map[VertexID]map[VertexID]SagaEdge{
		"1": {},
	}

	requestID1 := c.logs.NewRequestID()
	requestID2 := c.logs.NewRequestID()

	vertices := map[VertexID]SagaVertex{
		"1": SagaVertex{
			VertexID: "1",
			TFunc: SagaFunc{
				URL:       "http://localhost:51051/book",
				Method:    "POST",
				RequestID: requestID1,
				Body: map[string]string{
					"userID": userID,
					"roomID": roomID,
				},
				Resp: make(map[string]string),
			},
			CFunc: SagaFunc{
				URL:       "http://localhost:51051/cancel",
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
	sagaID := c.logs.NewSagaID()
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
		requestID1 := c.logs.NewRequestID()
		requestID2 := c.logs.NewRequestID()

		vertices[vID] = SagaVertex{
			VertexID: vID,
			TFunc: SagaFunc{
				URL:       "http://localhost:51051/book",
				Method:    "POST",
				RequestID: requestID1,
				Body: map[string]string{
					"userID": userID,
					"roomID": roomID,
				},
				Resp: make(map[string]string),
			},
			CFunc: SagaFunc{
				URL:       "http://localhost:51051/cancel",
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
	sagaID := c.logs.NewSagaID()
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
