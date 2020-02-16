package sagas

import (
	"errors"

	cmap "github.com/orcaman/concurrent-map"
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
	sagas    cmap.ConcurrentMap
	requests map[string]chan Saga

	createCh chan createMsg
	updateCh chan updateMsg
}

// NewCoordinator creates a new coordinator based on a config
func NewCoordinator(config *Config, logStore LogStore) *Coordinator {
	c := &Coordinator{
		Config:   config,
		logs:     logStore,
		sagas:    cmap.New(),
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

func (c *Coordinator) getSaga(id string) (Saga, bool) {
	val, exists := c.sagas.Get(id)
	if !exists {
		return Saga{}, exists
	}
	return val.(Saga), exists
}

// Run reads from update and create channels to serialize some operations
func (c *Coordinator) Run() {
	for {
		select {
		case msg := <-c.updateCh:
			sagaID := msg.sagaID
			vertex := msg.vertex

			// Check if sagas exists
			saga, ok := c.getSaga(sagaID)
			if !ok {
				panic(ErrSagaIDNotFound)
			}

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
				break
			}

			// If saga is aborted but we have not marked it as aborted,
			// set aborted to true and start compensating from source nodes
			if aborted && !saga.aborted.Load() {
				saga.aborted.Store(true)

				saga.dagMtx.RLock()
				sources := FindSourceVertices(saga.DAG)
				saga.dagMtx.RUnlock()

				for _, id := range sources {
					vtx, ok := saga.getVtx(id)
					if !ok {
						panic(ErrIDNotFound)
					}
					if !(vtx.Status == Status_NOT_REACHED || vtx.Status == Status_ABORT) {
						go c.ProcessC(saga.ID, vtx)
					}
				}

				break
			}

			// Transfer fields to children and run them concurrently
			children := func() (children []Vertex) {
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

					if !aborted {
						children = append(children, child)
					} else {
						// Only compensate if has not aborted and started
						if !(child.Status == Status_NOT_REACHED || child.Status == Status_ABORT) {
							children = append(children, child)
						}
					}
				}
				return
			}()

			// Run each child vertex
			for _, vtx := range children {
				if aborted {
					go c.ProcessC(saga.ID, vtx)
				} else {
					go c.ProcessT(saga.ID, vtx)
				}
			}

		case msg := <-c.createCh:
			saga := msg.saga

			// Check if this saga already exists in local map and requests
			if _, ok := c.getSaga(saga.ID); ok {
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
			c.sagas.Set(saga.ID, saga)
			c.requests[saga.ID] = msg.replyCh

			// Still need to check finished or aborted since recovery can create
			// in-progress or finished sagas
			finished, aborted := CheckFinishedOrAbort(saga)
			if finished {
				if replyCh, ok := c.requests[saga.ID]; ok {
					replyCh <- saga
					delete(c.requests, saga.ID)
				}
				break
			}

			// Find source vertices in dag
			saga.dagMtx.RLock()
			sources := FindSourceVertices(saga.DAG)
			saga.dagMtx.RUnlock()

			// Run source vertices in parallel
			for _, id := range sources {
				vtx, ok := saga.getVtx(id)
				if !ok {
					panic(ErrIDNotFound)
				}
				if aborted {
					go c.ProcessC(saga.ID, vtx)
				} else {
					go c.ProcessT(saga.ID, vtx)
				}
			}
		}
	}
}

// Cleanup removes saga coordinator persistent state
func (c *Coordinator) Cleanup() {
	c.logs.Close()
	c.logs.RemoveAll()
}
