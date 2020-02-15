package sagas

import (
	"errors"
	"sync"
	"testing"

	"gotest.tools/assert"

	cmap "github.com/orcaman/concurrent-map"

	"go.uber.org/atomic"
)

func TestLogStore(t *testing.T) {
	config := DefaultConfig()

	badger := LogStore(NewBadgerDB(config.Path, true))

	tests := []struct {
		name  string
		store LogStore
	}{
		{
			name:  "badger",
			store: badger,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.store
			defer store.Close()
			defer store.RemoveAll()

			t.Run("sagaID", func(t *testing.T) {
				m := cmap.New()
				dup := atomic.NewBool(false)

				var wg sync.WaitGroup

				for i := 0; i < 500; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						id := store.NewSagaID()
						ok := m.SetIfAbsent(id, struct{}{})
						if !ok {
							dup.Store(true)
						}
					}()
				}

				wg.Wait()
				assert.Assert(t, !dup.Load(), "Detected overlapping sagaIds")
			})

			t.Run("requestID", func(t *testing.T) {
				m := cmap.New()
				dup := atomic.NewBool(false)

				var wg sync.WaitGroup

				for i := 0; i < 500; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						id := store.NewRequestID()
						ok := m.SetIfAbsent(id, struct{}{})
						if !ok {
							dup.Store(true)
						}
					}()
				}

				wg.Wait()
				assert.Assert(t, !dup.Load(), "Detected overlapping requestIDs")
			})

			t.Run("sequential", func(t *testing.T) {
				tests := []struct {
					name    string
					data    interface{}
					logType LogType
				}{
					{
						name:    "init",
						data:    []byte{0},
						logType: InitLog,
					},
					{
						name:    "vertex",
						data:    Vertex{Id: "0", Status: Status_NOT_REACHED},
						logType: VertexLog,
					},
					{
						name:    "graph",
						data:    NewSaga(map[string]Vertex{"1": Vertex{Id: "0", Status: Status_NOT_REACHED}}, map[string]map[string][]string{"1": {}}),
						logType: GraphLog,
					},
				}

				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						sagaID := store.NewSagaID()
						data, err := func() ([]byte, error) {
							switch tt.logType {
							case GraphLog:
								saga, ok := tt.data.(Saga)
								if !ok {
									return nil, errors.New("data does not match expected LogType")
								}
								return encodeSaga(saga), nil
							case VertexLog:
								vertex, ok := tt.data.(Vertex)
								if !ok {
									return nil, errors.New("data does not match expected LogType")
								}
								return encodeVertex(vertex), nil
							case InitLog:
								return tt.data.([]byte), nil
							default:
								return nil, errors.New("invalid LogType")
							}
						}()
						assert.NilError(t, err)

						store.AppendLog(sagaID, tt.logType, data)
						index := store.LastIndex()
						log, err := store.GetLog(index)

						assert.NilError(t, err)
						assert.Equal(t, sagaID, log.SagaID)
						assert.Equal(t, tt.logType, log.LogType)

						switch tt.logType {
						case GraphLog:
							saga := decodeSaga(log.Data)
							assert.DeepEqual(t, tt.data.(Saga), saga)
						case VertexLog:
							vertex := decodeVertex(log.Data)
							assert.DeepEqual(t, tt.data.(Vertex), vertex)
						case InitLog:
							assert.DeepEqual(t, log.Data, []byte{0})
						default:
							t.Fatal("invalid LogType")
						}
					})
				}
			})

			t.Run("concurrent", func(t *testing.T) {
				logType := VertexLog
				vertex := Vertex{Id: "0", Status: Status_NOT_REACHED}
				data := encodeVertex(vertex)
				startIndex := store.LastIndex()

				var wg sync.WaitGroup

				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						sagaID := store.NewSagaID()
						store.AppendLog(sagaID, logType, data)
					}()
				}
				wg.Wait()

				endIndex := store.LastIndex()
				assert.Equal(t, startIndex+100, endIndex)

				for i := startIndex + 1; i <= endIndex; i++ {
					log, err := store.GetLog(i)
					assert.NilError(t, err)
					logVertex := decodeVertex(log.Data)
					assert.DeepEqual(t, vertex, logVertex)
				}
			})
		})
	}

}
