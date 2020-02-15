package sagas

import (
	"errors"
)

// Errors from incorrect coordinator logic
var (
	ErrInvalidStatus         = errors.New("Vertex's status is invalid in the context of the saga")
	ErrInvalidFuncID         = errors.New("invalid saga func ID")
	ErrInvalidFuncInputField = errors.New("field does not exist in input for saga func")
	ErrInvalidFuncInputType  = errors.New("incorrect type for input field in saga func")
	ErrSagaIDAlreadyExists   = errors.New("create saga's sagaID already exists")
	ErrSagaIDNotFound        = errors.New("update sagaID does not exist in coordinator's map")
)

type updateMsg struct {
	sagaID string
	vertex Vertex
}

type createMsg struct {
	saga    Saga
	replyCh chan Saga
}

// Coordinator handles saga requests by calling RPCs and persisting logs to disk
type Coordinator struct {
	Config   *Config
	logs     LogStore
	sagas    map[string]Saga
	requests map[string]chan Saga

	createCh chan createMsg
	updateCh chan updateMsg
}

// NewCoordinator creates a new coordinator based on a config
func NewCoordinator(config *Config, logStore LogStore) *Coordinator {
	c := &Coordinator{
		Config:   config,
		logs:     logStore,
		sagas:    make(map[string]Saga),
		requests: make(map[string]chan Saga),

		createCh: make(chan createMsg),
		updateCh: make(chan updateMsg),
	}

	go c.Run()

	if config.AutoRecover {
		sagas := Recover(c.logs)
		c.sagas = sagas
	}

	return c
}

// // RunSaga first checks if saga is already finished. Each unfinished saga vertex then runs in its own goroutine
// func (c *Coordinator) RunSaga(saga Saga) {
// 	// Check if saga already finished
// 	finished := CheckFinished(saga.Vertices)

// 	if finished {
// 		// Notify request that saga has finished
// 		if replyCh, ok := c.requests[saga.ID]; ok {
// 			replyCh <- saga
// 			delete(c.requests, saga.ID)
// 		}
// 		return
// 	}

// 	err := CheckValidSaga(saga, aborted)
// 	if err != nil {
// 		panic(err)
// 	}

// 	if aborted {
// 		c.RollbackSaga(saga)
// 	} else {
// 		c.ContinueSaga(saga)
// 	}
// }

// // RollbackSaga calls necessary compensating functions to undo saga
// func (c *Coordinator) RollbackSaga(saga Saga) {
// 	// Each vertex that is not Status_ABORT, Status_NOT_REACHED, or  Status_END_C still needs to fully rollback
// 	for _, vertex := range saga.Vertices {
// 		if !(vertex.Status == Status_ABORT || vertex.Status == Status_END_C || vertex.Status == Status_NOT_REACHED) {
// 			go c.ProcessC(saga.ID, vertex)
// 		}
// 	}
// }

// // ContinueSaga runs saga forward
// func (c *Coordinator) ContinueSaga(saga Saga) {
// 	// Build set of vertices that are children
// 	seen := make(map[string]struct{}, 0)
// 	for _, children := range saga.DAG {
// 		for id := range children {
// 			seen[id] = struct{}{}
// 		}
// 	}

// 	var highest []string
// 	// For all vertices, the ones not in seen are highest vertices
// 	for id := range saga.Vertices {
// 		isChild := func() bool {
// 			for seenID := range seen {
// 				if seenID == id {
// 					return true
// 				}
// 			}
// 			return false
// 		}()
// 		if !isChild {
// 			highest = append(highest, id)
// 		}
// 	}

// 	// Find list of vertices to process using bfs
// 	process := make(map[string]Vertex)
// 	var id string

// 	for len(highest) > 0 {
// 		// pop first vertex off of highest
// 		id, highest = highest[0], highest[1:]

// 		v, ok := saga.Vertices[id]
// 		if !ok {
// 			panic(ErrIDNotFound)
// 		}
// 		// if vertex is not Status_END_T, add to process
// 		if v.Status != Status_END_T {
// 			process[id] = v
// 		} else {
// 			// add all children to highest
// 			for childID := range saga.DAG[id] {
// 				highest = append(highest, childID)
// 			}
// 		}
// 	}

// 	// Run each vertex to process on a separate goroutine
// 	for _, vertex := range process {
// 		go c.ProcessT(saga.ID, vertex)
// 	}
// }

// ProcessT runs a Vertex's  T
func (c *Coordinator) ProcessT(sagaID string, vertex Vertex) {
	// Sanity check on vertex's status
	if vertex.Status == Status_END_T {
		return
	}
	if !(vertex.Status == Status_START_T || vertex.Status == Status_NOT_REACHED) {
		panic(ErrInvalidSaga)
	}

	// Append to log
	data := encodeVertex(vertex)
	c.logs.AppendLog(sagaID, VertexLog, data)

	// Evaluate vertex's function
	f := vertex.T

	resp, err := HTTPReq(f.GetUrl(), f.GetMethod(), f.GetRequestId(), f.GetBody())
	status := Status_END_T
	if err != nil {
		Error.Println(err)
		// Store error to output
		f.Resp["error"] = err.Error()
		// Set status to abort
		status = Status_ABORT
	} else {
		for k, v := range resp {
			f.Resp[k] = v
		}
		for _, k := range vertex.TransferFields {
			vertex.C.Body[k] = f.Resp[k]
		}
	}
	vertex.Status = status

	// Append to log
	c.logs.AppendLog(sagaID, VertexLog, encodeVertex(vertex))

	// Send newVertex to update chan for coordinator to update its map of sagas
	c.updateCh <- updateMsg{
		sagaID: sagaID,
		vertex: vertex,
	}
}

// ProcessC runs a Vertex's  C
func (c *Coordinator) ProcessC(sagaID string, vertex Vertex) {
	// Sanity check on vertex's status
	if vertex.Status == Status_END_C {
		return
	}
	if !(vertex.Status == Status_START_T || vertex.Status == Status_END_T || vertex.Status == Status_START_C) {
		panic(ErrInvalidSaga)
	}
	// If vertex status is  Status_START_T, execute Status_END_T first
	if vertex.Status == Status_START_T {
		c.ProcessT(sagaID, vertex)
		return
	}

	// Now vertex must either be Status_END_T or Status_START_C. Append to log
	data := encodeVertex(vertex)
	c.logs.AppendLog(sagaID, VertexLog, data)

	// Evaluate vertex's function
	f := vertex.C

	resp, err := HTTPReq(f.GetUrl(), f.GetMethod(), f.GetRequestId(), f.GetBody())
	status := Status_END_C
	if err != nil {
		Error.Println(err)
		// Store error to output
		f.Resp["error"] = err.Error()
		// set status to startC because did not succeed
		status = Status_START_C
	} else {
		for k, v := range resp {
			f.Resp[k] = v
		}
	}
	vertex.Status = status

	// Append to log
	c.logs.AppendLog(sagaID, VertexLog, encodeVertex(vertex))

	// Send newVertex to update chan for coordinator to update its map of sagas
	c.updateCh <- updateMsg{
		sagaID: sagaID,
		vertex: vertex,
	}
}

// Run reads from a channel to serialize updates to log and corresponding sagas
func (c *Coordinator) Run() {
	for {
		select {
		case msg := <-c.updateCh:
			sagaID := msg.sagaID
			vertex := msg.vertex

			// Check if sagas exists
			saga, ok := c.sagas[sagaID]
			if !ok {
				panic(ErrSagaIDNotFound)
			}

			// Update saga
			saga.Vertices.Set(vertex.Id, vertex)

			// Transfer fields to children and run them concurrently
			saga.dagMtx.RLock()
			defer saga.dagMtx.Unlock()

			for childID, fields := range saga.DAG[vertex.Id] {
				child, ok := saga.getVtx(childID)
				if !ok {
					panic(ErrIDNotFound)
				}

				if len(fields) > 0 {
					for _, field := range fields {
						child.T.Body[field] = vertex.T.Resp[field]
					}
					saga.Vertices.Set(childID, child)
				}

				// TODO: Run child vertex
			}

		case msg := <-c.createCh:
			saga := msg.saga

			// Check if this saga already exists in local map and requests
			if _, ok := c.sagas[saga.ID]; ok {
				panic(ErrSagaIDAlreadyExists)
			}
			if _, ok := c.requests[saga.ID]; ok {
				panic(ErrSagaIDAlreadyExists)
			}

			// Append new saga to log
			c.logs.AppendLog(saga.ID, GraphLog, encodeSaga(saga))

			// Insert new saga and request and run new saga
			c.sagas[saga.ID] = saga
			c.requests[saga.ID] = msg.replyCh

			// TODO: Run source vertices in parallel
		}
	}
}

// Cleanup removes saga coordinator persistent state
func (c *Coordinator) Cleanup() {
	c.logs.Close()
	c.logs.RemoveAll()
}
