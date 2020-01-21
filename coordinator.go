package sagas

import (
	"fmt"
	"os"

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

	if config.AutoRecover {
		c.Recover()
	}

	return c
}

func (c *Coordinator) Recover() {
	for i := uint64(1); i <= c.LastIndex(); i++ {
		log, err := c.GetLog(i)
		if err != nil {
			panic(err)
		}
		c.sagas[log.SagaID] = log.Saga
	}
}

func (c *Coordinator) StartSaga(saga Saga) {
	// Get new saga ID
	id := c.NewSagaID()
	// Persist start saga to log
	c.AppendLog(id, saga)
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
