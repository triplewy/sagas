package sagas

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"gotest.tools/assert"
)

func TestSagaFunc(t *testing.T) {
	tFunc := SagaFunc{
		ID:     "book",
		Input:  map[string]interface{}{"userID": "user0"},
		Output: make(map[string]interface{}, 0),
	}

	assert.Equal(t, tFunc.Input["userID"], "user0")

	tFuncCopy := SagaFunc{
		ID:     "book",
		Input:  map[string]interface{}{"userID": "user0"},
		Output: make(map[string]interface{}, 0),
	}

	assert.DeepEqual(t, tFunc, tFuncCopy)

	cFunc := SagaFunc{
		ID:     "cancel",
		Input:  map[string]interface{}{"reservationID": "reservation0"},
		Output: make(map[string]interface{}, 0),
	}

	assert.Assert(t, cmp.Equal(tFunc, cFunc) == false)
}

func TestSagaVertex(t *testing.T) {
	tFunc := SagaFunc{
		ID:     "book",
		Input:  map[string]interface{}{"userID": "user0"},
		Output: make(map[string]interface{}, 0),
	}

	cFunc := SagaFunc{
		ID:     "cancel",
		Input:  map[string]interface{}{"reservationID": "reservation0"},
		Output: make(map[string]interface{}, 0),
	}

	v := SagaVertex{
		VertexID: 0,
		TFunc:    tFunc,
		CFunc:    cFunc,
		Status:   StartT,
	}

	vCopy := SagaVertex{
		VertexID: 0,
		TFunc:    tFunc,
		CFunc:    cFunc,
		Status:   StartT,
	}

	assert.DeepEqual(t, v, vCopy)
}

func TestSaga(t *testing.T) {
	saga := Saga{}
}
