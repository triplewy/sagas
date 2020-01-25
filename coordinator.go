package sagas

import (
	"errors"
	"fmt"
	"os"

	"github.com/triplewy/sagas/hotels"
	bolt "go.etcd.io/bbolt"
)

// Errors from incorrect coordinator logic
var (
	ErrInvalidStatus         = errors.New("SagaVertex's status is invalid in the context of the saga")
	ErrInvalidFuncID         = errors.New("invalid saga func ID")
	ErrInvalidFuncInputField = errors.New("field does not exist in input for saga func")
	ErrInvalidFuncInputType  = errors.New("incorrect type for input field in saga func")
)

type updateMsg struct {
	sagaID uint64
	vertex SagaVertex
}

type createMsg struct {
	sagaID uint64
	saga   Saga
}

// Coordinator handles saga requests by calling RPCs and persisting logs to disk
type Coordinator struct {
	Config *Config
	db     *bolt.DB
	sagas  map[uint64]Saga

	createCh chan createMsg
	updateCh chan updateMsg

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

		createCh: make(chan createMsg),
		updateCh: make(chan updateMsg),

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
		case Init:
			continue
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
	for sagaID, saga := range c.sagas {
		c.RunSaga(sagaID, saga)
	}
}

// RunSaga first checks if saga is already finished. Each unfinished saga then runs in its own goroutine
func (c *Coordinator) RunSaga(sagaID uint64, saga Saga) {
	// Check if saga already finished
	finished, aborted := CheckFinishedOrAbort(saga.Vertices)
	if finished {
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

}

// ContinueSaga runs saga forward
func (c *Coordinator) ContinueSaga(sagaID uint64, saga Saga) {
	// Find lowest vertices
	var lowest []VertexID
	for vID, children := range saga.TopDownDAG {
		if len(children) == 0 {
			lowest = append(lowest, vID)
		}
	}

	// Find list of vertices to process using bfs
	process := make(map[VertexID]SagaVertex)
	for len(lowest) > 0 {
		// pop first vertex off of lowest
		var childID VertexID
		childID, lowest = lowest[0], lowest[1:]

		// Check if vertex is already finished
		child, ok := saga.Vertices[childID]
		if !ok {
			panic(ErrVertexIDNotFound)
		}
		if child.Status == EndT {
			continue
		}

		canProcess := true
		// Iterate through vertex parents to see if all have finished
		for parentID := range saga.BottomUpDAG[childID] {
			parent, ok := saga.Vertices[parentID]
			if !ok {
				panic(ErrVertexIDNotFound)
			}
			// If a parent has not ended, append parent to stack and set canProcess for this vertex to false
			if parent.Status != EndT {
				lowest = append(lowest, parentID)
				canProcess = false
			}
		}
		if canProcess {
			process[childID] = child
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
	c.AppendLog(sagaID, Vertex, data)

	// Evaluate vertex's function
	fn := vertex.TFunc
	switch fn.FuncID {
	case "hotel":
		value, ok := fn.Input["userID"]
		if !ok {
			panic(ErrInvalidFuncInputField)
		}
		userID, ok := value.(string)
		if !ok {
			panic(ErrInvalidFuncInputType)
		}
		value, ok = fn.Input["roomID"]
		if !ok {
			panic(ErrInvalidFuncInputField)
		}
		roomID, ok := value.(string)
		if !ok {
			panic(ErrInvalidFuncInputType)
		}
		reservationID, err := hotels.BookRoom(c.hotelsClient, userID, roomID)
		if err != nil {
			// If err, first log to stderr
			Error.Println(err)
			// Create new SagaFunc storing error to output
			output := fn.Output
			output["error"] = err.Error()
			newTFunc := SagaFunc{
				FuncID:    fn.FuncID,
				RequestID: fn.RequestID,
				Input:     fn.Input,
				Output:    output,
			}
			newVertex := SagaVertex{
				VertexID: vertex.VertexID,
				TFunc:    newTFunc,
				CFunc:    vertex.CFunc,
				Status:   Abort,
			}
			c.updateCh <- updateMsg{}
		}
	default:
		panic(ErrInvalidFuncID)
	}
}

// ProcessC runs a SagaVertex's CFunc
func (c *Coordinator) ProcessC(sagaID uint64, vertex SagaVertex) {

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
