package core

import (
	"testing"

	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func TestAccountState(t *testing.T) {
	a := NewAccountState()

	addr := crypto_lib.GeneratePrivateKey().PublicKey().Address()
	account := a.CreateAccount(addr)

	assert.Equal(t, account.Address, addr)
	assert.Equal(t, account.Balance, uint64(0))

	fetchAccount, err := a.GetAccount(addr)
	assert.Nil(t, err)
	assert.Equal(t, fetchAccount, account)
}

func TestAccountTransferFailing(t *testing.T) {
	a := NewAccountState()

	addrAlice := crypto_lib.GeneratePrivateKey().PublicKey().Address()
	addrBob := crypto_lib.GeneratePrivateKey().PublicKey().Address()

	accAlice := a.CreateAccount(addrAlice)
	accBob := a.CreateAccount(addrBob)

	assert.Equal(t, accAlice.Balance, uint64(0))
	assert.Equal(t, accBob.Balance, uint64(0))

	err := a.Transfer(addrAlice, addrBob, 100)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "insufficient funds")
}
func TestAccountTransferSuccess(t *testing.T) {
	a := NewAccountState()

	addrAlice := crypto_lib.GeneratePrivateKey().PublicKey().Address()
	addrBob := crypto_lib.GeneratePrivateKey().PublicKey().Address()

	accAlice := a.CreateAccount(addrAlice)
	accBob := a.CreateAccount(addrBob)
	accAlice.Balance = uint64(150)

	assert.Equal(t, accAlice.Balance, uint64(150))
	assert.Equal(t, accBob.Balance, uint64(0))

	err := a.Transfer(addrAlice, addrBob, 100)
	assert.Nil(t, err)
}
