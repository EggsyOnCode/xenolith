package core

import (
	"fmt"
	"sync"

	"github.com/EggsyOnCode/xenolith/core_types"
)

var (
	ErrAccountNotFound   = fmt.Errorf("account not found")
	ErrInsufficientFunds = fmt.Errorf("insufficient funds")
)

type Account struct {
	Address core_types.Address
	//TODO: Make the Balance a BigInt to store fractions values
	Balance uint64
}

func (a *Account) String() string {
	return fmt.Sprintf("%d", a.Balance)
}

type AccountState struct {
	mu sync.RWMutex
	//be careful with ptrs
	//TODO: use of atomic values when changing Account Data
	accounts map[core_types.Address]*Account
}

func NewAccountState() *AccountState {
	return &AccountState{
		accounts: make(map[core_types.Address]*Account),
	}
}

func (a *AccountState) CreateAccount(addr core_types.Address) *Account {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.accounts[addr] != nil {
		return a.accounts[addr]
	}

	account := &Account{
		Address: addr,
		Balance: 0,
	}

	a.accounts[addr] = account

	return account

}

func (a *AccountState) GetAccount(addr core_types.Address) (*Account, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.getAccountWithoutLock(addr)
}

func (a *AccountState) getAccountWithoutLock(addr core_types.Address) (*Account, error) {
	account, ok := a.accounts[addr]
	if !ok {
		return nil, ErrAccountNotFound
	}

	return account, nil
}

func (a *AccountState) GetBalance(addr core_types.Address) (uint64, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	account, err := a.getAccountWithoutLock(addr)
	if err != nil {
		return 0, err
	}

	return account.Balance, nil
}

func (a *AccountState) Transfer(from, to core_types.Address, amt uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	fromAccount, err := a.getAccountWithoutLock(from)
	if err != nil {
		return err
	}

	if fromAccount.Address.String() != "996fb92427ae41e4649b934ca495991b7852b855" {
		if fromAccount.Balance < amt {
			return ErrInsufficientFunds
		}
	}

	//usage of atomic vals here perhaps!! TODO
	//only transfer if the account has a balance
	if fromAccount.Balance != 0 {
		fromAccount.Balance -= amt
	}

	if a.accounts[to] == nil {
		a.accounts[to] = &Account{
			Address: to,
			Balance: amt,
		}
	}

	a.accounts[to].Balance += amt

	return nil
}
