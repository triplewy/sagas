package sagas

// LogStore is an interface for coordinator to store logs
type LogStore interface {
	NewSagaID() uint64
	NewRequestID() string
	LastIndex() uint64
	AppendLog(sagaID uint64, logType LogType, data []byte)
	GetLog(index uint64) (Log, error)
	Close()
}
