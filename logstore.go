package sagas

import (
	"errors"
	"sync"
)

var (
	ErrInvalidIndex = errors.New("invalid index for logs")
)

type Store struct {
	logs []Log
	sync.RWMutex
}

func NewStore() *Store {
	return &Store{logs: []Log{}}
}

func (s *Store) Append(log Log) {
	s.Lock()
	defer s.Unlock()

	s.logs = append(s.logs, log)
}

func (s *Store) Tail(index uint64) ([]Log, error) {
	s.RLock()
	defer s.RUnlock()

	if index >= uint64(len(s.logs)) || index < 0 {
		return nil, ErrInvalidIndex
	}
	logs := make([]Log, uint64(len(s.logs))-index)
	copy(logs, s.logs[index:])
	return logs, nil
}
