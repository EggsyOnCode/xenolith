package core

import (
	"fmt"
	"sync"

	"github.com/EggsyOnCode/xenolith/core_types"
)

type AccountState struct {
	mu    sync.RWMutex
	state map[core_types.Address]uint64
}

func NewAccountState() *AccountState {
	return &AccountState{
		state: make(map[core_types.Address]uint64),
	}
}

func checkAccountExistence(s *AccountState, addr core_types.Address) error {
	if _, ok := s.state[addr]; !ok {
		return fmt.Errorf("account not found %v", addr)
	}
	return nil
}

func (s *AccountState) GetBalance(addr core_types.Address) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := checkAccountExistence(s, addr); err != nil {
		return 0, err
	}

	return s.state[addr], nil
}

func (s *AccountState) AddBalance(addr core_types.Address, b uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	balance, ok := s.state[addr]
	if !ok {
		s.state[addr] = b
		return nil
	}
	s.state[addr] = balance + b
	return nil
}

func (s *AccountState) SubBalance(addr core_types.Address, amt uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := checkAccountExistence(s, addr); err != nil {
		return err
	}
	balance := s.state[addr]
	if balance < amt {
		return fmt.Errorf("insufficient balance %v", balance)
	}
	s.state[addr] -= amt
	return nil
}

func (s *AccountState) TransferFunds(from, to core_types.Address, amt uint64) error {
	if err := s.SubBalance(from, amt); err != nil {
		return err
	}

	return s.AddBalance(to, amt)
}

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
