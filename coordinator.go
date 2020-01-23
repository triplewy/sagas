package sagas

import (
	"fmt"
	"os"

	"github.com/triplewy/sagas/utils"

	"github.com/triplewy/sagas/hotels"
	bolt "go.etcd.io/bbolt"
)

type Coordinator struct {
	Config *Config
	db     *bolt.DB
	sagas  map[uint64]Saga

	updateCh chan Log

	hotelsClient hotels.HotelsClient
}

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

func (c *Coordinator) Recover() {
	// Repopulate all sagas into memory
	for i := uint64(1); i <= c.LastIndex(); i++ {
		log, err := c.GetLog(i)
		if err != nil {
			panic(err)
		}
		switch log.LogType {
		case Graph:
			if _, ok := c.sagas[log.SagaID]; ok {
				panic("multiple graphs with same sagaID")
			}
			var saga Saga
			err := utils.DecodeMsgPack(log.Data, &saga)
			if err != nil {
				panic(err)
			}
			c.sagas[log.SagaID] = saga
		case Vertex:
			saga, ok := c.sagas[log.SagaID]
			if !ok {
				panic("log of vertex has sagaID that does not exist")
			}
			var vertex SagaVertex
			err := utils.DecodeMsgPack(log.Data, &vertex)
			if err != nil {
				panic(err)
			}
			if _, ok := saga.Vertices[vertex.VertexID]; !ok {
				panic("vertexID does not exist in saga")
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

func (c *Coordinator) RollbackSaga(saga Saga) {

}

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
				panic("vertexID does not exist in saga")
			}
			if parent.Status != EndT {

			}
		}
	}
}

func (c *Coordinator) Run() {
	for {
		select {
		case log := <-c.updateCh:
			fmt.Println(log)
		}
	}
}

// // Call BookRoom
// // Log Start of transaction to log
// c.Append(Log{
// 	SagaID: id,
// 	Name:   "hotel",
// 	Status: Start,
// 	Data:   []byte{},
// })
// // Call remote rpc to book hotel
// reservationID, err := hotels.BookRoom(c.hotelsClient, userID, roomID)
// if err != nil {
// 	// if err, log Abort to log
// 	c.Append(Log{
// 		SagaID: id,
// 		Name:   "hotel",
// 		Status: Abort,
// 		Data:   []byte(err.Error()),
// 	})
// } else {
// 	// No err so log End to log
// 	c.Append(Log{
// 		SagaID: id,
// 		Name:   "hotel",
// 		Status: End,
// 		Data:   []byte(reservationID), // Need to store reservationID in data in case of comp transaction later on
// 	})
// }

// // Log end to log
// c.Append(Log{
// 	SagaID: id,
// 	Name:   "saga",
// 	Status: End,
// 	Data:   []byte{},
// })

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
