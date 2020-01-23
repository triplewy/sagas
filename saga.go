package sagas

import (
	"errors"

	"github.com/google/go-cmp/cmp"
)

// Errors involving incorrect sagas
var (
	ErrVertexIDNotFound = errors.New("vertexID does not exist in saga")
	ErrInvalidSaga      = errors.New("saga is invalid because it breaks an invariant")
	ErrUnequivalentDAGs = errors.New("DAGs in saga are not logically equivalent")
)

// VertexID is a unique id representing each vertex in each saga
type VertexID uint64

// Saga is a DAG of SagaVertices that keeps track of progress of a Saga transaction
type Saga struct {
	TopDownDAG  map[VertexID]map[VertexID]struct{}
	BottomUpDAG map[VertexID]map[VertexID]struct{}
	Vertices    map[VertexID]SagaVertex
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

// SagaVertex represents each vertex in a saga graph. Each vertex has a forward and compensating SagaFunc
type SagaVertex struct {
	VertexID VertexID
	TFunc    SagaFunc
	CFunc    SagaFunc
	Status   Status
}

// NewSaga creates a new saga from a dag and map of VertexIDs to vertices
func NewSaga(dag map[VertexID]map[VertexID]struct{}, vertices map[VertexID]SagaVertex) Saga {
	reverseDag := SwitchGraphDirection(dag)

	// Just in case...
	originalDag := SwitchGraphDirection(reverseDag)
	if !cmp.Equal(dag, originalDag) {
		panic(ErrUnequivalentDAGs)
	}

	return Saga{
		TopDownDAG:  dag,
		BottomUpDAG: reverseDag,
		Vertices:    vertices,
	}
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

// CheckEquivalentDAGs checks if TopDownDAG and BottomUpDAG are logically equivalent
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
// If abort in the saga, then all vertices except aborted nodes must have status EndC to be finished.
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
		// If status is not abort and not EndC, then impossible for saga to be finished compensating
		if v.Status != Abort && v.Status != EndC {
			finishedC = false
		}
	}
	// If saga aborted, we set finished status to if saga finished compensating
	if aborted {
		finished = finishedC
	}
	return
}

// CheckValidSaga returns error if any of 3 conditions below are met:
// 1. Saga has a compensating status (StartC, EndC) without an Abort status
// 2. Child has status Abort and parent has status NotReached or StartT or Abort
// 3. Child has status other than NotReached or Abort and parent has status other than EndT
// 4. If abort, parent has StartC or EndC status and child has StartT or EndT or StartC status
func CheckValidSaga(saga Saga, aborted bool) error {
	if aborted {
		// For each vertex in the saga...
		for parentID, parent := range saga.Vertices {
			children, ok := saga.TopDownDAG[parentID]
			if !ok {
				return ErrVertexIDNotFound
			}
			// For each vertex's children...
			for childVertexID := range children {
				child, ok := saga.Vertices[childVertexID]
				if !ok {
					return ErrVertexIDNotFound
				}
				// Invariant 2
				if child.Status == Abort && (parent.Status == NotReached || parent.Status == StartT || parent.Status == Abort) {
					return ErrInvalidSaga
				}
				// Invariant 3
				if child.Status != NotReached && child.Status != Abort && parent.Status != EndT {
					return ErrInvalidSaga
				}
				// Invariant 4
				if (parent.Status == StartC || parent.Status == EndC) && (child.Status == StartT || child.Status == EndT || child.Status == StartC) {
					return ErrInvalidSaga
				}
			}
		}
		return nil
	}

	// For each vertex in the saga...
	for childID, child := range saga.Vertices {
		// Invariant 1
		if child.Status == StartC || child.Status == EndC {
			return ErrInvalidSaga
		}
		parents, ok := saga.BottomUpDAG[childID]
		if !ok {
			return ErrVertexIDNotFound
		}
		// For each vertex's parents...
		for parentID := range parents {
			parent, ok := saga.Vertices[parentID]
			if !ok {
				return ErrVertexIDNotFound
			}
			// Invariant 2
			if child.Status != NotReached && parent.Status != EndT {
				return ErrInvalidSaga
			}
		}
	}

	return nil
}
