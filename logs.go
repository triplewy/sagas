package sagas

import (
	"errors"
	"fmt"

	"github.com/triplewy/sagas/utils"
	bolt "go.etcd.io/bbolt"
)

// ErrLogIndexNotFound is used when a log does not exist at a specified index
var ErrLogIndexNotFound = errors.New("log index not found in log store")

// LogType is used for encoding and decoding structs into byte slices
type LogType int

// Enum for LogType
const (
	Init LogType = iota + 1
	Graph
	Vertex
)

// GoString implements fmt GoString interface
func (t LogType) GoString() string {
	switch t {
	case Init:
		return "Init"
	case Graph:
		return "Graph"
	case Vertex:
		return "Vertex"
	default:
		return "Unknown"
	}
}

// Log is stored on persistent disk to keep track of sagas
type Log struct {
	Lsn     uint64
	SagaID  uint64
	LogType LogType
	Data    []byte
}

// GoString implements fmt GoString interface
func (log Log) GoString() string {
	data := func() string {
		switch log.LogType {
		case Init:
			return "{}"
		case Graph:
			saga := decodeSaga(log.Data)
			return saga.GoString()
		case Vertex:
			vertex := decodeSagaVertex(log.Data)
			return vertex.GoString()
		default:
			return "unknown data"
		}
	}()
	return fmt.Sprintf("Log{\n\tLsn: %v,\n\tSagaID: %v,\n\tLogType: %#v,\n\tData: %v\n}", log.Lsn, log.SagaID, log.LogType, data)
}

// OpenDB starts the DB process and makes sure there is an init log at index 0
func OpenDB(path string) *bolt.DB {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("state"))
		if err != nil {
			return err
		}
		b, err := tx.CreateBucketIfNotExists([]byte("logs"))
		if err != nil {
			return err
		}
		if b.Sequence() > 0 {
			return nil
		}
		key := utils.Uint64ToBytes(0)
		initLog := Log{
			Lsn:     0,
			SagaID:  0,
			LogType: Init,
			Data:    []byte{0},
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
	return db
}

// NewSagaID retrieves a unique saga ID by incrementing
func (c *Coordinator) NewSagaID() (sagaID uint64) {
	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("state"))
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		sagaID = id
		return nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// LastIndex returns the last written log index
func (c *Coordinator) LastIndex() uint64 {
	tx, err := c.db.Begin(false)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("logs"))
	return b.Sequence()
}

// AppendLog takes a sagaID, LogType, and a slice of bytes and formats them into a log to persist to disk
func (c *Coordinator) AppendLog(sagaID uint64, logType LogType, data []byte) {
	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		index, err := b.NextSequence()
		if err != nil {
			return err
		}
		key := utils.Uint64ToBytes(index)
		log := Log{
			Lsn:     index,
			SagaID:  sagaID,
			LogType: logType,
			Data:    data,
		}
		buf, err := utils.EncodeMsgPack(log)
		if err != nil {
			return err
		}
		return b.Put(key, buf.Bytes())
	})
	if err != nil {
		panic(err)
	}
}

// GetLog retrieves a log from the db. If a log does not exist at the index, GetLog returns ErrLogIndexNotFound
func (c *Coordinator) GetLog(index uint64) (Log, error) {
	tx, err := c.db.Begin(false)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("logs"))
	key := utils.Uint64ToBytes(index)
	value := b.Get(key)
	if value == nil {
		return Log{}, ErrLogIndexNotFound
	}
	var log Log
	err = utils.DecodeMsgPack(value, &log)
	if err != nil {
		panic(err)
	}
	return log, nil
}
