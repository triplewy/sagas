package sagas

import (
	"testing"

	"gotest.tools/assert"
)

func TestSaga(t *testing.T) {
	t.Run("equivalent dags", func(t *testing.T) {
		tests := []struct {
			name string
			dag  map[VertexID]map[VertexID]struct{}
		}{
			{
				name: "1 vertex",
				dag: map[VertexID]map[VertexID]struct{}{
					1: {},
				},
			},
			{
				name: "2 vertices parallel",
				dag: map[VertexID]map[VertexID]struct{}{
					1: {},
					2: {},
				},
			},
			{
				name: "2 vertices sequential",
				dag: map[VertexID]map[VertexID]struct{}{
					1: map[VertexID]struct{}{2: struct{}{}},
					2: {},
				},
			},
			{
				name: "3 vertices parallel",
				dag: map[VertexID]map[VertexID]struct{}{
					1: {},
					2: {},
					3: {},
				},
			},
			{
				name: "3 vertices sequential",
				dag: map[VertexID]map[VertexID]struct{}{
					1: map[VertexID]struct{}{2: struct{}{}},
					2: map[VertexID]struct{}{3: struct{}{}},
					3: {},
				},
			},
			{
				name: "3 vertices top peak",
				dag: map[VertexID]map[VertexID]struct{}{
					1: map[VertexID]struct{}{2: struct{}{}, 3: struct{}{}},
					2: {},
					3: {},
				},
			},
			{
				name: "3 vertices bottom peak",
				dag: map[VertexID]map[VertexID]struct{}{
					1: map[VertexID]struct{}{3: struct{}{}},
					2: map[VertexID]struct{}{3: struct{}{}},
					3: {},
				},
			},
			{
				name: "3 vertices 2 sequential 1 parallel",
				dag: map[VertexID]map[VertexID]struct{}{
					1: map[VertexID]struct{}{2: struct{}{}},
					2: {},
					3: {},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				reverseDag := SwitchGraphDirection(tt.dag)
				assert.NilError(t, CheckEquivalentDAGs(tt.dag, reverseDag))
				originalDag := SwitchGraphDirection(reverseDag)
				assert.DeepEqual(t, tt.dag, originalDag)
			})
		}
	})

	tFunc := SagaFunc{
		ID:     "t",
		Input:  map[string]interface{}{},
		Output: map[string]interface{}{},
	}

	cFunc := SagaFunc{
		ID:     "c",
		Input:  map[string]interface{}{},
		Output: map[string]interface{}{},
	}

	t.Run("finished or abort", func(t *testing.T) {
		t.Run("1 vertex", func(t *testing.T) {
			tests := []struct {
				name           string
				status         Status
				expectFinished bool
				expectAborted  bool
			}{
				{"NotReached", NotReached, false, false},
				{"StartT", StartT, false, false},
				{"EndT", EndT, true, false},
				{"Abort", Abort, true, true},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					vertices := map[VertexID]SagaVertex{1: SagaVertex{VertexID: 1, TFunc: tFunc, CFunc: cFunc, Status: tt.status}}
					finished, aborted := CheckFinishedOrAbort(vertices)
					assert.Assert(t, finished == tt.expectFinished && aborted == tt.expectAborted)
				})
			}
		})

		t.Run("2 vertices", func(t *testing.T) {
			tests := []struct {
				name           string
				status1        Status
				status2        Status
				expectFinished bool
				expectAborted  bool
			}{
				{"Abort NotReached", Abort, NotReached, false, true},
				{"Abort StartT", Abort, StartT, false, true},
				{"Abort EndT", Abort, EndT, false, true},
				{"Abort StartC", Abort, StartC, false, true},
				{"Abort EndC", Abort, EndC, true, true},
				{"Abort Abort", Abort, Abort, true, true},

				{"NotReached NotReached", NotReached, NotReached, false, false},
				{"NotReached StartT", NotReached, StartT, false, false},
				{"NotReached EndT", NotReached, EndT, false, false},

				{"StartT StartT", StartT, StartT, false, false},
				{"StartT EndT", StartT, EndT, false, false},

				{"EndT EndT", EndT, EndT, true, false},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					vertices := map[VertexID]SagaVertex{
						1: SagaVertex{VertexID: 1, TFunc: tFunc, CFunc: cFunc, Status: tt.status1},
						2: SagaVertex{VertexID: 2, TFunc: tFunc, CFunc: cFunc, Status: tt.status2},
					}
					finished, aborted := CheckFinishedOrAbort(vertices)
					assert.Assert(t, finished == tt.expectFinished && aborted == tt.expectAborted)
				})
			}
		})
	})

	t.Run("valid saga", func(t *testing.T) {
		t.Run("1 vertex", func(t *testing.T) {
			dag := map[VertexID]map[VertexID]struct{}{1: {}}

			tests := []struct {
				name        string
				status      Status
				expectError error
			}{
				{"NotReached", NotReached, nil},
				{"StartT", StartT, nil},
				{"EndT", EndT, nil},
				{"StartC", StartC, ErrInvalidSaga},
				{"EndC", EndC, ErrInvalidSaga},
				{"Abort", Abort, nil},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					vertices := map[VertexID]SagaVertex{1: SagaVertex{VertexID: 1, TFunc: tFunc, CFunc: cFunc, Status: tt.status}}
					saga := NewSaga(dag, vertices)
					_, aborted := CheckFinishedOrAbort(saga.Vertices)
					assert.Equal(t, CheckValidSaga(saga, aborted), tt.expectError)
				})
			}
		})

		t.Run("2 vertices", func(t *testing.T) {
			t.Run("parallel", func(t *testing.T) {
				dag := map[VertexID]map[VertexID]struct{}{1: {}, 2: {}}
				tests := []struct {
					name        string
					status1     Status
					status2     Status
					expectError error
				}{
					{"Abort NotReached", Abort, NotReached, nil},
					{"Abort StartT", Abort, StartT, nil},
					{"Abort EndT", Abort, EndT, nil},
					{"Abort StartC", Abort, StartC, nil},
					{"Abort EndC", Abort, EndC, nil},
					{"Abort Abort", Abort, Abort, nil},

					{"NotReached NotReached", NotReached, NotReached, nil},
					{"NotReached StartT", NotReached, StartT, nil},
					{"NotReached EndT", NotReached, EndT, nil},
					{"NotReached StartC", NotReached, StartC, ErrInvalidSaga},
					{"NotReached EndC", NotReached, EndC, ErrInvalidSaga},

					{"StartT StartT", StartT, StartT, nil},
					{"StartT EndT", StartT, EndT, nil},
					{"StartT StartC", StartT, StartC, ErrInvalidSaga},
					{"StartT EndC", StartT, EndC, ErrInvalidSaga},

					{"EndT EndT", EndT, EndT, nil},
					{"EndT StartC", EndT, StartC, ErrInvalidSaga},
					{"EndT EndC", EndT, EndC, ErrInvalidSaga},

					{"StartC StartC", StartC, StartC, ErrInvalidSaga},
					{"StartC EndC", StartC, EndC, ErrInvalidSaga},

					{"EndC EndC", EndC, EndC, ErrInvalidSaga},
				}

				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						vertices := map[VertexID]SagaVertex{
							1: SagaVertex{VertexID: 1, TFunc: tFunc, CFunc: cFunc, Status: tt.status1},
							2: SagaVertex{VertexID: 2, TFunc: tFunc, CFunc: cFunc, Status: tt.status2},
						}
						saga := NewSaga(dag, vertices)
						_, aborted := CheckFinishedOrAbort(saga.Vertices)
						assert.Equal(t, CheckValidSaga(saga, aborted), tt.expectError)
					})
				}
			})

			// Order matters
			t.Run("sequential", func(t *testing.T) {
				dag := map[VertexID]map[VertexID]struct{}{1: {2: struct{}{}}, 2: {}}
				tests := []struct {
					name        string
					status1     Status
					status2     Status
					expectError error
				}{
					{"Abort NotReached", Abort, NotReached, nil},
					{"Abort StartT", Abort, StartT, ErrInvalidSaga},
					{"Abort EndT", Abort, EndT, ErrInvalidSaga},
					{"Abort StartC", Abort, StartC, ErrInvalidSaga},
					{"Abort EndC", Abort, EndC, ErrInvalidSaga},
					{"Abort Abort", Abort, Abort, ErrInvalidSaga},

					{"NotReached NotReached", NotReached, NotReached, nil},
					{"NotReached StartT", NotReached, StartT, ErrInvalidSaga},
					{"NotReached EndT", NotReached, EndT, ErrInvalidSaga},
					{"NotReached StartC", NotReached, StartC, ErrInvalidSaga},
					{"NotReached EndC", NotReached, EndC, ErrInvalidSaga},
					{"NotReached Abort", NotReached, Abort, ErrInvalidSaga},

					{"StartT NotReached", StartT, NotReached, nil},
					{"StartT StartT", StartT, StartT, ErrInvalidSaga},
					{"StartT EndT", StartT, EndT, ErrInvalidSaga},
					{"StartT StartC", StartT, StartC, ErrInvalidSaga},
					{"StartT EndC", StartT, EndC, ErrInvalidSaga},
					{"StartT Abort", StartT, Abort, ErrInvalidSaga},

					{"EndT NotReached", EndT, NotReached, nil},
					{"EndT StartT", EndT, StartT, nil},
					{"EndT EndT", EndT, EndT, nil},
					{"EndT StartC", EndT, StartC, ErrInvalidSaga},
					{"EndT EndC", EndT, EndC, ErrInvalidSaga},
					{"EndT Abort", EndT, Abort, nil},

					{"StartC NotReached", StartC, NotReached, ErrInvalidSaga},
					{"StartC StartT", StartC, StartT, ErrInvalidSaga},
					{"StartC EndT", StartC, EndT, ErrInvalidSaga},
					{"StartC StartC", StartC, StartC, ErrInvalidSaga},
					{"StartC EndC", StartC, EndC, ErrInvalidSaga},
					{"StartC Abort", StartC, Abort, nil},

					{"EndC NotReached", StartC, NotReached, ErrInvalidSaga},
					{"EndC StartT", StartC, StartT, ErrInvalidSaga},
					{"EndC EndT", StartC, EndT, ErrInvalidSaga},
					{"EndC StartC", StartC, StartC, ErrInvalidSaga},
					{"EndC EndC", StartC, EndC, ErrInvalidSaga},
					{"EndC Abort", StartC, Abort, nil},
				}

				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						vertices := map[VertexID]SagaVertex{
							1: SagaVertex{VertexID: 1, TFunc: tFunc, CFunc: cFunc, Status: tt.status1},
							2: SagaVertex{VertexID: 2, TFunc: tFunc, CFunc: cFunc, Status: tt.status2},
						}
						saga := NewSaga(dag, vertices)
						_, aborted := CheckFinishedOrAbort(saga.Vertices)
						assert.Equal(t, CheckValidSaga(saga, aborted), tt.expectError)
					})
				}
			})
		})
	})
}
