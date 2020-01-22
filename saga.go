package sagas

import "errors"

// Errors involving incorrect sagas
var (
	ErrVertexIDNotFound    = errors.New("vertexID does not exist in saga")
	ErrSagaInvariantBroken = errors.New("one of two invariants broken in a saga")
	ErrUnequivalentDAGs    = errors.New("DAGs in saga are not logically equivalent")
)

// VertexID is a unique id representing each vertex in each saga
type VertexID uint64

// Saga is a DAG of SagaVertexes that keeps track of progress of a Saga transaction
type Saga struct {
	TopDownDAG  map[VertexID]map[VertexID]struct{}
	BottomUpDAG map[VertexID]map[VertexID]struct{}
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

// SagaVertex represents each vertex in a saga graph. Each vertex has a forward and compensating SagaFunc
type SagaVertex struct {
	VertexID VertexID
	TFunc    SagaFunc
	CFunc    SagaFunc
	Status   Status
}

// CheckSagaEquivalentDAGs checks if TopDownDAG and BottomUpDAG are logically equivalent
func CheckSagaEquivalentDAGs(saga Saga) error {
	// For each vertex in the saga...
	for parentID := range saga.Vertexes {
		children, ok := saga.TopDownDAG[parentID]
		if !ok {
			return ErrVertexIDNotFound
		}
		// For each vertex's children, check if child has vertex as part of its parents
		for childID := range children {
			childParents, ok := saga.BottomUpDAG[childID]
			if !ok {
				return ErrVertexIDNotFound
			}
			if _, ok := childParents[parentID]; !ok {
				return ErrUnequivalentDAGs
			}
		}
	}

	// For each vertex in the saga...
	for childID := range saga.Vertexes {
		parents, ok := saga.TopDownDAG[childID]
		if !ok {
			return ErrVertexIDNotFound
		}
		// For each vertex's parents, check if parent has vertex as part of its children
		for parentID := range parents {
			parentChildren, ok := saga.TopDownDAG[parentID]
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

// CheckSagaFinishedOrAbort checks if saga has finished or has been aborted
func CheckSagaFinishedOrAbort(saga Saga) (finished, aborted bool) {
	// For each vertex in the saga...
	for _, v := range saga.Vertexes {
		// If one vertex has status Abort, then whole saga should be aborted
		if v.Status == Abort {
			aborted = true
		}
		// If one vertex has NotReached or Start status, then the saga is not finished
		if v.Status == NotReached || v.Status == StartT || v.Status == StartC {
			finished = false
		}
	}
	return
}

// CheckValidSaga panics if 1 of two invariants are broken in a non-completed saga:
// 1. If saga is in continue mode, a vertex has EndT status but one of its parents does not have EndT status
// 2. If saga is in abort mode, a vertex has EndC status but one of its children does not have EndC, Abort, or NotReached status
func CheckValidSaga(saga Saga, isContinue bool) error {
	if isContinue {
		// For each vertex in the saga...
		for childVertexID, childVertex := range saga.Vertexes {
			// Skip if vertex does not have EndT status
			if childVertex.Status != EndT {
				continue
			}
			parents, ok := saga.BottomUpDAG[childVertexID]
			if !ok {
				return ErrVertexIDNotFound
			}
			// For each vertex's parents, panic if their status is not EndT
			for parentVertexID := range parents {
				parent, ok := saga.Vertexes[parentVertexID]
				if !ok {
					return ErrVertexIDNotFound
				}
				if parent.Status != EndT {
					return ErrSagaInvariantBroken
				}
			}
		}
		return nil
	}

	// For each vertex in the saga...
	for parentVertexID, parentVertex := range saga.Vertexes {
		// Skip if vertex does not have EndC status
		if parentVertex.Status != EndC {
			continue
		}
		children, ok := saga.TopDownDAG[parentVertexID]
		if !ok {
			return ErrVertexIDNotFound
		}
		// For each vertex's children, panic if their status is not StartT, EndT, or StartC
		for childVertexID := range children {
			child, ok := saga.Vertexes[childVertexID]
			if !ok {
				return ErrVertexIDNotFound
			}
			if child.Status == StartT || child.Status == EndT || child.Status == StartC {
				return ErrSagaInvariantBroken
			}
		}
	}

	return nil
}
