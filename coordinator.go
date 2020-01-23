package sagas

import (
	"fmt"
	"os"

	"github.com/triplewy/sagas/hotels"
	bolt "go.etcd.io/bbolt"
)

// Coordinator handles saga requests by calling RPCs and persisting logs to disk
type Coordinator struct {
	Config *Config
	db     *bolt.DB
	sagas  map[uint64]Saga

	updateCh chan Log

	hotelsClient hotels.HotelsClient
}

// NewCoordinator creates a new coordinator based on a config
func NewCoordinator(config *Config) *Coordinator {
	db := OpenDB(config.Path)
	client := hotels.NewClient(config.HotelsAddr)

	c := &Coordinator{
		Config: config,
		db:     db,
		sagas:  make(map[uint64]Saga),

		updateCh: make(chan Log),

		hotelsClient: client,
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
	for i := uint64(1); i <= c.LastIndex(); i++ {
		log, err := c.GetLog(i)
		if err != nil {
			// Error must be ErrLogIndexNotFound so we just skip the index
			continue
		}
		switch log.LogType {
		case Graph:
			if _, ok := c.sagas[log.SagaID]; ok {
				panic("multiple graphs with same sagaID")
			}
			saga := decodeSaga(log.Data)
			c.sagas[log.SagaID] = saga
		case Vertex:
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
	for _, saga := range c.sagas {
		c.RunSaga(saga)
	}
}

// RunSaga first checks if saga is already finished. Each unfinished saga then runs in its own goroutine
func (c *Coordinator) RunSaga(saga Saga) {
	// Check if saga already finished
	finished, aborted := CheckFinishedOrAbort(saga.Vertices)
	if finished {
		return
	}

	if aborted {
		c.RollbackSaga(saga)
	} else {
		c.ContinueSaga(saga)
	}

}

// RollbackSaga calls necessary compensating functions to undo saga
func (c *Coordinator) RollbackSaga(saga Saga) {

}

// ContinueSaga runs saga forward
func (c *Coordinator) ContinueSaga(saga Saga) {
	// Find lowest vertices
	var lowest []VertexID
	for vID, children := range saga.TopDownDAG {
		if len(children) == 0 {
			lowest = append(lowest, vID)
		}
	}

	// Find list of vertices to process using bfs
	// var process []VertexID
	for len(lowest) > 0 {
		// pop first vertex off of lowest
		var v VertexID
		v, lowest = lowest[0], lowest[1:]
		// Iterate through vertex parents to see if all have finished
		for parentID := range saga.BottomUpDAG[v] {
			parent, ok := saga.Vertices[parentID]
			if !ok {
				panic(ErrVertexIDNotFound)
			}
			if parent.Status != EndT {

			}
		}
	}
}

// Run reads from a channel to serialize updates to log and corresponding sagas
func (c *Coordinator) Run() {
	for {
		select {
		case log := <-c.updateCh:
			fmt.Println(log)
		}
	}
}

// Cleanup removes saga coordinator persistent state
func (c *Coordinator) Cleanup() {
	err := c.db.Close()
	if err != nil {
		panic(err)
	}
	err = os.RemoveAll(c.Config.Path)
	if err != nil {
		panic(err)
	}
}
