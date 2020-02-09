package sagas

import (
	"errors"
)

// Errors from incorrect coordinator logic
var (
	ErrInvalidStatus         = errors.New("SagaVertex's status is invalid in the context of the saga")
	ErrInvalidFuncID         = errors.New("invalid saga func ID")
	ErrInvalidFuncInputField = errors.New("field does not exist in input for saga func")
	ErrInvalidFuncInputType  = errors.New("incorrect type for input field in saga func")
	ErrSagaIDAlreadyExists   = errors.New("create saga's sagaID already exists")
	ErrSagaIDNotFound        = errors.New("update sagaID does not exist in coordinator's map")
)

type updateMsg struct {
	sagaID uint64
	vertex SagaVertex
}

type createMsg struct {
	sagaID  uint64
	saga    Saga
	replyCh chan Saga
}

// Coordinator handles saga requests by calling RPCs and persisting logs to disk
type Coordinator struct {
	Config   *Config
	logs     LogStore
	sagas    map[uint64]Saga
	requests map[uint64]chan Saga

	createCh chan createMsg
	updateCh chan updateMsg
}

// NewCoordinator creates a new coordinator based on a config
func NewCoordinator(config *Config, logStore LogStore) *Coordinator {
	c := &Coordinator{
		Config:   config,
		logs:     logStore,
		sagas:    make(map[uint64]Saga),
		requests: make(map[uint64]chan Saga),

		createCh: make(chan createMsg),
		updateCh: make(chan updateMsg),
	}

	go c.Run()

	if config.AutoRecover {
		c.Recover()
	}

	return c
}

// Recover reads logs from disks and reconstructs dags in memory
func (c *Coordinator) Recover() {
	// Repopulate all sagas into memory
	for i := uint64(1); i <= c.logs.LastIndex(); i++ {
		log, err := c.logs.GetLog(i)
		if err != nil {
			// Error must be ErrLogIndexNotFound so we just skip the index
			continue
		}
		switch log.LogType {
		case InitLog:
			continue
		case GraphLog:
			if _, ok := c.sagas[log.SagaID]; ok {
				panic("multiple graphs with same sagaID")
			}
			saga := decodeSaga(log.Data)
			c.sagas[log.SagaID] = saga
		case VertexLog:
			saga, ok := c.sagas[log.SagaID]
			if !ok {
				panic("log of vertex has sagaID that does not exist")
			}
			vertex := decodeSagaVertex(log.Data)
			if _, ok := saga.Vertices[vertex.VertexID]; !ok {
				panic(ErrVertexIDNotFound)
			}
			saga.Vertices[vertex.VertexID] = vertex
		default:
			panic("unrecognized log type")
		}
	}
	// Run all sagas
	for sagaID, saga := range c.sagas {
		c.RunSaga(sagaID, saga)
	}
}

// RunSaga first checks if saga is already finished. Each unfinished saga vertex then runs in its own goroutine
func (c *Coordinator) RunSaga(sagaID uint64, saga Saga) {
	// Check if saga already finished
	finished, aborted := CheckFinishedOrAbort(saga.Vertices)
	if finished {
		// Notify request that saga has finished
		if replyCh, ok := c.requests[sagaID]; ok {
			replyCh <- saga
			delete(c.requests, sagaID)
		}
		return
	}

	err := CheckValidSaga(saga, aborted)
	if err != nil {
		panic(err)
	}

	if aborted {
		c.RollbackSaga(sagaID, saga)
	} else {
		c.ContinueSaga(sagaID, saga)
	}
}

// RollbackSaga calls necessary compensating functions to undo saga
func (c *Coordinator) RollbackSaga(sagaID uint64, saga Saga) {
	// Each vertex that is not Abort, NotReached, or EndC still needs to fully rollback
	for _, vertex := range saga.Vertices {
		if !(vertex.Status == Abort || vertex.Status == EndC || vertex.Status == NotReached) {
			go c.ProcessC(sagaID, vertex)
		}
	}
}

// ContinueSaga runs saga forward
func (c *Coordinator) ContinueSaga(sagaID uint64, saga Saga) {
	// Build set of vertices that are children
	seen := make(map[VertexID]struct{}, 0)
	for _, children := range saga.DAG {
		for id := range children {
			seen[id] = struct{}{}
		}
	}

	var highest []VertexID
	// For all vertices, the ones not in seen are highest vertices
	for id := range saga.Vertices {
		isChild := func() bool {
			for seenID := range seen {
				if seenID == id {
					return true
				}
			}
			return false
		}()
		if !isChild {
			highest = append(highest, id)
		}
	}

	// Find list of vertices to process using bfs
	process := make(map[VertexID]SagaVertex)
	var id VertexID

	for len(highest) > 0 {
		// pop first vertex off of highest
		id, highest = highest[0], highest[1:]

		v, ok := saga.Vertices[id]
		if !ok {
			panic(ErrVertexIDNotFound)
		}
		// if vertex is not EndT, add to process
		if v.Status != EndT {
			process[id] = v
		} else {
			// add all children to highest
			for childID := range saga.DAG[id] {
				highest = append(highest, childID)
			}
		}
	}

	// Run each vertex to process on a separate goroutine
	for _, vertex := range process {
		go c.ProcessT(sagaID, vertex)
	}
}

// ProcessT runs a SagaVertex's TFunc
func (c *Coordinator) ProcessT(sagaID uint64, vertex SagaVertex) {
	// Sanity check on vertex's status
	if vertex.Status == EndT {
		return
	}
	if !(vertex.Status == StartT || vertex.Status == NotReached) {
		panic(ErrInvalidSaga)
	}

	// Append to log
	data := encodeSagaVertex(vertex)
	c.logs.AppendLog(sagaID, VertexLog, data)

	// Evaluate vertex's function
	f := vertex.TFunc

	resp, err := HTTPReq(f.URL, f.Method, f.RequestID, f.Body)
	status := EndT
	if err != nil {
		Error.Println(err)
		// Store error to output
		f.Resp["error"] = err.Error()
		// Set status to abort
		status = Abort
	} else {
		for k, v := range resp {
			f.Resp[k] = v
		}
		for _, k := range vertex.TransferFields {
			vertex.CFunc.Body[k] = f.Resp[k]
		}
	}
	vertex.Status = status

	// Append to log
	c.logs.AppendLog(sagaID, VertexLog, encodeSagaVertex(vertex))

	// Send newVertex to update chan for coordinator to update its map of sagas
	c.updateCh <- updateMsg{
		sagaID: sagaID,
		vertex: vertex,
	}
}

// ProcessC runs a SagaVertex's CFunc
func (c *Coordinator) ProcessC(sagaID uint64, vertex SagaVertex) {
	// Sanity check on vertex's status
	if vertex.Status == EndC {
		return
	}
	if !(vertex.Status == StartT || vertex.Status == EndT || vertex.Status == StartC) {
		panic(ErrInvalidSaga)
	}
	// If vertex status is StartT, execute EndT first
	if vertex.Status == StartT {
		c.ProcessT(sagaID, vertex)
		return
	}

	// Now vertex must either be EndT or StartC. Append to log
	data := encodeSagaVertex(vertex)
	c.logs.AppendLog(sagaID, VertexLog, data)

	// Evaluate vertex's function
	f := vertex.CFunc

	resp, err := HTTPReq(f.URL, f.Method, f.RequestID, f.Body)
	status := EndC
	if err != nil {
		Error.Println(err)
		// Store error to output
		f.Resp["error"] = err.Error()
		// set status to startC because did not succeed
		status = StartC
	} else {
		for k, v := range resp {
			f.Resp[k] = v
		}
	}
	vertex.Status = status

	// Append to log
	c.logs.AppendLog(sagaID, VertexLog, encodeSagaVertex(vertex))

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
			saga, ok := c.sagas[sagaID]
			if !ok {
				panic(ErrSagaIDNotFound)
			}
			saga.Vertices[vertex.VertexID] = vertex
			for childID, edge := range saga.DAG[vertex.VertexID] {
				child, ok := saga.Vertices[childID]
				if !ok {
					panic(ErrVertexIDNotFound)
				}
				for _, field := range edge.Fields {
					child.TFunc.Body[field] = vertex.TFunc.Resp[field]
				}
			}
			c.RunSaga(sagaID, saga)
		case msg := <-c.createCh:
			sagaID := msg.sagaID
			saga := msg.saga
			if _, ok := c.sagas[sagaID]; ok {
				panic(ErrSagaIDAlreadyExists)
			}
			if _, ok := c.requests[sagaID]; ok {
				panic(ErrSagaIDAlreadyExists)
			}
			// Append new saga to log
			c.logs.AppendLog(sagaID, GraphLog, encodeSaga(saga))
			// Insert new saga and request and run new saga
			c.sagas[sagaID] = saga
			c.requests[sagaID] = msg.replyCh
			c.RunSaga(msg.sagaID, saga)
		}
	}
}

// Cleanup removes saga coordinator persistent state
func (c *Coordinator) Cleanup() {
	c.logs.Close()
	c.logs.RemoveAll()
}
