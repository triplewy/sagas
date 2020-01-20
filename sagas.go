package sagas

import (
	"os"

	"github.com/triplewy/sagas/hotels"
	"github.com/triplewy/sagas/utils"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/atomic"
)

type Sagas struct {
	Config *Config
	currID *atomic.Uint64
	db     *bolt.DB

	hotelsClient hotels.HotelsClient
}

func NewSagas(config *Config) *Sagas {
	db, err := bolt.Open(config.Path, 0666, nil)
	if err != nil {
		panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("logs"))
		if err != nil {
			return err
		}
		if b.Sequence() > 0 {
			return nil
		}
		key := utils.Uint64ToBytes(0)
		initLog := Log{
			SagaID: 0,
			Name:   "init",
			Status: Start,
			Data:   []byte{},
		}
		buf, err := utils.EncodeMsgPack(initLog)
		if err != nil {
			return err
		}
		return b.Put(key, buf.Bytes())
	})
	if err != nil {
		panic(err)
	}

	client := hotels.NewClient(config.HotelsAddr)

	return &Sagas{
		Config: config,
		currID: atomic.NewUint64(0),
		db:     db,

		hotelsClient: client,
	}
}

func (s *Sagas) NewID() uint64 {
	return s.currID.Inc()
}

func (s *Sagas) StartSaga(userID, roomID string) {
	// Get new saga ID
	id := s.NewID()
	// Persist start saga to log
	s.Append(Log{
		SagaID: id,
		Name:   "saga",
		Status: Start,
		Data:   []byte{},
	})
	// Call BookRoom
	// Log Start of transaction to log
	s.Append(Log{
		SagaID: id,
		Name:   "hotel",
		Status: Start,
		Data:   []byte{},
	})
	// Call remote rpc to book hotel
	reservationID, err := hotels.BookRoom(s.hotelsClient, userID, roomID)
	if err != nil {
		// if err, log Abort to log
		s.Append(Log{
			SagaID: id,
			Name:   "hotel",
			Status: Abort,
			Data:   []byte(err.Error()),
		})
	} else {
		// No err so log End to log
		s.Append(Log{
			SagaID: id,
			Name:   "hotel",
			Status: End,
			Data:   []byte(reservationID), // Need to store reservationID in data in case of comp transaction later on
		})
	}

	// Log end to log
	s.Append(Log{
		SagaID: id,
		Name:   "saga",
		Status: End,
		Data:   []byte{},
	})
}

func (s *Sagas) Cleanup() {
	err := s.db.Close()
	if err != nil {
		panic(err)
	}
	err = os.RemoveAll(s.Config.Path)
	if err != nil {
		panic(err)
	}
}
