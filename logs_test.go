package sagas

import (
	"errors"
	"sync"
	"testing"

	"gotest.tools/assert"
)

func TestLog(t *testing.T) {
	config := DefaultConfig()
	s := NewCoordinator(config)
	defer s.Cleanup()

	t.Run("append", func(t *testing.T) {
		tests := []struct {
			name    string
			data    interface{}
			logType LogType
		}{
			{
				name:    "init",
				data:    []byte{0},
				logType: Init,
			},
			{
				name:    "vertex",
				data:    SagaVertex{VertexID: 0, Status: NotReached},
				logType: Vertex,
			},
			{
				name: "graph",
				data: Saga{
					DAG:      map[VertexID]map[VertexID]SagaEdge{1: {}},
					Vertices: map[VertexID]SagaVertex{1: SagaVertex{VertexID: 0, Status: NotReached}},
				},
				logType: Graph,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				sagaID := s.NewSagaID()
				data, err := func() ([]byte, error) {
					switch tt.logType {
					case Graph:
						saga, ok := tt.data.(Saga)
						if !ok {
							return nil, errors.New("data does not match expected LogType")
						}
						return encodeSaga(saga), nil
					case Vertex:
						vertex, ok := tt.data.(SagaVertex)
						if !ok {
							return nil, errors.New("data does not match expected LogType")
						}
						return encodeSagaVertex(vertex), nil
					case Init:
						return tt.data.([]byte), nil
					default:
						return nil, errors.New("invalid LogType")
					}
				}()
				assert.NilError(t, err)
				s.AppendLog(sagaID, tt.logType, data)
				index := s.LastIndex()
				log, err := s.GetLog(index)
				assert.NilError(t, err)
				assert.Equal(t, sagaID, log.SagaID)
				assert.Equal(t, tt.logType, log.LogType)
				switch tt.logType {
				case Graph:
					saga := decodeSaga(log.Data)
					assert.DeepEqual(t, tt.data.(Saga), saga)
				case Vertex:
					vertex := decodeSagaVertex(log.Data)
					assert.DeepEqual(t, tt.data.(SagaVertex), vertex)
				case Init:
					assert.DeepEqual(t, log.Data, []byte{0})
				default:
					t.Fatal("invalid LogType")
				}
			})
		}
	})

	t.Run("concurrent sagaIDs", func(t *testing.T) {
		var wg sync.WaitGroup
		ch := make(chan uint64, 100)
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				ch <- s.NewSagaID()
				wg.Done()
			}()
		}
		wg.Wait()
		close(ch)

		set := make(map[uint64]struct{}, 100)
		for id := range ch {
			if _, ok := set[id]; ok {
				t.Fatal("duplicate sagaID")
			}
			set[id] = struct{}{}
		}
	})

	t.Run("concurrent appends", func(t *testing.T) {
		logType := Vertex
		vertex := SagaVertex{VertexID: 0, Status: NotReached}
		data := encodeSagaVertex(vertex)
		startIndex := s.LastIndex()
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				sagaID := s.NewSagaID()
				s.AppendLog(sagaID, logType, data)
				wg.Done()
			}()
		}
		wg.Wait()

		endIndex := s.LastIndex()
		assert.Equal(t, startIndex+100, endIndex)

		for i := startIndex + 1; i <= endIndex; i++ {
			log, err := s.GetLog(i)
			assert.NilError(t, err)
			logVertex := decodeSagaVertex(log.Data)
			assert.DeepEqual(t, vertex, logVertex)
		}
	})
}
