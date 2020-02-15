package sagas

// LogStore is an interface for coordinator to store logs
type LogStore interface {
	// NewSagaID returns unique id for each saga
	NewSagaID() string

	// NewRequestID returns unique id for each request
	NewRequestID() string

	// LastIndex is used for recovery purposes
	LastIndex() uint64

	// AppendLog appends a log to the db
	AppendLog(sagaID string, logType LogType, data []byte)

	// GetLog returns a log at the specified index. Will return error if log doesn't exist
	GetLog(index uint64) (Log, error)

	// Close shuts down the db
	Close()

	// RemoveAll deletes all data from the db
	RemoveAll()
}
