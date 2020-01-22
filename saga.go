package sagas

type VertexID uint64

// Saga is a DAG of SagaVertexes that keeps track of progress of a Saga transaction
type Saga struct {
	TopDownDAG  map[VertexID][]VertexID
	BottomUpDAG map[VertexID][]VertexID
	Vertexes    map[VertexID]SagaVertex
}

// SagaFunc provides information to call a function in the Saga
type SagaFunc struct {
	ID     string
	Input  map[string]interface{}
	Output map[string]interface{}
}

// Status is a possible condition of a transaction in a saga
type Status int

// Types of transaction statuses
const (
	NotReached Status = iota + 1
	StartT
	EndT // End of forward transaction
	StartC
	EndC // End of compensating transaction
	Abort
)

type SagaVertex struct {
	VertexID VertexID
	TFunc    SagaFunc
	CFunc    SagaFunc
	Status   Status
	Children []VertexID
	Parents  []VertexID
}
