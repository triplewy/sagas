package sagas

import (
	"errors"
	"sync"
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
	Config *Config
	logs   LogStore

	// map[string]Saga
	sagas    map[string]Saga
	requests map[string]chan Saga

	createCh chan createMsg
	updateCh chan updateMsg

	mtx sync.Mutex
}

// NewCoordinator creates a new coordinator based on a config
func NewCoordinator(config *Config, logStore LogStore) *Coordinator {
	c := &Coordinator{
		Config:   config,
		logs:     logStore,
		sagas:    make(map[string]Saga, 0),
		requests: make(map[string]chan Saga),

		createCh: make(chan createMsg),
		updateCh: make(chan updateMsg),
	}

	go c.Run()

	if config.AutoRecover {
		sagas := Recover(c.logs)
		for _, saga := range sagas {
			c.createCh <- createMsg{saga: saga}
		}
	}

	return c
}

// Run reads from update and create channels to serialize some operations
func (c *Coordinator) Run() {
	for {
		select {
		case msg := <-c.updateCh:
			c.update(msg)
		case msg := <-c.createCh:
			c.create(msg)
		}
	}
}

func (c *Coordinator) create(msg createMsg) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	saga := msg.saga

	// Check if this saga already exists in local map and requests
	if _, ok := c.sagas[saga.ID]; ok {
		panic(ErrSagaIDAlreadyExists)
	}
	if _, ok := c.requests[saga.ID]; ok {
		panic(ErrSagaIDAlreadyExists)
	}

	// Check if saga is in a valid state
	err := CheckValidSaga(saga)
	if err != nil {
		panic(err)
	}

	// Append new saga to log
	c.logs.AppendLog(saga.ID, GraphLog, encodeSaga(saga))

	// Insert new saga and request and run new saga
	c.sagas[saga.ID] = saga
	c.requests[saga.ID] = msg.replyCh

	// Still need to check finished or aborted since recovery can create
	// in-progress or finished sagas
	finished, aborted := CheckFinishedOrAbort(saga)
	if finished {
		if replyCh, ok := c.requests[saga.ID]; ok {
			replyCh <- saga
			delete(c.requests, saga.ID)
		}
		return
	}

	// Find vertices to process
	process := SagaBFS(saga)

	// Update in memory saga for each vertex to process
	for _, vtx := range process {
		if aborted {
			// If aborted, status must be START_C
			vtx.Status = Status_START_C
		} else {
			// If not aborted, status must be START_T
			vtx.Status = Status_START_T
		}
		saga.Vertices.Set(vtx.Id, vtx)
	}

	// Update saga
	c.sagas[saga.ID] = saga

	// Run process vertices in parallel
	for _, vtx := range process {
		if aborted {
			go c.ProcessC(saga.ID, vtx)
		} else {
			go c.ProcessT(saga.ID, vtx)
		}
	}
}

func (c *Coordinator) update(msg updateMsg) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	sagaID := msg.sagaID
	vertex := msg.vertex

	// Check if sagas exists
	saga, ok := c.sagas[sagaID]
	if !ok {
		panic(ErrSagaIDNotFound)
	}

	// Update coordinator saga map
	saga.Vertices.Set(vertex.Id, vertex)
	c.sagas[sagaID] = saga

	// Check if saga is in a valid state
	err := CheckValidSaga(saga)
	if err != nil {
		panic(err)
	}

	// This operation is sequential so each new update should
	// have the latest state of the saga
	finished, aborted := CheckFinishedOrAbort(saga)

	// If saga is finished, reply to request and break
	if finished {
		// Notify request that saga has finished
		if replyCh, ok := c.requests[saga.ID]; ok {
			replyCh <- saga
			delete(c.requests, saga.ID)
		}
		return
	}

	// If saga is aborted but we have not marked it as aborted, set aborted to true
	if aborted && !saga.aborted.Load() {
		saga.aborted.Store(true)
	}

	// Transfer fields to children
	func() {
		saga.dagMtx.RLock()
		defer saga.dagMtx.RUnlock()

		for childID, fields := range saga.DAG[vertex.Id] {
			child, ok := saga.getVtx(childID)
			if !ok {
				panic(ErrIDNotFound)
			}

			// Update child fields and saga iff not aborted
			if len(fields) > 0 && !aborted {
				for _, field := range fields {
					child.T.Body[field] = vertex.T.Resp[field]
				}
				saga.Vertices.Set(childID, child)
			}
		}
	}()

	// Find vertices to process
	process := SagaBFS(saga)

	// Update in memory saga for each vertex to process
	for _, vtx := range process {
		if aborted {
			// If aborted, status must be START_C
			vtx.Status = Status_START_C
		} else {
			// If not aborted, status must be START_T
			vtx.Status = Status_START_T
		}
		saga.Vertices.Set(vtx.Id, vtx)
	}

	// Update saga
	c.sagas[saga.ID] = saga

	// Run process vertices in parallel
	for _, vtx := range process {
		if aborted {
			go c.ProcessC(saga.ID, vtx)
		} else {
			go c.ProcessT(saga.ID, vtx)
		}
	}
}

// Cleanup removes saga coordinator persistent state
func (c *Coordinator) Cleanup() {
	c.logs.Close()
	c.logs.RemoveAll()
}
