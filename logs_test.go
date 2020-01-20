package sagas

import "testing"

func TestLog(t *testing.T) {
	config := DefaultConfig()
	s := NewSagas(config)
	defer s.Cleanup()

	index := s.LastIndex()
	if index != 0 {
		t.Fatalf("Expected index: %d, Got: %d\n", 0, index)
	}

	a := Log{
		SagaID: 0,
		Name:   "test",
		Status: Start,
		Data:   []byte{},
	}

	s.Append(a)

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
