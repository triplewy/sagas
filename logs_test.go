package sagas

import (
	"testing"

	"github.com/triplewy/sagas/utils"

	"gotest.tools/assert"
)

func TestLog(t *testing.T) {
	config := DefaultConfig()
	s := NewCoordinator(config)
	defer s.Cleanup()

	index := s.LastIndex()
	assert.Equal(t, index, uint64(0))

	sagaID := s.NewSagaID()
	assert.Equal(t, sagaID, uint64(1))

	dag := map[VertexID]map[VertexID]struct{}{1: {}}
	vertices := map[VertexID]SagaVertex{1: SagaVertex{
		VertexID: 1,
		TFunc:    SagaFunc{},
		CFunc:    SagaFunc{},
		Status:   StartT,
	}}
	saga := NewSaga(dag, vertices)
	buf, err := utils.EncodeMsgPack(saga)
	if err != nil {
		t.Fatal(err)
	}

	s.AppendLog(sagaID, Graph, buf.Bytes())

	index = s.LastIndex()
	assert.Equal(t, index, uint64(1))

	log, err := s.GetLog(index)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, log.SagaID, uint64(1))
	assert.Equal(t, log.LogType, Graph)

	var readSaga Saga
	assert.NilError(t, utils.DecodeMsgPack(log.Data, &readSaga))

	assert.DeepEqual(t, saga, readSaga)

	_, err = s.GetLog(index + 1)
	assert.Equal(t, err, ErrLogIndexNotFound)

	sagaID = s.NewSagaID()
	assert.Equal(t, sagaID, uint64(2))
}
