package sagas

type VertexID uint64

// Saga is a DAG of SagaVertexes that keeps track of progress of a Saga transaction
type Saga struct {
	AdjGraph map[VertexID][]VertexID
	Vertexes map[VertexID]SagaVertex
}

// SagaFunc provides information to call a function in the Saga
type SagaFunc struct {
	ID     string
	Input  map[string]interface{}
	Output map[string]interface{}
}

type Status int

const (
	Start Status = iota + 1
	End
	Abort
	Comp
)

type SagaVertex struct {
	VertexID VertexID
	TFunc    SagaFunc
	CFunc    SagaFunc
	Status   Status
}
