package sagas

import (
	"testing"
)

func TestSaga(t *testing.T) {
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

	t.Run("equivalent dags", func(t *testing.T) {

	})

	t.Run("finished or abort", func(t *testing.T) {

	})

	t.Run("valid saga", func(t *testing.T) {

	})
}
