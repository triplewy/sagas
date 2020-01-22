package sagas

import (
	"testing"

	"gotest.tools/assert"
)

func TestLog(t *testing.T) {
	config := DefaultConfig()
	s := NewCoordinator(config)
	defer s.Cleanup()

	index := s.LastIndex()
	assert.Equal(t, index, 0)

	a := Log{
		0,
		0,
		Saga{},
	}

	s.AppendLog(a)

	index = s.LastIndex()
	if index != 1 {
		t.Fatalf("Expected index: %d, Got: %d\n", 2, index)
	}

	b, err := s.GetLog(index)
	if err != nil {
		t.Fatal(err)
	}

	if !a.Equal(b) {
		t.Fatalf("Expected: %v, Got: %v\n", a, b)
	}

	_, err = s.GetLog(index + 1)
	if err != ErrLogIndexNotFound {
		t.Fatalf("Expected: %v, Got: %v\n", ErrLogIndexNotFound, err)
	}
}
