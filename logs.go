package sagas

import (
	"errors"

	"github.com/triplewy/sagas/utils"
	bolt "go.etcd.io/bbolt"
)

var ErrLogIndexNotFound = errors.New("log index not found in log store")

type Log struct {
	Lsn    uint64
	SagaID uint64
	Saga   Saga
}

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
			Lsn:    0,
			SagaID: 0,
			Saga:   Saga{},
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

func (c *Coordinator) LastIndex() uint64 {
	tx, err := c.db.Begin(false)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("logs"))
	return b.Sequence()
}

func (c *Coordinator) AppendLog(sagaID uint64, saga Saga) {
	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		index, err := b.NextSequence()
		if err != nil {
			return err
		}
		key := utils.Uint64ToBytes(index)
		log := Log{
			Lsn:    index,
			SagaID: sagaID,
			Saga:   saga,
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
