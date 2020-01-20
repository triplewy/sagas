package sagas

import (
	"bytes"
	"errors"

	"github.com/triplewy/sagas/utils"
	bolt "go.etcd.io/bbolt"
)

var ErrLogIndexNotFound = errors.New("log index not found in log store")

type Status int

const (
	Start Status = iota + 1
	End
	Abort
	Comp
)

type Log struct {
	SagaID uint64
	Name   string
	Status Status
	Data   []byte
}

// Equal compares equality between two logs. For testing purposes
func (a Log) Equal(b Log) bool {
	if a.SagaID != b.SagaID {
		return false
	}
	if a.Name != b.Name {
		return false
	}
	if a.Status != b.Status {
		return false
	}
	if !bytes.Equal(a.Data, b.Data) {
		return false
	}
	return true
}

func (s *Sagas) LastIndex() uint64 {
	tx, err := s.db.Begin(false)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("logs"))
	return b.Sequence()
}

func (s *Sagas) Append(log Log) {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		index, err := b.NextSequence()
		if err != nil {
			return err
		}
		key := utils.Uint64ToBytes(index)
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

func (s *Sagas) GetLog(index uint64) (Log, error) {
	tx, err := s.db.Begin(false)
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
