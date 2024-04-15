package core

import (
	"testing"

	"github.com/EggsyOnCode/xenolith/crypto_lib"
	"github.com/stretchr/testify/assert"
)

func TestTokenTransferNoBalance(t *testing.T) {
	aS := NewAccountState()
	from := (crypto_lib.GeneratePrivateKey().PublicKey())
	fromPtr := crypto_lib.PublicKey(from)
	fromAddr := fromPtr.Address()
	to := crypto_lib.GeneratePrivateKey().PublicKey()
	toPtr := crypto_lib.PublicKey(to)
	toAddr := toPtr.Address()
	value := uint64(666)

	assert.NotNil(t, aS.TransferFunds(fromAddr, toAddr, value))
}

func TestTokenTransferWithBalance(t *testing.T) {
	aS := NewAccountState()
	from := (crypto_lib.GeneratePrivateKey().PublicKey())
	fromPtr := crypto_lib.PublicKey(from)
	fromAddr := fromPtr.Address()
	to := crypto_lib.GeneratePrivateKey().PublicKey()
	toPtr := crypto_lib.PublicKey(to)
	toAddr := toPtr.Address()
	value := uint64(666)
	transferValue := uint64(400)

	assert.Nil(t, aS.AddBalance(fromAddr, value))

	assert.Nil(t, aS.TransferFunds(fromAddr, toAddr, transferValue))

	balance, err := aS.GetBalance(fromAddr)
	assert.Nil(t, err)
	assert.Equal(t, value-transferValue, balance)
}
