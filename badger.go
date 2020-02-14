package sagas

import (
	"os"
	"strconv"

	"github.com/dgraph-io/badger"
	"github.com/triplewy/sagas/utils"
)

// Badger implements LogStore interface
type Badger struct {
	path        string
	db          *badger.DB
	reqCounter  *badger.Sequence
	sagaCounter *badger.Sequence
	logCounter  *badger.Sequence
}

// NewBadgerDB opens an in-memory BadgerDB
func NewBadgerDB(path string, inMemory bool) *Badger {
	if inMemory {
		path = ""
	}
	opts := badger.DefaultOptions(path)
	opts.InMemory = inMemory
	opts.EventLogging = false
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	reqCounter, err := db.GetSequence([]byte("requests"), 100)
	if err != nil {
		panic(err)
	}
	sagaCounter, err := db.GetSequence([]byte("sagas"), 100)
	if err != nil {
		panic(err)
	}
	logCounter, err := db.GetSequence([]byte("logs"), 100)
	if err != nil {
		panic(err)
	}

	b := &Badger{
		path:        path,
		db:          db,
		reqCounter:  reqCounter,
		sagaCounter: sagaCounter,
		logCounter:  logCounter,
	}

	b.AppendLog(0, InitLog, []byte{0})

	return b
}

// NewSagaID retrieves a unique saga ID by incrementing
func (b *Badger) NewSagaID() (sagaID uint64) {
	var err error
	sagaID, err = b.sagaCounter.Next()
	if err != nil {
		panic(err)
	}
	return
}

// NewRequestID retrieves a unique request ID by incrementing
func (b *Badger) NewRequestID() (requestID string) {
	num, err := b.reqCounter.Next()
	if err != nil {
		panic(err)
	}
	return strconv.FormatUint(num, 10)
}

// LastIndex returns the last written log index
func (b *Badger) LastIndex() (index uint64) {
	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("log:")

		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Reverse = true
		opts.Prefix = prefix

		it := txn.NewIterator(opts)
		defer it.Close()

		it.Seek(append(prefix, 0xFF))

		if !it.Valid() {
			return nil
		}

		item := it.Item()

		key := item.KeyCopy(nil)
		index = utils.BytesToUint64(key[len(key)-8:])
		return nil
	})
	if err != nil {
		panic(err)
	}
	return
}

// AppendLog takes a sagaID, LogType, and a slice of bytes and formats them into a log to persist to disk
func (b *Badger) AppendLog(sagaID uint64, logType LogType, data []byte) {
	err := b.db.Update(func(txn *badger.Txn) error {
		index, err := b.logCounter.Next()
		if err != nil {
			return err
		}
		log := Log{
			Lsn:     index,
			SagaID:  sagaID,
			LogType: logType,
			Data:    data,
		}
		buf := encodeLog(log)

		key := append([]byte("log:"), utils.Uint64ToBytes(index)...)

		return txn.Set(key, buf)
	})
	if err != nil {
		panic(err)
	}
}

// GetLog retrieves a log from the db. If a log does not exist at the index, GetLog returns ErrLogIndexNotFound
func (b *Badger) GetLog(index uint64) (Log, error) {
	txn := b.db.NewTransaction(false)
	defer txn.Discard()

	key := append([]byte("log:"), utils.Uint64ToBytes(index)...)
	item, err := txn.Get(key)
	if err != nil {
		return Log{}, err
	}
	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		panic(err)
	}
	return decodeLog(valCopy), nil
}

// Close releases all counters and closes badgerDB
func (b *Badger) Close() {
	if err := b.reqCounter.Release(); err != nil {
		panic(err)
	}
	if err := b.sagaCounter.Release(); err != nil {
		panic(err)
	}
	if err := b.db.Close(); err != nil {
		panic(err)
	}
}

// RemoveAll removes all db data on disk
func (b *Badger) RemoveAll() {
	if b.path != "" {
		if err := os.RemoveAll(b.path); err != nil {
			panic(err)
		}
	}
}
