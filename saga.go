package sagas

import (
	"errors"
	"fmt"

	"github.com/triplewy/sagas/utils"
)

// Errors involving incorrect sagas
var (
	ErrVertexIDNotFound = errors.New("vertexID does not exist in saga")
	ErrInvalidSaga      = errors.New("saga is invalid because it breaks an invariant")
	ErrUnequivalentDAGs = errors.New("DAGs in saga are not logically equivalent")
)

// VertexID is a unique id representing each vertex in each saga
type VertexID string

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

// GoString implements fmt GoString interface
func (s Status) GoString() string {
	switch s {
	case NotReached:
		return "NotReached"
	case StartT:
		return "StartT"
	case EndT:
		return "EndT"
	case StartC:
		return "StartC"
	case EndC:
		return "EndC"
	case Abort:
		return "Abort"
	default:
		return "Unknown"
	}
}

// Saga is a DAG of SagaVertices that keeps track of progress of a Saga transaction
type Saga struct {
	DAG      map[VertexID]map[VertexID]SagaEdge
	Vertices map[VertexID]SagaVertex
}

// GoString implements fmt GoString interface
func (s Saga) GoString() string {
	return fmt.Sprintf("Saga{DAG: %v, Vertices: %#v}", s.DAG, s.Vertices)
}

// SagaVertex represents each vertex in a saga graph. Each vertex has a forward and compensating SagaFunc.
// Each SagaVertex should be able to rollback a forward transaction with its own data. This means that
// rollbacks in a Saga do not have to be sequential
type SagaVertex struct {
	VertexID       VertexID
	TFunc          SagaFunc
	CFunc          SagaFunc
	TransferFields []string // fields to transfer from TFunc's resp to CFunc's body
	Status         Status
}

// GoString implements fmt GoString interface
func (v SagaVertex) GoString() string {
	return fmt.Sprintf("SagaVertex{\n\tVertexID: %v,\n\tTFunc: %#v,\n\tCFunc: %#v,\n\tStatus: %v\n}", v.VertexID, v.TFunc, v.CFunc, v.Status.GoString())
}

// SagaEdge connects two SagaVertex's and transfers data between them
type SagaEdge struct {
	Fields []string
}

// SagaFunc provides information to call a function in the Saga
type SagaFunc struct {
	URL       string
	Method    string
	RequestID string
	Body      map[string]string
	Resp      map[string]string
}

// GoString implements fmt GoString interface
func (f SagaFunc) GoString() string {
	return fmt.Sprintf("SagaFunc{\n\tURL: %v,\n\tMethod: %#v,\n\tRequestID: %v,\n\tBody: %#v, \n\tResp: %#v\n}", f.URL, f.Method, f.RequestID, f.Body, f.Resp)
}

// SwitchGraphDirection returns the opposite direction equivalent of inputted DAG
func SwitchGraphDirection(dag map[VertexID]map[VertexID]struct{}) map[VertexID]map[VertexID]struct{} {
	result := make(map[VertexID]map[VertexID]struct{}, len(dag))

	for parentID, children := range dag {
		if _, ok := result[parentID]; !ok {
			result[parentID] = make(map[VertexID]struct{})
		}
		for childID := range children {
			if _, ok := result[childID]; !ok {
				result[childID] = make(map[VertexID]struct{})
			}
			result[childID][parentID] = struct{}{}
		}
	}

	return result
}

// CheckEquivalentDAGs checks if DAG and BottomUpDAG are logically equivalent
func CheckEquivalentDAGs(a, b map[VertexID]map[VertexID]struct{}) error {
	// Check if both graphs have same amount of vertices
	if len(a) != len(b) {
		return ErrUnequivalentDAGs
	}

	// For each vertex in a...
	for parentID, children := range a {
		// For each vertex's children, check if child has vertex as part of its parents
		for childID := range children {
			childParents, ok := b[childID]
			if !ok {
				return ErrVertexIDNotFound
			}
			if _, ok := childParents[parentID]; !ok {
				return ErrUnequivalentDAGs
			}
		}
	}

	// For each vertex in b...
	for childID, parents := range b {
		// For each vertex's parents, check if parent has vertex as part of its children
		for parentID := range parents {
			parentChildren, ok := a[parentID]
			if !ok {
				return ErrVertexIDNotFound
			}
			if _, ok := parentChildren[childID]; !ok {
				return ErrUnequivalentDAGs
			}
		}
	}

	return nil
}

// CheckFinishedOrAbort checks if saga has finished or has been aborted.
// If no abort in the saga, then all vertices must have status EndT to be finished.
// If abort in the saga, then all vertices except aborted and not-reached vertexes
// must have status EndC to be finished.
func CheckFinishedOrAbort(vertices map[VertexID]SagaVertex) (finished, aborted bool) {
	finished = true
	finishedC := true
	// For each vertex in the saga...
	for _, v := range vertices {
		// If vertex has status Abort, then whole saga should be aborted
		// Edge case: If we have abort in saga and other vertices have EndT, saga is not finished!
		if v.Status == Abort {
			aborted = true
		}
		// If vertex is not EndT, then impossible for saga to be finished forward
		if v.Status != EndT {
			finished = false
		}
		// If status is not abort, notReached, nor EndC, then impossible for saga to be finished compensating
		if !(v.Status == Abort || v.Status == EndC || v.Status == NotReached) {
			finishedC = false
		}
	}
	// If saga aborted, we set finished status to if saga finished compensating
	if aborted {
		finished = finishedC
	}
	return
}

// CheckValidSaga returns error if any of saga is in invalid state
func CheckValidSaga(saga Saga, aborted bool) error {
	// If not aborted, saga is valid iff:
	// 1. Each vertex is either NotReached, StartT, EndT
	// 2. Parent is not EndT, then child should be NotReached
	if !aborted {
		// For each vertex in the saga...
		for parentID, parent := range saga.Vertices {
			// #1
			if !(parent.Status == NotReached || parent.Status == StartT || parent.Status == EndT) {
				return ErrInvalidSaga
			}
			children, ok := saga.DAG[parentID]
			if !ok {
				return ErrVertexIDNotFound
			}
			for childID := range children {
				child, ok := saga.Vertices[childID]
				if !ok {
					return ErrVertexIDNotFound
				}
				// #2
				if parent.Status != EndT && child.Status != NotReached {
					return ErrInvalidSaga
				}
			}
		}
		return nil
	}

	// If aborted, saga is valid iff: Parent is NotReached, StartT, or Abort, then child should be NotReached
	for parentID, parent := range saga.Vertices {
		children, ok := saga.DAG[parentID]
		if !ok {
			return ErrVertexIDNotFound
		}
		for childID := range children {
			child, ok := saga.Vertices[childID]
			if !ok {
				return ErrVertexIDNotFound
			}
			if (parent.Status == NotReached || parent.Status == StartT || parent.Status == Abort) && child.Status != NotReached {
				return ErrInvalidSaga
			}
		}
	}
	return nil
}

func encodeSaga(saga Saga) []byte {
	buf, err := utils.EncodeMsgPack(saga)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func decodeSaga(data []byte) Saga {
	var saga Saga
	err := utils.DecodeMsgPack(data, &saga)
	if err != nil {
		panic(err)
	}
	return saga
}

func encodeSagaVertex(vertex SagaVertex) []byte {
	buf, err := utils.EncodeMsgPack(vertex)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func decodeSagaVertex(data []byte) SagaVertex {
	var vertex SagaVertex
	err := utils.DecodeMsgPack(data, &vertex)
	if err != nil {
		panic(err)
	}
	return vertex
}
