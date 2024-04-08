package core

import "fmt"

type State struct {
	data map[string][]byte
}

func NewState() *State {
	return &State{
		data: make(map[string][]byte),
	}
}

func (s *State) Put(k, value []byte) error {
	s.data[string(k)] = value
	return nil
}

func (s *State) Delete(k []byte) error {
	delete(s.data, string(k))
	return nil
}

func (s *State) Get(k []byte) ([]byte, error) {
	key := string(k)
	value, ok := s.data[key]
	if !ok {
		return nil, fmt.Errorf("key not found %v", key)
	}

	return value, nil

}
