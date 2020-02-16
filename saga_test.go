package sagas

import (
	"sort"
	"testing"

	"gotest.tools/assert"
)

func TestSaga(t *testing.T) {
	t.Run("equivalent dags", func(t *testing.T) {
		tests := []struct {
			name string
			dag  map[string]map[string]struct{}
		}{
			{
				name: "1 vertex",
				dag: map[string]map[string]struct{}{
					"1": {},
				},
			},
			{
				name: "2 vertices parallel",
				dag: map[string]map[string]struct{}{
					"1": {},
					"2": {},
				},
			},
			{
				name: "2 vertices sequential",
				dag: map[string]map[string]struct{}{
					"1": map[string]struct{}{"2": struct{}{}},
					"2": {},
				},
			},
			{
				name: "3 vertices parallel",
				dag: map[string]map[string]struct{}{
					"1": {},
					"2": {},
					"3": {},
				},
			},
			{
				name: "3 vertices sequential",
				dag: map[string]map[string]struct{}{
					"1": map[string]struct{}{"2": struct{}{}},
					"2": map[string]struct{}{"3": struct{}{}},
					"3": {},
				},
			},
			{
				name: "3 vertices top peak",
				dag: map[string]map[string]struct{}{
					"1": map[string]struct{}{"2": struct{}{}, "3": struct{}{}},
					"2": {},
					"3": {},
				},
			},
			{
				name: "3 vertices bottom peak",
				dag: map[string]map[string]struct{}{
					"1": map[string]struct{}{"3": struct{}{}},
					"2": map[string]struct{}{"3": struct{}{}},
					"3": {},
				},
			},
			{
				name: "3 vertices 2 sequential 1 parallel",
				dag: map[string]map[string]struct{}{
					"1": map[string]struct{}{"2": struct{}{}},
					"2": {},
					"3": {},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				reverseDag := SwitchDAGDirection(tt.dag)
				assert.NilError(t, CheckEquivalentDAGs(tt.dag, reverseDag))
				originalDag := SwitchDAGDirection(reverseDag)
				assert.DeepEqual(t, tt.dag, originalDag)
			})
		}
	})

	t.Run("FindSourceVertices", func(t *testing.T) {
		tests := []struct {
			name string
			dag  map[string]map[string][]string
			ids  []string
		}{
			{
				"1 vertex",
				map[string]map[string][]string{"1": {}},
				[]string{"1"},
			},
			{
				"2 vertices parallel",
				map[string]map[string][]string{"1": {}, "2": {}},
				[]string{"1", "2"},
			},
			{
				"2 vertices sequential",
				map[string]map[string][]string{"1": {"2": nil}, "2": {}},
				[]string{"1"},
			},
			{
				"3 vertices 1 source",
				map[string]map[string][]string{"1": {"2": nil, "3": nil}, "2": {}, "3": {}},
				[]string{"1"},
			},
			{
				"3 vertices 2 source",
				map[string]map[string][]string{"1": {"3": nil}, "2": {"3": nil}, "3": {}},
				[]string{"1", "2"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ids := FindSourceVertices(tt.dag)
				sort.Strings(ids)
				sort.Strings(tt.ids)
				assert.DeepEqual(t, ids, tt.ids)
			})
		}
	})

	t.Run("finished or abort", func(t *testing.T) {
		t.Run("1 vertex", func(t *testing.T) {
			tests := []struct {
				name           string
				status         Status
				expectFinished bool
				expectAborted  bool
			}{
				{"NOT_REACHED", Status_NOT_REACHED, false, false},
				{"START_T", Status_START_T, false, false},
				{"END_T", Status_END_T, true, false},
				{"ABORT", Status_ABORT, true, true},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					vertices := map[string]Vertex{"1": Vertex{Id: "1", Status: tt.status}}
					saga := NewSaga(vertices, nil)
					finished, aborted := CheckFinishedOrAbort(saga)
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
				{"ABORT NOT_REACHED", Status_ABORT, Status_NOT_REACHED, true, true},
				{"ABORT START_T", Status_ABORT, Status_START_T, false, true},
				{"ABORT END_T", Status_ABORT, Status_END_T, false, true},
				{"ABORT START_C", Status_ABORT, Status_START_C, false, true},
				{"ABORT END_C", Status_ABORT, Status_END_C, true, true},
				{"ABORT ABORT", Status_ABORT, Status_ABORT, true, true},

				{"NOT_REACHED NOT_REACHED", Status_NOT_REACHED, Status_NOT_REACHED, false, false},
				{"NOT_REACHED START_T", Status_NOT_REACHED, Status_START_T, false, false},
				{"NOT_REACHED END_T", Status_NOT_REACHED, Status_END_T, false, false},

				{"START_T START_T", Status_START_T, Status_START_T, false, false},
				{"START_T END_T", Status_START_T, Status_END_T, false, false},

				{"END_T END_T", Status_END_T, Status_END_T, true, false},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					vertices := map[string]Vertex{
						"1": Vertex{Id: "1", Status: tt.status1},
						"2": Vertex{Id: "2", Status: tt.status2},
					}
					saga := NewSaga(vertices, nil)
					finished, aborted := CheckFinishedOrAbort(saga)
					assert.Assert(t, finished == tt.expectFinished && aborted == tt.expectAborted)
				})
			}
		})
	})

	t.Run("valid saga", func(t *testing.T) {
		t.Run("1 vertex", func(t *testing.T) {
			dag := map[string]map[string][]string{"1": {}}

			tests := []struct {
				name        string
				status      Status
				expectError error
			}{
				{"NOT_REACHED", Status_NOT_REACHED, nil},
				{"START_T", Status_START_T, nil},
				{"END_T", Status_END_T, nil},
				{"START_C", Status_START_C, ErrInvalidSaga},
				{"END_C", Status_END_C, ErrInvalidSaga},
				{"ABORT", Status_ABORT, nil},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					vertices := map[string]Vertex{"1": Vertex{Id: "1", Status: tt.status}}
					saga := NewSaga(vertices, dag)
					assert.Equal(t, CheckValidSaga(saga), tt.expectError)
				})
			}
		})

		t.Run("2 vertices", func(t *testing.T) {
			t.Run("parallel", func(t *testing.T) {
				dag := map[string]map[string][]string{"1": {}, "2": {}}
				tests := []struct {
					name        string
					status1     Status
					status2     Status
					expectError error
				}{
					{"ABORT NOT_REACHED", Status_ABORT, Status_NOT_REACHED, nil},
					{"ABORT START_T", Status_ABORT, Status_START_T, nil},
					{"ABORT END_T", Status_ABORT, Status_END_T, nil},
					{"ABORT START_C", Status_ABORT, Status_START_C, nil},
					{"ABORT END_C", Status_ABORT, Status_END_C, nil},
					{"ABORT ABORT", Status_ABORT, Status_ABORT, nil},

					{"NOT_REACHED NOT_REACHED", Status_NOT_REACHED, Status_NOT_REACHED, nil},
					{"NOT_REACHED START_T", Status_NOT_REACHED, Status_START_T, nil},
					{"NOT_REACHED END_T", Status_NOT_REACHED, Status_END_T, nil},
					{"NOT_REACHED START_C", Status_NOT_REACHED, Status_START_C, ErrInvalidSaga},
					{"NOT_REACHED END_C", Status_NOT_REACHED, Status_END_C, ErrInvalidSaga},

					{"START_T START_T", Status_START_T, Status_START_T, nil},
					{"START_T END_T", Status_START_T, Status_END_T, nil},
					{"START_T START_C", Status_START_T, Status_START_C, ErrInvalidSaga},
					{"START_T END_C", Status_START_T, Status_END_C, ErrInvalidSaga},

					{"END_T END_T", Status_END_T, Status_END_T, nil},
					{"END_T START_C", Status_END_T, Status_START_C, ErrInvalidSaga},
					{"END_T END_C", Status_END_T, Status_END_C, ErrInvalidSaga},

					{"START_C START_C", Status_START_C, Status_START_C, ErrInvalidSaga},
					{"START_C END_C", Status_START_C, Status_END_C, ErrInvalidSaga},

					{"END_C END_C", Status_END_C, Status_END_C, ErrInvalidSaga},
				}

				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						vertices := map[string]Vertex{
							"1": Vertex{Id: "1", Status: tt.status1},
							"2": Vertex{Id: "2", Status: tt.status2},
						}
						saga := NewSaga(vertices, dag)
						assert.Equal(t, CheckValidSaga(saga), tt.expectError)
					})
				}
			})

			// Order matters
			t.Run("sequential", func(t *testing.T) {
				dag := map[string]map[string][]string{"1": {"2": []string{}}, "2": {}}
				tests := []struct {
					name        string
					status1     Status
					status2     Status
					expectError error
				}{
					{"ABORT NOT_REACHED", Status_ABORT, Status_NOT_REACHED, nil},
					{"ABORT START_T", Status_ABORT, Status_START_T, ErrInvalidSaga},
					{"ABORT END_T", Status_ABORT, Status_END_T, ErrInvalidSaga},
					{"ABORT START_C", Status_ABORT, Status_START_C, ErrInvalidSaga},
					{"ABORT END_C", Status_ABORT, Status_END_C, ErrInvalidSaga},
					{"ABORT ABORT", Status_ABORT, Status_ABORT, ErrInvalidSaga},

					{"NOT_REACHED NOT_REACHED", Status_NOT_REACHED, Status_NOT_REACHED, nil},
					{"NOT_REACHED START_T", Status_NOT_REACHED, Status_START_T, ErrInvalidSaga},
					{"NOT_REACHED END_T", Status_NOT_REACHED, Status_END_T, ErrInvalidSaga},
					{"NOT_REACHED START_C", Status_NOT_REACHED, Status_START_C, ErrInvalidSaga},
					{"NOT_REACHED END_C", Status_NOT_REACHED, Status_END_C, ErrInvalidSaga},
					{"NOT_REACHED ABORT", Status_NOT_REACHED, Status_ABORT, ErrInvalidSaga},

					{"START_T NOT_REACHED", Status_START_T, Status_NOT_REACHED, nil},
					{"START_T START_T", Status_START_T, Status_START_T, ErrInvalidSaga},
					{"START_T END_T", Status_START_T, Status_END_T, ErrInvalidSaga},
					{"START_T START_C", Status_START_T, Status_START_C, ErrInvalidSaga},
					{"START_T END_C", Status_START_T, Status_END_C, ErrInvalidSaga},
					{"START_T ABORT", Status_START_T, Status_ABORT, ErrInvalidSaga},

					{"END_T NOT_REACHED", Status_END_T, Status_NOT_REACHED, nil},
					{"END_T START_T", Status_END_T, Status_START_T, nil},
					{"END_T END_T", Status_END_T, Status_END_T, nil},
					{"END_T START_C", Status_END_T, Status_START_C, ErrInvalidSaga},
					{"END_T END_C", Status_END_T, Status_END_C, ErrInvalidSaga},
					{"END_T ABORT", Status_END_T, Status_ABORT, nil},

					{"START_C NOT_REACHED", Status_START_C, Status_NOT_REACHED, ErrInvalidSaga},
					{"START_C START_T", Status_START_C, Status_START_T, ErrInvalidSaga},
					{"START_C END_T", Status_START_C, Status_END_T, ErrInvalidSaga},
					{"START_C START_C", Status_START_C, Status_START_C, ErrInvalidSaga},
					{"START_C END_C", Status_START_C, Status_END_C, ErrInvalidSaga},
					{"START_C ABORT", Status_START_C, Status_ABORT, nil},

					{"END_C NOT_REACHED", Status_START_C, Status_NOT_REACHED, ErrInvalidSaga},
					{"END_C START_T", Status_START_C, Status_START_T, ErrInvalidSaga},
					{"END_C END_T", Status_START_C, Status_END_T, ErrInvalidSaga},
					{"END_C START_C", Status_START_C, Status_START_C, ErrInvalidSaga},
					{"END_C END_C", Status_START_C, Status_END_C, ErrInvalidSaga},
					{"END_C ABORT", Status_START_C, Status_ABORT, nil},
				}

				for _, tt := range tests {
					t.Run(tt.name, func(t *testing.T) {
						vertices := map[string]Vertex{
							"1": Vertex{Id: "1", Status: tt.status1},
							"2": Vertex{Id: "2", Status: tt.status2},
						}
						saga := NewSaga(vertices, dag)
						assert.Equal(t, CheckValidSaga(saga), tt.expectError)
					})
				}
			})
		})

		t.Run("3 vertices", func(t *testing.T) {
			dag := map[string]map[string][]string{
				"1": {"2": []string{}, "3": []string{}},
				"2": {},
				"3": {},
			}
			tests := []struct {
				name        string
				status1     Status
				status2     Status
				status3     Status
				expectError error
			}{
				{"END_T ABORT Status_NOT_REACHED", Status_END_T, Status_ABORT, Status_NOT_REACHED, nil},
				{"END_T ABORT Status_START_T", Status_END_T, Status_ABORT, Status_START_T, nil},
				{"END_T ABORT Status_END_T", Status_END_T, Status_ABORT, Status_END_T, nil},
				{"END_T ABORT Status_START_C", Status_END_T, Status_ABORT, Status_START_C, nil},
				{"END_T ABORT Status_END_C", Status_END_T, Status_ABORT, Status_END_C, nil},
				{"END_T ABORT Status_ABORT", Status_END_T, Status_ABORT, Status_ABORT, nil},
				{"END_T NOT_REACHED Status_NOT_REACHED", Status_END_T, Status_NOT_REACHED, Status_NOT_REACHED, nil},
				{"END_T NOT_REACHED Status_START_T", Status_END_T, Status_NOT_REACHED, Status_START_T, nil},
				{"END_T NOT_REACHED Status_END_T", Status_END_T, Status_NOT_REACHED, Status_END_T, nil},
				{"END_T NOT_REACHED Status_START_C", Status_END_T, Status_NOT_REACHED, Status_START_C, ErrInvalidSaga},
				{"END_T NOT_REACHED Status_END_C", Status_END_T, Status_NOT_REACHED, Status_END_C, ErrInvalidSaga},
				{"END_T START_T Status_START_T", Status_END_T, Status_START_T, Status_START_T, nil},
				{"END_T START_T Status_END_T", Status_END_T, Status_START_T, Status_END_T, nil},
				{"END_T START_T Status_START_C", Status_END_T, Status_START_T, Status_START_C, ErrInvalidSaga},
				{"END_T START_T Status_END_C", Status_END_T, Status_START_T, Status_END_C, ErrInvalidSaga},
				{"END_T END_T Status_END_T", Status_END_T, Status_END_T, Status_END_T, nil},
				{"END_T END_T Status_START_C", Status_END_T, Status_END_T, Status_START_C, ErrInvalidSaga},
				{"END_T END_T Status_END_C", Status_END_T, Status_END_T, Status_END_C, ErrInvalidSaga},
				{"END_T START_C Status_START_C", Status_END_T, Status_START_C, Status_START_C, ErrInvalidSaga},
				{"END_T START_C Status_END_C", Status_END_T, Status_START_C, Status_END_C, ErrInvalidSaga},
				{"END_T END_C Status_END_C", Status_END_T, Status_END_C, Status_END_C, ErrInvalidSaga},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					vertices := map[string]Vertex{
						"1": Vertex{Id: "1", Status: tt.status1},
						"2": Vertex{Id: "2", Status: tt.status2},
						"3": Vertex{Id: "2", Status: tt.status3},
					}
					saga := NewSaga(vertices, dag)
					assert.Equal(t, CheckValidSaga(saga), tt.expectError)
				})
			}
		})
	})
}
