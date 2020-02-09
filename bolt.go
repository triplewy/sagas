package sagas

import (
	"os"
	"strconv"

	"github.com/triplewy/sagas/utils"
	bolt "go.etcd.io/bbolt"
)

// Bolt implements LogStore interface
type Bolt struct {
	db   *bolt.DB
	path string
}

// NewBoltDB starts the DB process and makes sure there is an init log at index 0
func NewBoltDB(path string) *Bolt {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("requests")); err != nil {
			return err
		}
		if _, err = tx.CreateBucketIfNotExists([]byte("state")); err != nil {
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
			LogType: InitLog,
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
	return &Bolt{
		db:   db,
		path: path,
	}
}

// NewSagaID retrieves a unique saga ID by incrementing
func (b *Bolt) NewSagaID() (sagaID uint64) {
	err := b.db.Update(func(tx *bolt.Tx) error {
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

// NewRequestID retrieves a unique request ID by incrementing
func (b *Bolt) NewRequestID() (requestID string) {
	err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("requests"))
		id, err := b.NextSequence()
		if err != nil {
			return err
		}

		requestID = strconv.FormatUint(id, 10)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// LastIndex returns the last written log index
func (b *Bolt) LastIndex() uint64 {
	tx, err := b.db.Begin(false)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	bu := tx.Bucket([]byte("logs"))
	return bu.Sequence()
}

// AppendLog takes a sagaID, LogType, and a slice of bytes and formats them into a log to persist to disk
func (b *Bolt) AppendLog(sagaID uint64, logType LogType, data []byte) {
	err := b.db.Update(func(tx *bolt.Tx) error {
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
func (b *Bolt) GetLog(index uint64) (Log, error) {
	tx, err := b.db.Begin(false)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	bu := tx.Bucket([]byte("logs"))
	key := utils.Uint64ToBytes(index)
	value := bu.Get(key)
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

// Close closes boltDB and deletes data from disk
func (b *Bolt) Close() {
	if err := b.db.Close(); err != nil {
		panic(err)
	}
}

// RemoveAll removes all db data on disk
func (b *Bolt) RemoveAll() {
	if err := os.RemoveAll(b.path); err != nil {
		panic(err)
	}
}
