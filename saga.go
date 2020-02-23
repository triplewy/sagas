package sagas

import (
	"errors"
	"fmt"
	"sync"

	"github.com/triplewy/sagas/utils"

	"github.com/google/go-cmp/cmp"
	cmap "github.com/orcaman/concurrent-map"
	"go.uber.org/atomic"
)

// Errors involving incorrect sagas
var (
	ErrIDNotFound       = errors.New("vertexID does not exist in saga")
	ErrInvalidSaga      = errors.New("saga is invalid because it breaks an invariant")
	ErrUnequivalentDAGs = errors.New("DAGs in saga are not logically equivalent")
)

// Saga is different from proto in that it uses nested map for DAG
// rather than an array of edges
type Saga struct {
	// Unique ID for each Saga
	ID string

	// map[string]Vertex
	Vertices cmap.ConcurrentMap

	// DAG only requires RWMutex rather than Concurrent Map
	// since most of time we will only be reading from it
	DAG    map[string]map[string][]string
	dagMtx *sync.RWMutex

	// atomic boolean that signifies if saga has already been marked as aborted
	aborted *atomic.Bool
}

// NewSaga creates a new saga and initializes concurrent data structures
// NOTE: THIS DOES NOT INSTANTIATE **ID** FIELD
func NewSaga(vertices map[string]Vertex, dag map[string]map[string][]string) Saga {
	vtcs := cmap.New()
	for k, vtx := range vertices {
		vtcs.Set(k, vtx)
	}

	return Saga{
		Vertices: vtcs,
		DAG:      dag,
		dagMtx:   new(sync.RWMutex),
		aborted:  atomic.NewBool(false),
	}
}

// GoString implements fmt GoString interface
func (s Saga) GoString() string {
	return fmt.Sprintf("Saga{\n\tID: %v\n\tVertices: %#v,\n\tDAG: %#v\n}", s.ID, s.Vertices.Items(), s.DAG)
}

// Equal is custom equality operator between two sagas
func (s Saga) Equal(t Saga) bool {
	if s.ID != t.ID {
		return false
	}
	if !cmp.Equal(s.DAG, t.DAG) {
		return false
	}
	if !cmp.Equal(s.Vertices.Items(), t.Vertices.Items()) {
		return false
	}
	if s.aborted.Load() != t.aborted.Load() {
		return false
	}
	return true
}

func (s Saga) getVtx(id string) (Vertex, bool) {
	value, exists := s.Vertices.Get(id)
	if !exists {
		return Vertex{}, exists
	}
	return value.(Vertex), exists
}

// SwitchDAGDirection returns the opposite direction equivalent of inputted DAG
func SwitchDAGDirection(dag map[string]map[string]struct{}) map[string]map[string]struct{} {
	result := make(map[string]map[string]struct{}, len(dag))

	for parentID, children := range dag {
		if _, ok := result[parentID]; !ok {
			result[parentID] = make(map[string]struct{})
		}
		for childID := range children {
			if _, ok := result[childID]; !ok {
				result[childID] = make(map[string]struct{})
			}
			result[childID][parentID] = struct{}{}
		}
	}

	return result
}

// CheckEquivalentDAGs checks if DAG and BottomUpDAG are logically equivalent
func CheckEquivalentDAGs(a, b map[string]map[string]struct{}) error {
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
				return ErrIDNotFound
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
				return ErrIDNotFound
			}
			if _, ok := parentChildren[childID]; !ok {
				return ErrUnequivalentDAGs
			}
		}
	}

	return nil
}

// CheckFinishedOrAbort checks if saga has finished. If no abort in the saga,
// then all vertices must have status Status_END_T to be finished.
// If abort in the saga, then all vertices except aborted and not-reached vertexes
// must have status Status_END_C to be finished.
func CheckFinishedOrAbort(saga Saga) (finished, aborted bool) {
	aborted = false
	finished = true
	finishedC := true

	saga.Vertices.IterCb(func(key string, value interface{}) {
		v := value.(Vertex)

		// If vertex has status Status_ABORT, then whole saga should be aborted
		// Edge case: If we have abort in saga and other vertices have Status_END_T, saga is not finished!
		if v.Status == Status_ABORT {
			aborted = true
		}
		// If vertex is not Status_END_T, then impossible for saga to be finished forward
		if v.Status != Status_END_T {
			finished = false
		}
		// If status is not abort, notReached, nor  Status_END_C, then impossible for saga to be finished compensating
		if !(v.Status == Status_ABORT || v.Status == Status_END_C || v.Status == Status_NOT_REACHED) {
			finishedC = false
		}
	})

	// If saga aborted, we set finished status to if saga finished compensating
	if aborted {
		finished = finishedC
	}

	return
}

// CheckValidSaga returns error if any of saga is in invalid state
func CheckValidSaga(saga Saga) error {
	saga.dagMtx.RLock()
	defer saga.dagMtx.RUnlock()

	// Get aborted from vertices rather than from saga for now
	aborted := func(saga Saga) bool {
		for tuple := range saga.Vertices.IterBuffered() {
			vtx := tuple.Val.(Vertex)
			if vtx.Status == Status_ABORT {
				return true
			}
		}
		return false
	}(saga)

	// If not aborted, saga is valid iff:
	// 1. Each vertex is either Status_NOT_REACHED, Status_START_T, Status_END_T
	// 2. Parent is not Status_END_T, then child should be Status_NOT_REACHED
	if !aborted {
		// For each vertex in the saga...
		for tuple := range saga.Vertices.IterBuffered() {
			parentID := tuple.Key
			parent := tuple.Val.(Vertex)
			// #1
			if !(parent.Status == Status_NOT_REACHED || parent.Status == Status_START_T || parent.Status == Status_END_T) {
				return ErrInvalidSaga
			}
			children, ok := saga.DAG[parentID]
			if !ok {
				return ErrIDNotFound
			}
			for childID := range children {
				child, ok := saga.getVtx(childID)
				if !ok {
					return ErrIDNotFound
				}
				// #2
				if parent.Status != Status_END_T && child.Status != Status_NOT_REACHED {
					return ErrInvalidSaga
				}
			}
		}
		return nil
	}

	// If aborted, saga is valid iff: Parent is Status_NOT_REACHED, Status_START_T, or Status_ABORT, then child should be Status_NOT_REACHED
	for tuple := range saga.Vertices.IterBuffered() {
		parentID := tuple.Key
		parent := tuple.Val.(Vertex)

		children, ok := saga.DAG[parentID]
		if !ok {
			return ErrIDNotFound
		}
		for childID := range children {
			child, ok := saga.getVtx(childID)
			if !ok {
				return ErrIDNotFound
			}
			if (parent.Status == Status_NOT_REACHED || parent.Status == Status_START_T || parent.Status == Status_ABORT) && child.Status != Status_NOT_REACHED {
				return ErrInvalidSaga
			}
		}
	}
	return nil
}

// SagaBFS finds set of all vertex ids to start processing using BFS
func SagaBFS(saga Saga) []Vertex {
	saga.dagMtx.RLock()
	defer saga.dagMtx.RUnlock()

	// Get aborted from vertices rather than from saga for now
	aborted := func(saga Saga) bool {
		for tuple := range saga.Vertices.IterBuffered() {
			vtx := tuple.Val.(Vertex)
			if vtx.Status == Status_ABORT {
				return true
			}
		}
		return false
	}(saga)

	// Get source nodes of graph
	sources := findSourceVertices(saga.DAG)

	process := make(map[string]Vertex, 0)
	var vtxID string

	// If not aborted, find top most nodes with NOT_REACHED or START_T
	if !aborted {
		// Use bfs to find the nodes with NOT_REACHED or START_T to add to process
		for len(sources) > 0 {
			vtxID, sources = sources[0], sources[1:]
			vtx, ok := saga.getVtx(vtxID)
			if !ok {
				panic(ErrIDNotFound)
			}
			// If node has NOT_REACHED or START_T, add to process, and stop traveling down current path
			if vtx.Status == Status_NOT_REACHED || vtx.Status == Status_START_T {
				process[vtxID] = vtx
				continue
			}
			// Node must have END_T status, add children to queue
			if vtx.Status == Status_END_T {
				for child := range saga.DAG[vtxID] {
					sources = append(sources, child)
				}
			}
		}
	} else {
		// If aborted, find top most nodes with END_T or START_C to add to process
		for len(sources) > 0 {
			vtxID, sources = sources[0], sources[1:]
			vtx, ok := saga.getVtx(vtxID)
			if !ok {
				panic(ErrIDNotFound)
			}
			// If NOT_REACHED or ABORTED, just continue
			if vtx.Status == Status_NOT_REACHED || vtx.Status == Status_ABORT {
				continue
			}
			// If END_T or START_C, add to process and stop travelling down current path
			if vtx.Status == Status_END_T || vtx.Status == Status_START_C {
				process[vtxID] = vtx
				continue
			}
			// Node must have END_C status, add children to queue
			if vtx.Status == Status_END_C {
				for child := range saga.DAG[vtxID] {
					sources = append(sources, child)
				}
			}
		}
	}

	var res []Vertex
	for _, vtx := range process {
		res = append(res, vtx)
	}

	return res
}

// findSourceVertices finds set of all vertex ids who have no parents.
// dagMtx MUST BE RLOCKED before calling function
func findSourceVertices(dag map[string]map[string][]string) (ids []string) {
	// Build set of vertices that are children
	childNodes := make(map[string]struct{}, 0)
	for _, children := range dag {
		for id := range children {
			childNodes[id] = struct{}{}
		}
	}

	// For all vertices, the ones not in childNodes are source nodes
	for id := range dag {
		if _, ok := childNodes[id]; !ok {
			ids = append(ids, id)
		}
	}

	return
}

type sagaPack struct {
	ID       string
	Vertices map[string]Vertex
	DAG      map[string]map[string][]string
	aborted  bool
}

func encodeSaga(saga Saga) []byte {
	sp := sagaPack{
		ID:       saga.ID,
		Vertices: make(map[string]Vertex, saga.Vertices.Count()),
		DAG:      saga.DAG,
		aborted:  saga.aborted.Load(),
	}

	saga.Vertices.IterCb(func(k string, v interface{}) {
		sp.Vertices[k] = v.(Vertex)
	})

	buf, err := utils.EncodeMsgPack(sp)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func decodeSaga(data []byte) Saga {
	var sp sagaPack

	err := utils.DecodeMsgPack(data, &sp)
	if err != nil {
		panic(err)
	}

	vtxs := cmap.New()

	for key, value := range sp.Vertices {
		vtxs.Set(key, value)
	}

	return Saga{
		ID:       sp.ID,
		Vertices: vtxs,
		DAG:      sp.DAG,
		dagMtx:   new(sync.RWMutex),
		aborted:  atomic.NewBool(sp.aborted),
	}
}

func encodeVertex(vertex Vertex) []byte {
	buf, err := utils.EncodeMsgPack(vertex)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func decodeVertex(data []byte) Vertex {
	var vertex Vertex
	err := utils.DecodeMsgPack(data, &vertex)
	if err != nil {
		panic(err)
	}
	return vertex
}
