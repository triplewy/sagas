package sagas

import (
	"errors"
	"fmt"
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
