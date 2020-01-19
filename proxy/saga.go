package proxy

type Sagas struct {
	currID uint64
}

type Status int

// Enum for types of Statuses
const (
	InProgress Status = iota + 1
	Completed
	Abort
)

type SagaVertex struct {
	name                string
	request             string
	compensatingRequest string
	children            []SagaVertex
	status              Status
}

func NewSaga(root SagaVertex) error {
	return nil
}

func StartSaga() (uint64, error) {
	return 0, nil
}
